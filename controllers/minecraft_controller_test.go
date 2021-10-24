package controllers

import (
	"context"
	"time"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/constants"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
			Scheme:             scheme,
			LeaderElection:     false,
			MetricsBindAddress: "0",
		})
		Expect(err).ToNot(HaveOccurred())

		log := ctrl.Log.WithName("controllers")

		r := NewMinecraftReconciler(
			mgr.GetClient(),
			log,
			mgr.GetScheme(),
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
			return k8sClient.Get(ctx, types.NamespacedName{Name: mc.Name, Namespace: namespace}, s)
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
		Expect(s.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(s.Spec.Template.Spec.Containers[0]).To(MatchFields(IgnoreExtras, Fields{
			"Name":  Equal(constants.ServerContainerName),
			"Image": Equal(constants.DefaultServerImage),
			"Ports": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": MatchFields(IgnoreExtras, Fields{
					"Name":          Equal("server-port"),
					"ContainerPort": Equal(int32(25565)),
					"Protocol":      Equal(corev1.ProtocolTCP),
				}),
				"1": MatchFields(IgnoreExtras, Fields{
					"Name":          Equal("rcon-port"),
					"ContainerPort": Equal(int32(25575)),
					"Protocol":      Equal(corev1.ProtocolUDP),
				}),
			}),
			"VolumeMounts": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": MatchFields(IgnoreExtras, Fields{
					"Name":      Equal("minecraft-data"),
					"MountPath": Equal("/data"),
				}),
			}),
		}))
		Expect(s.Spec.VolumeClaimTemplates).To(HaveLen(1))
		Expect(s.Spec.VolumeClaimTemplates[0].ObjectMeta.Name).To(Equal("minecraft-data"))
	})
})
