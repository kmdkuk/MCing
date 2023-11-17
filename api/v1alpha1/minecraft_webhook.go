package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var minecraftlog = logf.Log.WithName("minecraft-resource")

func (r *Minecraft) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-mcing-kmdkuk-com-v1alpha1-minecraft,mutating=true,failurePolicy=fail,sideEffects=None,groups=mcing.kmdkuk.com,resources=minecrafts,verbs=create;update,versions=v1alpha1,name=mminecraft.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Minecraft{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Minecraft) Default() {
	minecraftlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-mcing-kmdkuk-com-v1alpha1-minecraft,mutating=false,failurePolicy=fail,sideEffects=None,groups=mcing.kmdkuk.com,resources=minecrafts,verbs=create;update,versions=v1alpha1,name=vminecraft.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Minecraft{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Minecraft) ValidateCreate() (admission.Warnings, error) {
	errs := r.Spec.validateCreate()
	if len(errs) != 0 {
		return admission.Warnings{}, apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Minecraft"}, r.Name, errs)
	}

	return admission.Warnings{}, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Minecraft) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	errs := r.Spec.validateUpdate(old.(*Minecraft).Spec)
	if len(errs) != 0 {
		return admission.Warnings{}, apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Minecraft"}, r.Name, errs)
	}

	return admission.Warnings{}, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Minecraft) ValidateDelete() (admission.Warnings, error) {
	return admission.Warnings{}, nil
}
