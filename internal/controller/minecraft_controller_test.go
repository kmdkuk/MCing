package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"      //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega"         //nolint:revive // dot imports for tests
	. "github.com/onsi/gomega/gstruct" //nolint:revive // dot imports for tests
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/version"
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
			err = k8sClient.Update(ctx, m)
			Expect(err).NotTo(HaveOccurred())
		}
		svcs := &corev1.ServiceList{}
		err = k8sClient.List(ctx, svcs, client.InNamespace(namespace))
		Expect(err).NotTo(HaveOccurred())
		for i := range svcs.Items {
			svc := &svcs.Items[i]
			err = k8sClient.Delete(ctx, svc)
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
			Controller: config.Controller{
				SkipNameValidation: ptr.To(true),
			},
		})
		Expect(err).ToNot(HaveOccurred())

		log := ctrl.Log.WithName("controllers")

		mockMinecraftMgr := &mockManager{ //nolint:exhaustruct // internal struct
			minecrafts: make(map[string]struct{}),
		}

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

		mgrCtx, mgrCancel = context.WithCancel(context.Background()) //nolint:fatcontext // test logic
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
			"Lifecycle": PointTo(MatchFields(IgnoreExtras, Fields{
				"PreStop": PointTo(MatchFields(IgnoreExtras, Fields{
					"Exec": PointTo(MatchFields(IgnoreExtras, Fields{
						"Command": Equal([]string{
							"/bin/sh",
							"-c",
							"rcon-cli stop || true",
						}),
					})),
				})),
			})),
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
			if err := k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: mc.Namespace, Name: mc.PrefixedName()},
				generatedCm,
			); err != nil {
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
			if err := k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: mc.Namespace, Name: mc.PrefixedName()},
				cm,
			); err != nil {
				return err
			}

			if !cmp.Equal(generatedCm.Data[constants.ServerPropsName], cm.Data[constants.ServerPropsName]) {
				return errors.New("the generated ConfigMap has not been updated")
			}
			return nil
		}).Should(Succeed())
	})
	Context("RCON Secret", func() {
		It("should create default RCON secret if not specified", func() {
			mc := makeMinecraft("default-rcon", namespace)
			Expect(k8sClient.Create(ctx, mc)).To(Succeed())

			By("checking generated Secret")
			secret := &corev1.Secret{}
			Eventually(func() error {
				return k8sClient.Get(
					ctx,
					types.NamespacedName{Name: mc.PrefixedName() + "-rcon-password", Namespace: namespace},
					secret,
				)
			}).Should(Succeed())
			Expect(secret.Data).To(HaveKey(constants.RconPasswordSecretKey))

			By("checking StatefulSet env var")
			sts := &appsv1.StatefulSet{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: mc.PrefixedName(), Namespace: namespace}, sts)
			}).Should(Succeed())

			Expect(sts.Spec.Template.Spec.Containers[0].Env).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(constants.RconPasswordEnvName),
				"ValueFrom": PointTo(MatchFields(IgnoreExtras, Fields{
					"SecretKeyRef": PointTo(MatchFields(IgnoreExtras, Fields{
						"LocalObjectReference": MatchFields(IgnoreExtras, Fields{
							"Name": Equal(secret.Name),
						}),
						"Key": Equal(constants.RconPasswordSecretKey),
					})),
				})),
			})))
		})

		It("should use specified RCON secret", func() {
			secretName := "my-rcon-secret"
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					constants.RconPasswordSecretKey: []byte("password"),
				},
			}
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			mc := makeMinecraft("custom-rcon", namespace)
			mc.Spec.RconPasswordSecretName = &secretName
			Expect(k8sClient.Create(ctx, mc)).To(Succeed())

			By("checking StatefulSet env var")
			sts := &appsv1.StatefulSet{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: mc.PrefixedName(), Namespace: namespace}, sts)
			}).Should(Succeed())

			Expect(sts.Spec.Template.Spec.Containers[0].Env).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(constants.RconPasswordEnvName),
				"ValueFrom": PointTo(MatchFields(IgnoreExtras, Fields{
					"SecretKeyRef": PointTo(MatchFields(IgnoreExtras, Fields{
						"LocalObjectReference": MatchFields(IgnoreExtras, Fields{
							"Name": Equal(secretName),
						}),
						"Key": Equal(constants.RconPasswordSecretKey),
					})),
				})),
			})))

			By("ensuring default secret is not created")
			Consistently(func() error {
				return k8sClient.Get(
					ctx,
					types.NamespacedName{Name: mc.PrefixedName() + "-rcon-password", Namespace: namespace},
					&corev1.Secret{},
				)
			}).ShouldNot(Succeed())
		})
	})
	It("should enable auto-pause configurations", func() {
		By("creating ConfigMap with custom port")
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "autopause-config",
				Namespace: namespace,
			},
			Data: map[string]string{
				"server-port": "12345", // This property setting is ignored.
				"motd":        "AutoPause Test",
			},
		}
		Expect(k8sClient.Create(ctx, cm)).To(Succeed())

		By("deploying Minecraft resource with AutoPause enabled")
		mc := makeMinecraft("autopause-test", namespace)
		mc.Spec.ServerPropertiesConfigMapName = &cm.Name
		mc.Spec.AutoPause = mcingv1alpha1.AutoPause{
			Enabled:        true,
			TimeoutSeconds: 600,
		}
		Expect(k8sClient.Create(ctx, mc)).To(Succeed())

		By("getting the created StatefulSet")
		s := new(appsv1.StatefulSet)
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: mc.PrefixedName(), Namespace: namespace}, s)
		}).Should(Succeed())

		// Verify ConfigMap content for lazymc.toml
		generatedCm := &corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: mc.Namespace, Name: mc.PrefixedName()},
				generatedCm,
			)
		}).Should(Succeed())
		val, ok := generatedCm.Data[constants.LazymcConfigName]
		Expect(ok).To(BeTrue())
		Expect(val).To(ContainSubstring(fmt.Sprintf("address = \"0.0.0.0:%d\"", constants.ServerPort)))
		Expect(val).To(ContainSubstring(fmt.Sprintf("address = \"127.0.0.1:%d\"", constants.InternalServerPort)))
		Expect(val).To(ContainSubstring("sleep_after = 600"))

		// Verify Main Container Command
		Expect(s.Spec.Template.Spec.Containers[0].Command).To(Equal([]string{"/opt/lazymc/lazymc"}))
		Expect(s.Spec.Template.Spec.Containers[0].Args).To(
			Equal([]string{"--config", "/opt/lazymc/lazymc.toml"}),
		)

		// Verify Probes use tcpSocket
		Expect(s.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket).NotTo(BeNil())
		Expect(
			s.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket.Port.IntVal,
		).To(Equal(constants.ServerPort))
		Expect(s.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket).NotTo(BeNil())
		Expect(
			s.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket.Port.IntVal,
		).To(Equal(constants.ServerPort))

		// Verify ConfigMap override
		generatedCm = &corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: mc.Namespace, Name: mc.PrefixedName()},
				generatedCm,
			)
		}).Should(Succeed())
		val, ok = generatedCm.Data[constants.ServerPropsName]
		Expect(ok).To(BeTrue())
		Expect(val).To(ContainSubstring(fmt.Sprintf("server-port=%d", constants.InternalServerPort)))
		Expect(val).To(ContainSubstring("motd=AutoPause Test"))
	})

	It("should disable auto-pause configurations", func() {
		By("deploying Minecraft resource with AutoPause disabled")
		mc := makeMinecraft("no-autopause-test", namespace)
		mc.Spec.AutoPause = mcingv1alpha1.AutoPause{
			Enabled:        false,
			TimeoutSeconds: 600,
		}
		Expect(k8sClient.Create(ctx, mc)).To(Succeed())

		By("getting the created StatefulSet")
		s := new(appsv1.StatefulSet)
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: mc.PrefixedName(), Namespace: namespace}, s)
		}).Should(Succeed())

		// Verify ConfigMap content for lazymc.toml
		generatedCm := &corev1.ConfigMap{}
		Eventually(func() error {
			err := k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: mc.Namespace, Name: mc.PrefixedName()},
				generatedCm,
			)
			if err != nil {
				return err
			}
			_, ok := generatedCm.Data[constants.LazymcConfigName]
			if ok {
				return fmt.Errorf("lazymc.toml found in %s", generatedCm.Name)
			}
			return nil
		}).Should(Succeed())

		// Verify Main Container Command what dont use lazymc
		Expect(s.Spec.Template.Spec.Containers[0].Command).To(BeEmpty())
		Expect(s.Spec.Template.Spec.Containers[0].Args).To(BeEmpty())

		// Verify Probes use mc-health
		Expect(s.Spec.Template.Spec.Containers[0].LivenessProbe.Exec).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Command": Equal([]string{"mc-health"}),
		})))
		Expect(s.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Command": Equal([]string{"mc-health"}),
		})))
	})
})
