package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var _ = Describe("Minecraft controller", func() {
	namespace := "test"

	ctx := context.Background()
	var mgrCtx context.Context
	var mgrCancel context.CancelFunc

	BeforeEach(func() {
		ms := &mcingv1alpha1.MinecraftList{}
		err := k8sClient.List(ctx, ms, client.InNamespace(namespace))
		Expect(err).NotTo(HaveOccurred())
		for i := range ms.Items {
			m := &ms.Items[i]
			m.Finalizers = nil
			err := k8sClient.Update(ctx, m)
			Expect(err).NotTo(HaveOccurred())
		}
		svcs := &corev1.ServiceList{}
		err = k8sClient.List(ctx, svcs, client.InNamespace(namespace))
		Expect(err).NotTo(HaveOccurred())
		for i := range svcs.Items {
			svc := &svcs.Items[i]
			err := k8sClient.Delete(ctx, svc)
			Expect(err).NotTo(HaveOccurred())
		}
		err = k8sClient.DeleteAllOf(ctx, &mcingv1alpha1.Minecraft{}, client.InNamespace(namespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &appsv1.StatefulSet{}, client.InNamespace(namespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace(namespace))
		Expect(err).NotTo(HaveOccurred())

		mgr, err := ctrl.NewManager(k8sCfg, ctrl.Options{
			Scheme:         scheme,
			LeaderElection: false,
			Metrics:        metricsserver.Options{BindAddress: "0"},
		})
		Expect(err).ToNot(HaveOccurred())

		log := ctrl.Log.WithName("controllers")

		mockMinecraftMgr := &mockManager{}

		r := NewMinecraftReconciler(
			mgr.GetClient(),
			log,
			mgr.GetScheme(),
			"ghcr.io/kmdkuk/mcing-init:"+strings.TrimPrefix(version.Version, "v"),
			"ghcr.io/kmdkuk/mcing-agent:"+strings.TrimPrefix(version.Version, "v"),
			mockMinecraftMgr,
		)
		err = r.SetupWithManager(mgr)
		Expect(err).ToNot(HaveOccurred())

		mgrCtx, mgrCancel = context.WithCancel(context.Background())
		go func() {
			err := mgr.Start(mgrCtx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(time.Second)
	})

	AfterEach(func() {
		mgrCancel()
		time.Sleep(100 * time.Millisecond)
	})

	It("should create Namespace", func() {
		createNamespaces(ctx, namespace)
	})

	It("should create and delete minecrafts", func() {
		By("deploying Minecraft resource")
		mc := makeMinecraft("test", namespace)
		Expect(k8sClient.Create(ctx, mc)).To(Succeed())

		By("getting the created StatefulSet")
		s := new(appsv1.StatefulSet)
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: mc.PrefixedName(), Namespace: namespace}, s)
		}).Should(Succeed())

		// labels
		Expect(s.Labels).To(MatchAllKeys(Keys{
			constants.LabelAppName:      Equal(constants.AppName),
			constants.LabelAppComponent: Equal(constants.AppComponentServer),
			constants.LabelAppInstance:  Equal(mc.Name),
			constants.LabelAppCreatedBy: Equal(constants.ControllerName),
		}))
		Expect(s.Spec.Selector.MatchLabels).To(MatchAllKeys(Keys{
			constants.LabelAppName:      Equal(constants.AppName),
			constants.LabelAppComponent: Equal(constants.AppComponentServer),
			constants.LabelAppInstance:  Equal(mc.Name),
			constants.LabelAppCreatedBy: Equal(constants.ControllerName),
		}))
		Expect(s.Spec.Template.Labels).To(MatchAllKeys(Keys{
			constants.LabelAppName:      Equal(constants.AppName),
			constants.LabelAppComponent: Equal(constants.AppComponentServer),
			constants.LabelAppInstance:  Equal(mc.Name),
			constants.LabelAppCreatedBy: Equal(constants.ControllerName),
		}))

		// statefulset/pod spec
		Expect(s.Spec.Replicas).To(PointTo(BeNumerically("==", 1)))
		Expect(s.Spec.Template.Spec.Containers).To(HaveLen(2))
		Expect(s.Spec.Template.Spec.Containers[0]).To(MatchFields(IgnoreExtras, Fields{
			"Name":  Equal(constants.MinecraftContainerName),
			"Image": Equal(constants.DefaultServerImage),
			"Ports": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": MatchFields(IgnoreExtras, Fields{
					"Name":          Equal(constants.ServerPortName),
					"ContainerPort": Equal(constants.ServerPort),
					"Protocol":      Equal(corev1.ProtocolTCP),
				}),
				"1": MatchFields(IgnoreExtras, Fields{
					"Name":          Equal(constants.RconPortName),
					"ContainerPort": Equal(constants.RconPort),
					"Protocol":      Equal(corev1.ProtocolUDP),
				}),
			}),
			"VolumeMounts": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": MatchFields(IgnoreExtras, Fields{
					"Name":      Equal(constants.DataVolumeName),
					"MountPath": Equal(constants.DataPath),
				}),
				"1": MatchFields(IgnoreExtras, Fields{
					"Name":      Equal(constants.ConfigVolumeName),
					"MountPath": Equal(constants.ConfigPath),
				}),
			}),
		}))
		Expect(s.Spec.Template.Spec.Containers[1]).To(MatchFields(IgnoreExtras, Fields{
			"Name":  Equal(constants.AgentContainerName),
			"Image": Equal("ghcr.io/kmdkuk/mcing-agent:" + strings.TrimPrefix(version.Version, "v")),
			"Ports": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": MatchFields(IgnoreExtras, Fields{
					"Name":          Equal(constants.AgentPortName),
					"ContainerPort": Equal(constants.AgentPort),
					"Protocol":      Equal(corev1.ProtocolTCP),
				}),
			}),
			"VolumeMounts": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": MatchFields(IgnoreExtras, Fields{
					"Name":      Equal(constants.DataVolumeName),
					"MountPath": Equal(constants.DataPath),
				}),
				"1": MatchFields(IgnoreExtras, Fields{
					"Name":      Equal(constants.ConfigVolumeName),
					"MountPath": Equal(constants.ConfigPath),
				}),
			}),
		}))
		Expect(s.Spec.VolumeClaimTemplates).To(HaveLen(1))
		Expect(s.Spec.VolumeClaimTemplates[0].ObjectMeta.Name).To(Equal("minecraft-data"))
	})

	It("should update generated ConfigMap, when update specified ConfigMap", func() {
		By("deploying ConfigMap and Minecraft resource")
		testCmName := "test-configmap"
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCmName,
				Namespace: namespace,
			},
			Data: map[string]string{
				"motd":       "A vanila",
				"difficulty": "hard",
				"pvp":        "false",
			},
		}
		mc := makeMinecraft("test", namespace)
		mc.Spec.ServerPropertiesConfigMapName = &cm.Name
		Expect(k8sClient.Create(ctx, cm)).To(Succeed())
		Expect(k8sClient.Create(ctx, mc)).To(Succeed())

		By("getting generated ConfigMap")
		generatedCm := &corev1.ConfigMap{}
		Eventually(func() error {
			if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: mc.Namespace, Name: mc.PrefixedName()}, generatedCm); err != nil {
				return err
			}
			return nil
		}).Should(Succeed())
		By("updating ConfigMap")
		cm.Data = map[string]string{
			"motd":       "updated",
			"difficulty": "easy",
			"pvp":        "true",
		}
		Expect(k8sClient.Update(ctx, cm)).To(Succeed())

		By("getting generated ConfigMap")
		Eventually(func() error {
			cm := &corev1.ConfigMap{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: mc.Namespace, Name: mc.PrefixedName()}, cm); err != nil {
				return err
			}

			if !cmp.Equal(generatedCm.Data[constants.ServerPropsName], cm.Data[constants.ServerPropsName]) {
				return fmt.Errorf("The generated ConfigMap has not been updated.")
			}
			return nil
		}).Should(Succeed())
	})
})
