package controller

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/constants"
)

// Gateway controller probe constants.
const (
	gatewayReadinessInitialDelaySeconds = 5
	gatewayReadinessPeriodSeconds       = 10
	gatewayLivenessInitialDelaySeconds  = 15
	gatewayLivenessPeriodSeconds        = 20
)

// GatewayConfig holds mc-router gateway configuration.
type GatewayConfig struct {
	Enabled        bool
	DefaultDomain  string
	Namespace      string
	ServiceAccount string
	ServiceType    corev1.ServiceType
	Image          string
}

// GatewayReconciler reconciles the mc-router gateway.
type GatewayReconciler struct {
	client.Client

	log    logr.Logger
	scheme *runtime.Scheme
	config GatewayConfig
}

// NewGatewayReconciler returns a new GatewayReconciler.
func NewGatewayReconciler(
	client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	config GatewayConfig,
) *GatewayReconciler {
	return &GatewayReconciler{
		Client: client,
		log:    log.WithName("Gateway"),
		scheme: scheme,
		config: config,
	}
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch

// Reconcile reconciles the mc-router gateway resources.
func (r *GatewayReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	log := r.log

	// Only reconcile if mc-router is enabled
	if !r.config.Enabled {
		return ctrl.Result{}, nil
	}

	// Ensure gateway namespace exists
	if err := r.ensureNamespace(ctx); err != nil {
		log.Error(err, "failed to ensure gateway namespace")
		return ctrl.Result{}, err
	}

	// Ensure service account exists
	if err := r.reconcileServiceAccount(ctx); err != nil {
		log.Error(err, "failed to reconcile service account")
		return ctrl.Result{}, err
	}

	// Reconcile mc-router Deployment
	if err := r.reconcileDeployment(ctx); err != nil {
		log.Error(err, "failed to reconcile mc-router deployment")
		return ctrl.Result{}, err
	}

	// Reconcile mc-router Service
	if err := r.reconcileService(ctx); err != nil {
		log.Error(err, "failed to reconcile mc-router service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GatewayReconciler) ensureNamespace(ctx context.Context) error {
	ns := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: r.config.Namespace}, ns)
	if apierrors.IsNotFound(err) {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   r.config.Namespace,
				Labels: r.gatewayLabels(),
			},
		}
		return r.Create(ctx, ns)
	}
	return err
}

func (r *GatewayReconciler) reconcileServiceAccount(ctx context.Context) error {
	sa := &corev1.ServiceAccount{}
	sa.Namespace = r.config.Namespace
	sa.Name = r.config.ServiceAccount

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sa, func() error {
		sa.Labels = r.gatewayLabels()
		return nil
	})
	return err
}

func (r *GatewayReconciler) reconcileDeployment(ctx context.Context) error {
	deploy := &appsv1.Deployment{}
	deploy.Namespace = r.config.Namespace
	deploy.Name = constants.MCRouterAppName

	labels := r.gatewayLabels()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		deploy.Labels = labels
		deploy.Spec = appsv1.DeploymentSpec{
			Replicas: ptr.To[int32](1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: r.config.ServiceAccount,
					Containers: []corev1.Container{
						{
							Name:  constants.MCRouterAppName,
							Image: r.config.Image,
							Args: []string{
								"--in-kube-cluster",
								"--api-binding=:8080",
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          constants.MCRouterPortName,
									ContainerPort: constants.MCRouterPort,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          constants.MCRouterAPIPortName,
									ContainerPort: constants.MCRouterAPIPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromInt32(constants.MCRouterPort),
									},
								},
								InitialDelaySeconds: gatewayReadinessInitialDelaySeconds,
								PeriodSeconds:       gatewayReadinessPeriodSeconds,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromInt32(constants.MCRouterPort),
									},
								},
								InitialDelaySeconds: gatewayLivenessInitialDelaySeconds,
								PeriodSeconds:       gatewayLivenessPeriodSeconds,
							},
						},
					},
				},
			},
		}
		return nil
	})
	return err
}

func (r *GatewayReconciler) reconcileService(ctx context.Context) error {
	svc := &corev1.Service{}
	svc.Namespace = r.config.Namespace
	svc.Name = constants.MCRouterAppName

	labels := r.gatewayLabels()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Labels = labels
		svc.Spec.Type = r.config.ServiceType
		svc.Spec.Selector = labels

		// Preserve NodePort if already set
		var serverNodePort int32
		for _, p := range svc.Spec.Ports {
			if p.Name == constants.MCRouterPortName {
				serverNodePort = p.NodePort
			}
		}

		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       constants.MCRouterPortName,
				Port:       constants.MCRouterPort,
				TargetPort: intstr.FromString(constants.MCRouterPortName),
				Protocol:   corev1.ProtocolTCP,
				NodePort:   serverNodePort,
			},
		}
		return nil
	})
	return err
}

func (r *GatewayReconciler) gatewayLabels() map[string]string {
	return map[string]string{
		constants.LabelAppName:      constants.MCRouterAppName,
		constants.LabelAppComponent: constants.MCRouterAppComponent,
		constants.LabelAppCreatedBy: constants.ControllerName,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if !r.config.Enabled {
		return nil
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcingv1alpha1.Minecraft{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
