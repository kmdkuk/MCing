package cmd

import (

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	_ "k8s.io/client-go/plugin/pkg/client/auth" // required for authentication

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/go-logr/logr"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/internal/controller"
	"github.com/kmdkuk/mcing/internal/minecraft"
	"github.com/kmdkuk/mcing/pkg/agent"
	//+kubebuilder:scaffold:imports
)

// subMain is the actual main function.
func subMain(config Config) error {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&config.zapOpts)))
	setupLog := ctrl.Log.WithName("setup")
	mcMgrLog := ctrl.Log.WithName("minecraft-manager")

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(mcingv1alpha1.AddToScheme(scheme))

	webhookCertWatcher, webhookTLSOpts, err := setupWebhookCertWatcher(config, setupLog)
	if err != nil {
		setupLog.Error(err, "Failed to initialize webhook certificate watcher")
		os.Exit(1)
	}

	webhookAddr := ":9443"
	host, p, err := net.SplitHostPort(webhookAddr)
	if err != nil {
		return fmt.Errorf("invalid webhook address: %s, %w", webhookAddr, err)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return fmt.Errorf("invalid webhook address: %s, %w", webhookAddr, err)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		Host:    host,
		Port:    port,
		TLSOpts: webhookTLSOpts,
	})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: config.metricsAddr},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: config.probeAddr,
		LeaderElection:         config.enableLeaderElection,
		LeaderElectionID:       "6f987ab0.kmdkuk.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	af := agent.NewFactory()

	minecraftMgr := minecraft.NewManager(af, config.interval, mgr, mcMgrLog)

	if err = (controller.NewMinecraftReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers"),
		mgr.GetScheme(),
		config.initImageName,
		config.agentImageName,
		minecraftMgr,
	)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Minecraft")
		return err
	}

	if err = (&mcingv1alpha1.Minecraft{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Minecraft")
		return err
	}

	if webhookCertWatcher != nil {
		setupLog.Info("Adding webhook certificate watcher to manager")
		if err := mgr.Add(webhookCertWatcher); err != nil {
			setupLog.Error(err, "unable to add webhook certificate watcher to manager")
			return err
		}
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return err
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}

func setupWebhookCertWatcher(
	config Config,
	setupLog logr.Logger,
) (*certwatcher.CertWatcher, []func(*tls.Config), error) {
	// Create watchers for metrics and webhooks certificates
	var webhookCertWatcher *certwatcher.CertWatcher
	var tlsOpts []func(*tls.Config)
	// Initial webhook TLS options
	webhookTLSOpts := tlsOpts

	if len(config.webhookCertPath) > 0 {
		setupLog.Info(
			"Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path",
			config.webhookCertPath,
			"webhook-cert-name",
			config.webhookCertName,
			"webhook-cert-key",
			config.webhookCertKey,
		)

		var err error
		webhookCertWatcher, err = certwatcher.New(
			filepath.Join(config.webhookCertPath, config.webhookCertName),
			filepath.Join(config.webhookCertPath, config.webhookCertKey),
		)
		if err != nil {
			return nil, nil, err
		}

		webhookTLSOpts = append(webhookTLSOpts, func(config *tls.Config) {
			config.GetCertificate = webhookCertWatcher.GetCertificate
		})
	}
	return webhookCertWatcher, webhookTLSOpts, nil
}
