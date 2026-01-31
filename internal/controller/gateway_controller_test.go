package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"      //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"         //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega/gstruct" //nolint:revive // dot imports for tests
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/config"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/kmdkuk/mcing/pkg/constants"
)

var _ = Describe("Gateway controller", func() {
	const (
		gatewayNamespace      = "mcing-gateway"
		gatewayServiceAccount = "mc-router"
		defaultDomain         = "minecraft.local"
		mcRouterImage         = "itzg/mc-router:latest"
	)

	ctx := context.Background()
	var mgrCtx context.Context
	var mgrCancel context.CancelFunc

	cleanupGatewayResources := func() {
		// Delete deployment if exists
		deploy := &appsv1.Deployment{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Namespace: gatewayNamespace,
			Name:      constants.MCRouterAppName,
		}, deploy)
		if err == nil {
			_ = k8sClient.Delete(ctx, deploy)
		}

		// Delete service if exists
		svc := &corev1.Service{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Namespace: gatewayNamespace,
			Name:      constants.MCRouterAppName,
		}, svc)
		if err == nil {
			_ = k8sClient.Delete(ctx, svc)
		}

		// Delete service account if exists
		sa := &corev1.ServiceAccount{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Namespace: gatewayNamespace,
			Name:      gatewayServiceAccount,
		}, sa)
		if err == nil {
			_ = k8sClient.Delete(ctx, sa)
		}

		// Note: We don't delete the namespace because it takes too long
		// due to namespace finalizers. The gateway tests will reuse the namespace.
	}

	setupGatewayController := func(enabled bool, serviceType corev1.ServiceType) {
		cleanupGatewayResources()

		mgr, err := ctrl.NewManager(k8sCfg, ctrl.Options{
			Scheme:         scheme,
			LeaderElection: false,
			Metrics:        metricsserver.Options{BindAddress: "0"},
			Controller: config.Controller{
				SkipNameValidation: ptr.To(true),
			},
		})
		Expect(err).ToNot(HaveOccurred())

		log := ctrl.Log.WithName("controllers")

		gatewayConfig := GatewayConfig{
			Enabled:        enabled,
			DefaultDomain:  defaultDomain,
			Namespace:      gatewayNamespace,
			ServiceAccount: gatewayServiceAccount,
			ServiceType:    serviceType,
			Image:          mcRouterImage,
		}

		r := NewGatewayReconciler(
			mgr.GetClient(),
			log,
			mgr.GetScheme(),
			gatewayConfig,
		)
		err = r.SetupWithManager(mgr)
		Expect(err).ToNot(HaveOccurred())

		mgrCtx, mgrCancel = context.WithCancel(context.Background()) //nolint:fatcontext // test logic
		go func() {
			err := mgr.Start(mgrCtx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(time.Second)
	}

	teardownManager := func() {
		mgrCancel()
		time.Sleep(100 * time.Millisecond)
	}

	Context("when mc-router is enabled", func() {
		BeforeEach(func() {
			setupGatewayController(true, corev1.ServiceTypeLoadBalancer)
		})

		AfterEach(func() {
			teardownManager()
		})

		It("should create mcing-gateway namespace", func() {
			// Create a Minecraft resource to trigger reconciliation
			mc := makeMinecraft("gateway-test", "default")
			Expect(k8sClient.Create(ctx, mc)).To(Succeed())
			defer func() {
				mc.Finalizers = nil
				_ = k8sClient.Update(ctx, mc)
				_ = k8sClient.Delete(ctx, mc)
			}()

			Eventually(func(g Gomega) {
				ns := &corev1.Namespace{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: gatewayNamespace}, ns)
				g.Expect(err).ShouldNot(HaveOccurred())
				g.Expect(ns.Labels).To(HaveKeyWithValue(constants.LabelAppName, constants.MCRouterAppName))
				g.Expect(ns.Labels).To(HaveKeyWithValue(constants.LabelAppComponent, constants.MCRouterAppComponent))
			}).Should(Succeed())
		})

		It("should create mc-router service account", func() {
			mc := makeMinecraft("gateway-test-sa", "default")
			Expect(k8sClient.Create(ctx, mc)).To(Succeed())
			defer func() {
				mc.Finalizers = nil
				_ = k8sClient.Update(ctx, mc)
				_ = k8sClient.Delete(ctx, mc)
			}()

			Eventually(func(g Gomega) {
				sa := &corev1.ServiceAccount{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: gatewayNamespace,
					Name:      gatewayServiceAccount,
				}, sa)
				g.Expect(err).ShouldNot(HaveOccurred())
				g.Expect(sa.Labels).To(HaveKeyWithValue(constants.LabelAppName, constants.MCRouterAppName))
			}).Should(Succeed())
		})

		It("should create mc-router deployment", func() {
			mc := makeMinecraft("gateway-test-deploy", "default")
			Expect(k8sClient.Create(ctx, mc)).To(Succeed())
			defer func() {
				mc.Finalizers = nil
				_ = k8sClient.Update(ctx, mc)
				_ = k8sClient.Delete(ctx, mc)
			}()

			Eventually(func(g Gomega) {
				deploy := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: gatewayNamespace,
					Name:      constants.MCRouterAppName,
				}, deploy)
				g.Expect(err).ShouldNot(HaveOccurred())

				// Check deployment labels
				g.Expect(deploy.Labels).To(HaveKeyWithValue(constants.LabelAppName, constants.MCRouterAppName))

				// Check container
				g.Expect(deploy.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := deploy.Spec.Template.Spec.Containers[0]
				g.Expect(container.Name).To(Equal(constants.MCRouterAppName))
				g.Expect(container.Image).To(Equal(mcRouterImage))

				// Check args
				g.Expect(container.Args).To(ContainElements("--in-kube-cluster", "--api-binding=:8080"))

				// Check ports
				g.Expect(container.Ports).To(ContainElements(
					MatchFields(IgnoreExtras, Fields{
						"Name":          Equal(constants.MCRouterPortName),
						"ContainerPort": Equal(constants.MCRouterPort),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Name":          Equal(constants.MCRouterAPIPortName),
						"ContainerPort": Equal(constants.MCRouterAPIPort),
					}),
				))

				// Check service account
				g.Expect(deploy.Spec.Template.Spec.ServiceAccountName).To(Equal(gatewayServiceAccount))
			}).Should(Succeed())
		})

		It("should create mc-router service with LoadBalancer type", func() {
			mc := makeMinecraft("gateway-test-svc", "default")
			Expect(k8sClient.Create(ctx, mc)).To(Succeed())
			defer func() {
				mc.Finalizers = nil
				_ = k8sClient.Update(ctx, mc)
				_ = k8sClient.Delete(ctx, mc)
			}()

			Eventually(func(g Gomega) {
				svc := &corev1.Service{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: gatewayNamespace,
					Name:      constants.MCRouterAppName,
				}, svc)
				g.Expect(err).ShouldNot(HaveOccurred())

				// Check service type
				g.Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))

				// Check labels
				g.Expect(svc.Labels).To(HaveKeyWithValue(constants.LabelAppName, constants.MCRouterAppName))

				// Check port
				g.Expect(svc.Spec.Ports).To(ContainElement(
					MatchFields(IgnoreExtras, Fields{
						"Name": Equal(constants.MCRouterPortName),
						"Port": Equal(constants.MCRouterPort),
					}),
				))
			}).Should(Succeed())
		})
	})

	Context("when mc-router is enabled with NodePort", func() {
		BeforeEach(func() {
			setupGatewayController(true, corev1.ServiceTypeNodePort)
		})

		AfterEach(func() {
			teardownManager()
		})

		It("should create mc-router service with NodePort type", func() {
			mc := makeMinecraft("gateway-test-nodeport", "default")
			Expect(k8sClient.Create(ctx, mc)).To(Succeed())
			defer func() {
				mc.Finalizers = nil
				_ = k8sClient.Update(ctx, mc)
				_ = k8sClient.Delete(ctx, mc)
			}()

			Eventually(func(g Gomega) {
				svc := &corev1.Service{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: gatewayNamespace,
					Name:      constants.MCRouterAppName,
				}, svc)
				g.Expect(err).ShouldNot(HaveOccurred())

				// Check service type
				g.Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
			}).Should(Succeed())
		})
	})

	Context("when mc-router is disabled", func() {
		BeforeEach(func() {
			setupGatewayController(false, corev1.ServiceTypeLoadBalancer)
		})

		AfterEach(func() {
			teardownManager()
		})

		It("should not create deployment when disabled", func() {
			// Wait a bit to ensure nothing is created
			time.Sleep(2 * time.Second)

			// Verify deployment is not created (don't check namespace as it may exist from previous tests)
			deploy := &appsv1.Deployment{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Namespace: gatewayNamespace,
				Name:      constants.MCRouterAppName,
			}, deploy)
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})
	})
})
