package controller

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/internal/minecraft"
	"github.com/kmdkuk/mcing/pkg/config"
	"github.com/kmdkuk/mcing/pkg/constants"
)

const (
	defaultTerminationGracePeriodSeconds = 30
	defaultConfigMode                    = 0o644
	livenessInitialDelaySeconds          = 120
	livenessPeriodSeconds                = 60
	readinessInitialDelaySeconds         = 20
	readinessPeriodSeconds               = 10
	readinessFailureThreshold            = 12
	rconPasswordLength                   = 24
	// Autopause probe constants.
	autopauseLivenessInitialDelay      = 30
	autopauseLivenessPeriodSeconds     = 60
	autopauseReadinessInitialDelay     = 10
	autopauseReadinessPeriodSeconds    = 10
	autopauseReadinessFailureThreshold = 12
)

// MinecraftReconciler reconciles a Minecraft object.
type MinecraftReconciler struct {
	client.Client

	log              logr.Logger
	scheme           *runtime.Scheme
	initImageName    string
	agentImageName   string
	minecraftManager minecraft.MinecraftManager
	gatewayConfig    GatewayConfig
}

// NewMinecraftReconciler returns a new MinecraftReconciler.
func NewMinecraftReconciler(
	client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	initImageName, agentImageName string,
	minecraftManager minecraft.MinecraftManager,
	gatewayConfig GatewayConfig,
) *MinecraftReconciler {
	l := log.WithName("Minecraft")
	return &MinecraftReconciler{
		Client:           client,
		log:              l,
		scheme:           scheme,
		initImageName:    initImageName,
		agentImageName:   agentImageName,
		minecraftManager: minecraftManager,
		gatewayConfig:    gatewayConfig,
	}
}

//+kubebuilder:rbac:groups=mcing.kmdkuk.com,resources=minecrafts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mcing.kmdkuk.com,resources=minecrafts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mcing.kmdkuk.com,resources=minecrafts/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

//go:embed lazymc.toml.tmpl
var lazymcTomlTmpl string

// LazymcConfig holds configuration for lazymc template rendering.
type LazymcConfig struct {
	PublicPort  int32
	ServerPort  int32
	Command     string
	SleepAfter  int32
	RconEnabled bool
	RconPort    int32
}

// Reconcile implements Reconciler interface.
// See https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *MinecraftReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("minecraft", req.NamespacedName)
	log.Info("start reconciliation loop")

	mc := &mcingv1alpha1.Minecraft{}
	if err := r.Get(ctx, req.NamespacedName, mc); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Minecraft is not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to get Minecraft")
		return ctrl.Result{}, err
	}

	if mc.DeletionTimestamp != nil {
		if !controllerutil.ContainsFinalizer(mc, constants.Finalizer) {
			return ctrl.Result{}, nil
		}

		log.Info("start finalizing Minecraft")

		controllerutil.RemoveFinalizer(mc, constants.Finalizer)
		if err := r.Update(ctx, mc); err != nil {
			log.Error(err, "failed to remove finalizer")
			return ctrl.Result{}, err
		}

		log.Info("finalizing Minecraft is completed")
		return ctrl.Result{}, nil
	}

	props, err := r.reconcileConfigMap(ctx, mc)
	if err != nil {
		log.Error(err, "failed to reconcile configmap")
		return ctrl.Result{}, err
	}

	if err := r.reconcileRconSecret(ctx, mc); err != nil {
		log.Error(err, "failed to reconcile rcon secret")
		return ctrl.Result{}, err
	}

	if err := r.reconcileAllService(ctx, mc); err != nil {
		log.Error(err, "failed to reconcile service")
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatefulSet(ctx, mc, props); err != nil {
		log.Error(err, "failed to reconcile statefulset")
		return ctrl.Result{}, err
	}

	if err := r.minecraftManager.Update(client.ObjectKeyFromObject(mc)); err != nil {
		log.Error(err, "failed to update MinecraftManager")
		return ctrl.Result{}, err
	}
	log.Info("finish reconciliation")
	return ctrl.Result{}, nil
}

//nolint:gocognit,funlen // debug logic increases complexity
func (r *MinecraftReconciler) reconcileStatefulSet(
	ctx context.Context,
	mc *mcingv1alpha1.Minecraft,
	props *corev1.ConfigMap,
) error {
	logger := r.log.WithName("statefulset")

	sts := &appsv1.StatefulSet{}
	sts.Namespace = mc.Namespace
	sts.Name = mc.PrefixedName()

	var orig, updated *appsv1.StatefulSetSpec

	result, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if logger.V(1).Enabled() {
			orig = sts.Spec.DeepCopy()
		}
		labels := labelSet(mc, constants.AppComponentServer)
		sts.Labels = config.MergeMap(sts.Labels, labels)

		sts.Spec.Replicas = ptr.To[int32](1)
		sts.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}
		sts.Spec.ServiceName = mc.HeadlessServiceName()
		sts.Spec.VolumeClaimTemplates = make([]corev1.PersistentVolumeClaim, len(mc.Spec.VolumeClaimTemplates))
		for i, v := range mc.Spec.VolumeClaimTemplates {
			pvc := v.ToCoreV1()
			pvc.Namespace = mc.Namespace
			if err := ctrl.SetControllerReference(mc, &pvc, r.scheme); err != nil {
				panic(err)
			}
			pvc.Namespace = ""
			sts.Spec.VolumeClaimTemplates[i] = pvc
		}

		sts.Spec.Template.Annotations = config.MergeMap(sts.Spec.Template.Annotations, mc.Spec.PodTemplate.Annotations)
		sts.Spec.Template.Labels = config.MergeMap(sts.Spec.Template.Labels, mc.Spec.PodTemplate.Labels)
		sts.Spec.Template.Labels = config.MergeMap(sts.Spec.Template.Labels, labels)

		podSpec := mc.Spec.PodTemplate.Spec.DeepCopy()
		podSpec.DeprecatedServiceAccount = sts.Spec.Template.Spec.DeprecatedServiceAccount
		if len(podSpec.RestartPolicy) == 0 {
			podSpec.RestartPolicy = sts.Spec.Template.Spec.RestartPolicy
		}
		if podSpec.TerminationGracePeriodSeconds == nil {
			podSpec.TerminationGracePeriodSeconds = ptr.To[int64](defaultTerminationGracePeriodSeconds)
		}
		if len(podSpec.DNSPolicy) == 0 {
			podSpec.DNSPolicy = sts.Spec.Template.Spec.DNSPolicy
		}
		if podSpec.SecurityContext == nil {
			podSpec.SecurityContext = sts.Spec.Template.Spec.SecurityContext
		}
		if len(podSpec.SchedulerName) == 0 {
			podSpec.SchedulerName = sts.Spec.Template.Spec.SchedulerName
		}

		podSpec.Volumes = append(podSpec.Volumes,
			corev1.Volume{
				Name: constants.ConfigVolumeName, VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: props.Name,
						},
						DefaultMode: ptr.To[int32](defaultConfigMode),
					},
				},
			},
		)

		if *mc.Spec.AutoPause.Enabled {
			podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
				Name: constants.LazymcVolumeName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
		}

		containers := make([]corev1.Container, 0)
		minecraftContainer, err := makeMinecraftContainer(
			mc,
			podSpec.Containers,
			sts.Spec.Template.Spec.Containers,
			constants.ServerPort,
		)
		if err != nil {
			return err
		}
		containers = append(containers, minecraftContainer)
		containers = append(containers, r.makeAgentContainer(mc))
		podSpec.Containers = containers
		podSpec.InitContainers = r.makeInitContainer(mc)

		podSpec.DeepCopyInto(&sts.Spec.Template.Spec)

		if logger.V(1).Enabled() {
			updated = sts.Spec.DeepCopy()
		}
		return ctrl.SetControllerReference(mc, sts, r.scheme)
	})
	if err != nil {
		logger.Error(err, "failed to reconcile stateful set")
		return err
	}
	if result != controllerutil.OperationResultNone {
		logger.Info("reconciled stateful set", "operation", string(result))
		if logger.V(1).Enabled() {
			logger.V(1).Info("diff", "diff", cmp.Diff(orig, updated))
		}
	}
	return nil
}

//nolint:funlen // container setup requires many fields
func makeMinecraftContainer(
	mc *mcingv1alpha1.Minecraft,
	desired, _ []corev1.Container,
	publicPort int32,
) (corev1.Container, error) {
	var source *corev1.Container
	for i := range desired {
		c := &desired[i]
		if c.Name == constants.MinecraftContainerName {
			source = c
			break
		}
	}
	if source == nil {
		return corev1.Container{}, errors.New("minecraft container not found")
	}

	c := source.DeepCopy()
	c.Stdin = true
	c.TTY = true
	c.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"mc-health",
				},
			},
		},
		InitialDelaySeconds: livenessInitialDelaySeconds,
		PeriodSeconds:       livenessPeriodSeconds,
	}
	c.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"mc-health",
				},
			},
		},
		InitialDelaySeconds: readinessInitialDelaySeconds,
		PeriodSeconds:       readinessPeriodSeconds,
		FailureThreshold:    readinessFailureThreshold,
	}
	rconSecretName := mc.RconSecretName()
	c.Env = append(c.Env, corev1.EnvVar{
		Name: constants.RconPasswordEnvName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: rconSecretName,
				},
				Key: constants.RconPasswordSecretKey,
			},
		},
	})
	c.Ports = append(
		c.Ports,
		corev1.ContainerPort{
			ContainerPort: constants.ServerPort,
			Name:          constants.ServerPortName,
			Protocol:      corev1.ProtocolTCP,
		},
		corev1.ContainerPort{
			ContainerPort: constants.RconPort,
			Name:          constants.RconPortName,
			Protocol:      corev1.ProtocolTCP,
		},
	)
	c.VolumeMounts = append(c.VolumeMounts,
		corev1.VolumeMount{
			MountPath: constants.DataPath,
			Name:      constants.DataVolumeName,
		},
		corev1.VolumeMount{
			MountPath: constants.ConfigPath,
			Name:      constants.ConfigVolumeName,
			ReadOnly:  true,
		},
	)
	c.Lifecycle = &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				// '|| true' is to prevent the container from being killed if rcon-cli fails
				Command: []string{"/bin/sh", "-c", "rcon-cli stop || true"},
			},
		},
	}

	if *mc.Spec.AutoPause.Enabled {
		// Override probes to check the public port (lazymc)
		// Use tcpSocket because lazymc accepts connections even when backend is sleeping
		c.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt32(publicPort),
				},
			},
			InitialDelaySeconds: autopauseLivenessInitialDelay,
			PeriodSeconds:       autopauseLivenessPeriodSeconds,
		}
		c.ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt32(publicPort),
				},
			},
			InitialDelaySeconds: autopauseReadinessInitialDelay,
			PeriodSeconds:       autopauseReadinessPeriodSeconds,
			FailureThreshold:    autopauseReadinessFailureThreshold,
		}

		// Wrap command with lazymc via tini to handle zombies and signals
		c.Command = []string{filepath.Join(constants.LazymcPath, constants.LazymcBinName)}
		c.Args = []string{"--config", filepath.Join(constants.LazymcPath, constants.LazymcConfigName)}
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
			Name:      constants.LazymcVolumeName,
			MountPath: constants.LazymcPath,
		})
	}
	return *c, nil
}

func (r *MinecraftReconciler) makeAgentContainer(mc *mcingv1alpha1.Minecraft) corev1.Container {
	c := corev1.Container{}
	c.Name = constants.AgentContainerName
	c.Image = r.agentImageName
	c.Ports = []corev1.ContainerPort{
		{
			ContainerPort: constants.AgentPort,
			Name:          constants.AgentPortName,
			Protocol:      corev1.ProtocolTCP,
		},
	}
	c.VolumeMounts = append(c.VolumeMounts,
		corev1.VolumeMount{
			MountPath: constants.DataPath,
			Name:      constants.DataVolumeName,
		},
		corev1.VolumeMount{
			MountPath: constants.ConfigPath,
			Name:      constants.ConfigVolumeName,
			ReadOnly:  true,
		},
	)

	rconSecretName := mc.RconSecretName()
	if mc.Spec.RconPasswordSecretName != nil {
		rconSecretName = *mc.Spec.RconPasswordSecretName
	}

	c.Env = append(c.Env, corev1.EnvVar{
		Name: constants.RconPasswordEnvName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: rconSecretName,
				},
				Key: constants.RconPasswordSecretKey,
			},
		},
	})

	return c
}

func (r *MinecraftReconciler) makeInitContainer(mc *mcingv1alpha1.Minecraft) []corev1.Container {
	image := r.initImageName
	c := corev1.Container{
		Name:  constants.InitContainerName,
		Image: image,
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: constants.ConfigPath,
				Name:      constants.ConfigVolumeName,
			},
			{
				MountPath: constants.DataPath,
				Name:      constants.DataVolumeName,
			},
		},
	}

	if *mc.Spec.AutoPause.Enabled {
		c.Args = append(c.Args, "--enable-lazymc")
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
			Name:      constants.LazymcVolumeName,
			MountPath: constants.LazymcPath,
		})
		c.Env = append(c.Env, corev1.EnvVar{
			Name: constants.RconPasswordEnvName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: mc.RconSecretName(),
					},
					Key: constants.RconPasswordSecretKey,
				},
			},
		})
	}

	var initContainers []corev1.Container
	initContainers = append(initContainers, c)

	for _, given := range mc.Spec.PodTemplate.Spec.InitContainers {
		ic := given.DeepCopy()
		initContainers = append(initContainers, *ic)
	}
	return initContainers
}

func (r *MinecraftReconciler) reconcileAllService(ctx context.Context, mc *mcingv1alpha1.Minecraft) error {
	err := r.reconcileService(ctx, mc, true)
	if err != nil {
		return err
	}
	err = r.reconcileService(ctx, mc, false)
	if err != nil {
		return err
	}
	return nil
}

//nolint:gocognit,funlen // debug logic increases complexity and length
func (r *MinecraftReconciler) reconcileService(ctx context.Context, mc *mcingv1alpha1.Minecraft, headless bool) error {
	logger := r.log.WithName("service")

	svc := &corev1.Service{}
	svc.Namespace = mc.Namespace
	svc.Name = mc.PrefixedName()
	if headless {
		svc.Name = mc.HeadlessServiceName()
	}
	var orig, updated *corev1.ServiceSpec
	result, err := ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if logger.V(1).Enabled() {
			orig = svc.Spec.DeepCopy()
		}
		labels := labelSet(mc, constants.AppComponentServer)
		sSpec := &corev1.ServiceSpec{}
		tmpl := mc.Spec.ServiceTemplate

		// Handle service configuration based on type and mc-router settings
		switch {
		case !headless && r.gatewayConfig.Enabled:
			// When mc-router is enabled and this is NOT the headless service,
			// force ClusterIP type and add the mc-router annotation
			externalServerName := mc.GetExternalServerName(r.gatewayConfig.DefaultDomain)
			if svc.Annotations == nil {
				svc.Annotations = make(map[string]string)
			}
			svc.Annotations[constants.MCRouterAnnotation] = externalServerName

			// Still allow other annotations and labels from template
			if tmpl != nil {
				svc.Annotations = config.MergeMap(svc.Annotations, tmpl.Annotations)
				// Ensure mc-router annotation is not overwritten
				svc.Annotations[constants.MCRouterAnnotation] = externalServerName
				svc.Labels = config.MergeMap(svc.Labels, tmpl.Labels)
			}
			svc.Labels = config.MergeMap(svc.Labels, labels)

			// Force ClusterIP type when mc-router is enabled
			sSpec.Type = corev1.ServiceTypeClusterIP
		case !headless && tmpl != nil:
			svc.Annotations = config.MergeMap(svc.Annotations, tmpl.Annotations)
			svc.Labels = config.MergeMap(svc.Labels, tmpl.Labels)
			svc.Labels = config.MergeMap(svc.Labels, labels)

			if tmpl.Spec != nil {
				tmpl.Spec.DeepCopyInto(sSpec)
			}
		default:
			svc.Labels = config.MergeMap(svc.Labels, labels)
		}

		if headless {
			sSpec.ClusterIP = corev1.ClusterIPNone
			sSpec.ClusterIPs = svc.Spec.ClusterIPs
			sSpec.Type = corev1.ServiceTypeClusterIP
			sSpec.PublishNotReadyAddresses = true
		} else {
			sSpec.ClusterIP = svc.Spec.ClusterIP
			sSpec.ClusterIPs = svc.Spec.ClusterIPs
			// Only set type from existing service if not already set (e.g., by mc-router or template)
			if len(sSpec.Type) == 0 {
				sSpec.Type = svc.Spec.Type
			}
		}
		if len(sSpec.SessionAffinity) == 0 {
			sSpec.SessionAffinity = svc.Spec.SessionAffinity
		}
		if len(sSpec.ExternalTrafficPolicy) == 0 {
			sSpec.ExternalTrafficPolicy = svc.Spec.ExternalTrafficPolicy
		}
		if sSpec.HealthCheckNodePort == 0 {
			sSpec.HealthCheckNodePort = svc.Spec.HealthCheckNodePort
		}
		if sSpec.IPFamilies == nil {
			sSpec.IPFamilies = svc.Spec.IPFamilies
		}
		if sSpec.IPFamilyPolicy == nil {
			sSpec.IPFamilyPolicy = svc.Spec.IPFamilyPolicy
		}
		sSpec.Selector = labels

		var serverNodePort, rconNodePort int32
		for _, p := range svc.Spec.Ports {
			switch p.Name {
			case constants.ServerPortName:
				serverNodePort = p.NodePort
			case constants.RconPortName:
				rconNodePort = p.NodePort
			}
		}
		sSpec.Ports = append(sSpec.Ports, corev1.ServicePort{
			Name:       constants.ServerPortName,
			Protocol:   corev1.ProtocolTCP,
			Port:       constants.ServerPort,
			TargetPort: intstr.FromString(constants.ServerPortName),
			NodePort:   serverNodePort,
		})

		if headless || sSpec.Type != corev1.ServiceTypeLoadBalancer {
			sSpec.Ports = append(sSpec.Ports, corev1.ServicePort{
				Name:       constants.RconPortName,
				Protocol:   corev1.ProtocolTCP,
				Port:       constants.RconPort,
				TargetPort: intstr.FromString(constants.RconPortName),
				NodePort:   rconNodePort,
			})
		}

		sSpec.DeepCopyInto(&svc.Spec)

		if logger.V(1).Enabled() {
			updated = svc.Spec.DeepCopy()
		}
		return ctrl.SetControllerReference(mc, svc, r.scheme)
	})
	if err != nil {
		return fmt.Errorf("failed to reconcile service: %w", err)
	}
	if result != controllerutil.OperationResultNone {
		logger.Info("reconciled service", "operation", string(result))
		if logger.V(1).Enabled() {
			logger.V(1).Info("diff", "diff", cmp.Diff(orig, updated))
		}
	}

	return nil
}

// buildServerCommand constructs the server command from container spec.
func buildServerCommand(mc *mcingv1alpha1.Minecraft) string {
	cmd := "/start"
	if len(mc.Spec.PodTemplate.Spec.Containers) == 0 {
		return cmd
	}
	for _, container := range mc.Spec.PodTemplate.Spec.Containers {
		if container.Name != constants.MinecraftContainerName || len(container.Command) == 0 {
			continue
		}
		parts := make([]string, 0, len(container.Command)+len(container.Args))
		parts = append(parts, container.Command...)
		parts = append(parts, container.Args...)
		return strings.Join(parts, " ")
	}
	return cmd
}

//nolint:gocognit,funlen // config map reconciliation has many conditional paths
func (r *MinecraftReconciler) reconcileConfigMap(
	ctx context.Context,
	mc *mcingv1alpha1.Minecraft,
) (*corev1.ConfigMap, error) {
	logger := r.log.WithName("configmap")

	var userProps map[string]string
	if mc.Spec.ServerPropertiesConfigMapName != nil {
		cm := &corev1.ConfigMap{}
		err := r.Get(
			ctx,
			types.NamespacedName{Namespace: mc.Namespace, Name: *mc.Spec.ServerPropertiesConfigMapName},
			cm,
		)
		if err != nil {
			logger.Error(err, "failed to get specified configmap", "configmap", *mc.Spec.ServerPropertiesConfigMapName)
			return nil, err
		}
		userProps = cm.Data
	}

	props, err := config.GenServerProps(userProps)
	if err != nil {
		return nil, err
	}

	if *mc.Spec.AutoPause.Enabled {
		// Force internal port by replacing the enforced standard port
		target := fmt.Sprintf("server-port=%d", constants.ServerPort)
		replacement := fmt.Sprintf("server-port=%d", constants.InternalServerPort)
		props = strings.Replace(props, target, replacement, 1)
	}

	var otherProps map[string]string
	if mc.Spec.OtherConfigMapName != nil {
		cm := &corev1.ConfigMap{}
		err = r.Get(ctx, types.NamespacedName{Namespace: mc.Namespace, Name: *mc.Spec.OtherConfigMapName}, cm)
		if err != nil {
			logger.Error(err, "failed to get configmap", "configmap", *mc.Spec.OtherConfigMapName)
		}
		otherProps = cm.Data
	}

	cm := &corev1.ConfigMap{}
	cm.Namespace = mc.Namespace
	cm.Name = mc.PrefixedName()
	result, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		cm.Labels = config.MergeMap(cm.Labels, labelSet(mc, constants.AppComponentServer))
		cm.Data = map[string]string{
			constants.ServerPropsName: props,
		}
		if v, ok := otherProps[constants.BanIPName]; ok {
			cm.Data[constants.BanIPName] = v
		}
		if v, ok := otherProps[constants.BanPlayerName]; ok {
			cm.Data[constants.BanPlayerName] = v
		}
		if v, ok := otherProps[constants.OpsName]; ok {
			cm.Data[constants.OpsName] = v
		}
		if v, ok := otherProps[constants.WhiteListName]; ok {
			cm.Data[constants.WhiteListName] = v
		}

		// Generate lazymc.toml from templates
		//nolint:nestif // autopause configuration adds necessary nesting
		if *mc.Spec.AutoPause.Enabled {
			// Determine backend command
			cmd := buildServerCommand(mc)

			rconEnabled := true
			rconPort := constants.RconPort

			if v, ok := userProps["enable-rcon"]; ok {
				if v == "false" {
					rconEnabled = false
				}
			}
			if v, ok := userProps["rcon.port"]; ok {
				if portVal, parseErr := strconv.Atoi(v); parseErr == nil {
					rconPort = int32(portVal) //nolint:gosec // port values are within int32 range
				}
			}

			// The rcon password is injected by mcing-init from secret via env.
			lazymcConfig := LazymcConfig{
				PublicPort:  constants.ServerPort,
				ServerPort:  constants.InternalServerPort,
				Command:     cmd,
				SleepAfter:  int32(mc.Spec.AutoPause.TimeoutSeconds), //nolint:gosec // timeout is within int32 range
				RconEnabled: rconEnabled,
				RconPort:    rconPort,
			}

			lazymcToml, lazymcTomlErr := config.ExecuteTemplate(lazymcTomlTmpl, lazymcConfig)
			if lazymcTomlErr != nil {
				return lazymcTomlErr
			}
			cm.Data["lazymc.toml"] = lazymcToml
		}

		return ctrl.SetControllerReference(mc, cm, r.scheme)
	})
	if err != nil {
		return nil, err
	}

	if result != controllerutil.OperationResultNone {
		logger.Info("reconciled server.properties configmap", "operation", string(result))
	}

	return cm, nil
}

func (r *MinecraftReconciler) reconcileRconSecret(ctx context.Context, mc *mcingv1alpha1.Minecraft) error {
	logger := r.log.WithName("rcon-secret")
	if mc.Spec.RconPasswordSecretName != nil {
		return nil
	}
	// default
	secretName := mc.RconSecretName()
	secret := &corev1.Secret{}
	secret.Namespace = mc.Namespace
	secret.Name = secretName
	result, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		if _, ok := secret.Data[constants.RconPasswordSecretKey]; !ok {
			secret.Data[constants.RconPasswordSecretKey] = []byte(rand.String(rconPasswordLength))
		}
		return ctrl.SetControllerReference(mc, secret, r.scheme)
	})
	if err != nil {
		return err
	}
	if result != controllerutil.OperationResultNone {
		logger.Info("reconciled rcon secret", "operation", string(result))
	}
	return nil
}

func labelSet(mc *mcingv1alpha1.Minecraft, component string) map[string]string {
	return map[string]string{
		constants.LabelAppInstance:  mc.Name,
		constants.LabelAppName:      constants.AppName,
		constants.LabelAppComponent: component,
		constants.LabelAppCreatedBy: constants.ControllerName,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MinecraftReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.Add(r.minecraftManager); err != nil {
		return err
	}
	configMapHandler := handler.EnqueueRequestsFromMapFunc(
		func(ctx context.Context, a client.Object) []reconcile.Request {
			mcs := &mcingv1alpha1.MinecraftList{}
			if err := r.List(ctx, mcs, client.InNamespace(a.GetNamespace())); err != nil {
				return nil
			}
			var reqs []reconcile.Request
			for _, mc := range mcs.Items {
				if mc.Spec.ServerPropertiesConfigMapName == nil {
					continue
				}
				if *mc.Spec.ServerPropertiesConfigMapName == a.GetName() {
					reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&mc)})
				}
			}
			return reqs
		},
	)
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcingv1alpha1.Minecraft{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Watches(&corev1.ConfigMap{}, configMapHandler).
		Complete(r)
}
