package controllers

import (
	"context"

	"github.com/go-logr/logr"
	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MinecraftReconciler reconciles a Minecraft object
type MinecraftReconciler struct {
	client.Client
	log    logr.Logger
	scheme *runtime.Scheme
}

func NewMinecraftReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *MinecraftReconciler {
	l := log.WithName("Minecraft")
	return &MinecraftReconciler{
		Client: client,
		log:    l,
		scheme: scheme,
	}
}

//+kubebuilder:rbac:groups=mcing.kmdkuk.com,resources=minecrafts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mcing.kmdkuk.com,resources=minecrafts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mcing.kmdkuk.com,resources=minecrafts/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

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

	if mc.ObjectMeta.DeletionTimestamp != nil {
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

	if err := r.reconcileStatefulSet(ctx, mc); err != nil {
		log.Error(err, "failed to reconcile statefulset")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *MinecraftReconciler) reconcileStatefulSet(ctx context.Context, mc *mcingv1alpha1.Minecraft) error {
	logger := r.log.WithName("statefulset")

	owner, err := ownerRef(mc, r.scheme)
	if err != nil {
		return err
	}

	labels := labelSet(mc, constants.AppComponentServer)

	sts := appsv1apply.StatefulSet(mc.Name, mc.Namespace).
		WithLabels(labels).
		WithOwnerReferences(owner).
		WithSpec(appsv1apply.StatefulSetSpec().
			WithReplicas(1).
			WithSelector(metav1apply.LabelSelector().WithMatchLabels(labels)).
			WithTemplate(corev1apply.PodTemplateSpec().
				WithLabels(labels).
				WithSpec(corev1apply.PodSpec().
					WithContainers(serverContainer(mc.Spec.Image)),
				),
			).
			WithVolumeClaimTemplates(corev1apply.PersistentVolumeClaim("minecraft-data", mc.Namespace).
				WithSpec(corev1apply.PersistentVolumeClaimSpec().
					WithAccessModes(corev1.ReadWriteOnce).
					WithResources(corev1apply.ResourceRequirements().
						WithLimits(mc.Spec.VolumeClaimSpec.Resources.Limits).
						WithRequests(mc.Spec.VolumeClaimSpec.Resources.Requests),
					),
				),
			),
		)

	if mc.Spec.VolumeClaimSpec.StorageClassName != nil {
		sts.Spec.VolumeClaimTemplates[0].Spec = sts.Spec.VolumeClaimTemplates[0].Spec.WithStorageClassName(*mc.Spec.VolumeClaimSpec.StorageClassName)
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sts)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current appsv1.StatefulSet
	err = r.Get(ctx, types.NamespacedName{Namespace: mc.Namespace, Name: mc.Name}, &current)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	currApplyConfig, err := appsv1apply.ExtractStatefulSet(&current, constants.ControllerName)
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(sts, currApplyConfig) {
		return nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: constants.ControllerName,
		Force:        pointer.Bool(true),
	})
	if err != nil {
		logger.Error(err, "unable to create or update StatefulSet")
		return err
	}

	logger.Info("reconcile StatefulSet successfully")
	return nil
}

func serverContainer(image string) *corev1apply.ContainerApplyConfiguration {
	i := constants.DefaultServerImage
	if image != "" {
		i = image
	}
	return corev1apply.Container().
		WithName(constants.ServerContainerName).
		WithImage(i).
		WithCommand("pause").
		WithPorts(
			corev1apply.ContainerPort().
				WithName("server-port").
				WithProtocol(corev1.ProtocolTCP).
				WithContainerPort(25565),
			corev1apply.ContainerPort().
				WithName("rcon-port").
				WithProtocol(corev1.ProtocolUDP).
				WithContainerPort(25575),
		).
		WithVolumeMounts(
			corev1apply.VolumeMount().
				WithName("minecraft-data").
				WithMountPath("/minecraft-data"),
		)
}

func ownerRef(mc *mcingv1alpha1.Minecraft, scheme *runtime.Scheme) (*metav1apply.OwnerReferenceApplyConfiguration, error) {
	gvk, err := apiutil.GVKForObject(mc, scheme)
	if err != nil {
		return nil, err
	}
	return metav1apply.OwnerReference().
		WithAPIVersion(gvk.GroupVersion().String()).
		WithKind(gvk.Kind).
		WithName(mc.Name).
		WithUID(mc.GetUID()).
		WithBlockOwnerDeletion(true).
		WithController(true), nil
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcingv1alpha1.Minecraft{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
