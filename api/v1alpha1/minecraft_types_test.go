package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMinecraft_PrefixedName(t *testing.T) {
	m := &Minecraft{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}
	want := "mcing-test"
	if got := m.PrefixedName(); got != want {
		t.Errorf("Minecraft.PrefixedName() = %v, want %v", got, want)
	}
}

func TestMinecraft_PodName(t *testing.T) {
	m := &Minecraft{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}
	want := "mcing-test-0"
	if got := m.PodName(); got != want {
		t.Errorf("Minecraft.PodName() = %v, want %v", got, want)
	}
}

func TestMinecraft_HeadlessServiceName(t *testing.T) {
	m := &Minecraft{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}
	want := "mcing-test-headless"
	if got := m.HeadlessServiceName(); got != want {
		t.Errorf("Minecraft.HeadlessServiceName() = %v, want %v", got, want)
	}
}

func TestMinecraft_RconSecretName(t *testing.T) {
	tests := []struct {
		name string
		m    *Minecraft
		want string
	}{
		{
			name: "default",
			m: &Minecraft{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: MinecraftSpec{
					RconPasswordSecretName: nil,
				},
			},
			want: "mcing-test-rcon-password",
		},
		{
			name: "custom",
			m: &Minecraft{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: MinecraftSpec{
					RconPasswordSecretName: stringPtr("custom-secret"),
				},
			},
			want: "custom-secret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.RconSecretName(); got != tt.want {
				t.Errorf("Minecraft.RconSecretName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodTemplateSpec_ToCoreV1(t *testing.T) {
	p := &PodTemplateSpec{
		ObjectMeta: ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"app": "minecraft",
			},
			Annotations: map[string]string{
				"test": "annotation",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "minecraft"},
			},
		},
	}
	want := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"app": "minecraft",
			},
			Annotations: map[string]string{
				"test": "annotation",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "minecraft"},
			},
		},
	}

	got := p.ToCoreV1()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("PodTemplateSpec.ToCoreV1() mismatch (-want +got):\n%s", diff)
	}
}

func TestPersistentVolumeClaim_ToCoreV1(t *testing.T) {
	modeFilesystem := corev1.PersistentVolumeFilesystem
	p := &PersistentVolumeClaim{
		ObjectMeta: ObjectMeta{
			Name: "test-pvc",
			Labels: map[string]string{
				"app": "minecraft",
			},
			Annotations: map[string]string{
				"test": "annotation",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		},
	}
	want := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pvc",
			Labels: map[string]string{
				"app": "minecraft",
			},
			Annotations: map[string]string{
				"test": "annotation",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			VolumeMode:  &modeFilesystem,
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimPending,
		},
	}

	got := p.ToCoreV1()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("PersistentVolumeClaim.ToCoreV1() mismatch (-want +got):\n%s", diff)
	}
}

func stringPtr(s string) *string {
	return &s
}
