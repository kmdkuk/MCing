/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var minecraftlog = logf.Log.WithName("minecraft-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *Minecraft) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(&Minecraft{}).
		WithValidator(&Minecraft{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-mcing-kmdkuk-com-v1alpha1-minecraft,mutating=true,failurePolicy=fail,sideEffects=None,groups=mcing.kmdkuk.com,resources=minecrafts,verbs=create;update,versions=v1alpha1,name=mminecraft.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &Minecraft{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Minecraft) Default(ctx context.Context, obj runtime.Object) error {
	minecraftlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-mcing-kmdkuk-com-v1alpha1-minecraft,mutating=false,failurePolicy=fail,sideEffects=None,groups=mcing.kmdkuk.com,resources=minecrafts,verbs=create;update,versions=v1alpha1,name=vminecraft.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &Minecraft{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Minecraft) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	minecraftlog.Info("validate create", "name", r.Name)
	m := obj.(*Minecraft)
	errs := m.Spec.validateCreate()
	if len(errs) != 0 {
		return admission.Warnings{}, apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Minecraft"}, m.Name, errs)
	}

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Minecraft) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) (admission.Warnings, error) {
	minecraftlog.Info("validate update", "name", r.Name)

	oldM := old.(*Minecraft)
	newM := new.(*Minecraft)
	errs := newM.Spec.validateUpdate(oldM.Spec)
	if len(errs) != 0 {
		return admission.Warnings{}, apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Minecraft"}, newM.Name, errs)
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Minecraft) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	m := obj.(*Minecraft)
	minecraftlog.Info("validate delete", "name", m.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
