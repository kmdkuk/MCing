package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2" //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"    //nolint:revive // dot imports for tests
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kmdkuk/mcing/pkg/constants"
)

var _ = Describe("Minecraft Webhook", func() {
	var minecraft *Minecraft

	BeforeEach(func() {
		minecraft = &Minecraft{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-minecraft",
				Namespace: "default",
			},
			Spec: MinecraftSpec{
				VolumeClaimTemplates: []PersistentVolumeClaim{
					{
						ObjectMeta: ObjectMeta{
							Name: constants.DataVolumeName,
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							Resources: corev1.VolumeResourceRequirements{
								//nolint:exhaustive // test data
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
				PodTemplate: PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  constants.MinecraftContainerName,
								Image: "itzg/minecraft-server",
								Env: []corev1.EnvVar{
									{
										Name:  constants.EulaEnvName,
										Value: "true",
									},
								},
							},
						},
					},
				},
			},
		}
	})

	Context("ValidateCreate", func() {
		It("should validate a valid Minecraft resource", func() {
			warnings, err := minecraft.ValidateCreate(ctx, minecraft)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeEmpty())
		})

		It("should fail if minecraft-data PVC is missing", func() {
			minecraft.Spec.VolumeClaimTemplates = []PersistentVolumeClaim{}
			_, err := minecraft.ValidateCreate(ctx, minecraft)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("required volume claim template minecraft-data is missing"))
		})

		It("should fail if minecraft container is missing", func() {
			minecraft.Spec.PodTemplate.Spec.Containers[0].Name = "invalid-name"
			_, err := minecraft.ValidateCreate(ctx, minecraft)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("required container minecraft is missing"))
		})

		It("should fail if EULA is missing", func() {
			minecraft.Spec.PodTemplate.Spec.Containers[0].Env = []corev1.EnvVar{}
			_, err := minecraft.ValidateCreate(ctx, minecraft)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("EULA is required"))
		})

		It("should fail if reserved port 25565 is used", func() {
			minecraft.Spec.PodTemplate.Spec.Containers[0].Ports = []corev1.ContainerPort{
				{
					ContainerPort: constants.ServerPort,
				},
			}
			_, err := minecraft.ValidateCreate(ctx, minecraft)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("reserved port"))
		})

		It("should fail if reserved port 25575 is used", func() {
			minecraft.Spec.PodTemplate.Spec.Containers[0].Ports = []corev1.ContainerPort{
				{
					ContainerPort: constants.RconPort,
				},
			}
			_, err := minecraft.ValidateCreate(ctx, minecraft)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("reserved port"))
		})
	})

	Context("ValidateUpdate", func() {
		var oldMinecraft *Minecraft

		BeforeEach(func() {
			oldMinecraft = minecraft.DeepCopy()
		})

		It("should validate a valid update", func() {
			minecraft.Spec.PodTemplate.Spec.Containers[0].Image = "itzg/minecraft-server:latest"
			warnings, err := minecraft.ValidateUpdate(ctx, oldMinecraft, minecraft)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeEmpty())
		})

		It("should fail if update creates an invalid state (missing EULA)", func() {
			minecraft.Spec.PodTemplate.Spec.Containers[0].Env = []corev1.EnvVar{}
			_, err := minecraft.ValidateUpdate(ctx, oldMinecraft, minecraft)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("EULA is required"))
		})
	})
})
