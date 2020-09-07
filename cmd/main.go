package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubernetesscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/klogr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	_ "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/safanaj/k8s-generic-validator/pkg/reconcilers"
	"github.com/safanaj/k8s-generic-validator/pkg/utils/configuration"
	"github.com/safanaj/k8s-generic-validator/pkg/utils/predicates"
	utilstls "github.com/safanaj/k8s-generic-validator/pkg/utils/tls"
	utilswebhook "github.com/safanaj/k8s-generic-validator/pkg/utils/webhook"
	"github.com/safanaj/k8s-generic-validator/pkg/webhooks"
)

var version string
var progname string
var log = logf.Log.WithName(progname)
var scheme *runtime.Scheme

func main() {
	// by default controller-runtime is using k8s.io/client-go/kubernetes/scheme
	// we can just modify that to add certmanager apis to avoid to pass any additional
	// options to the manager and/or client
	scheme = kubernetesscheme.Scheme

	// logf.SetLogger(zap.Logger(false))
	logf.SetLogger(klogr.New())
	entryLog := log.WithName("entrypoint")

	flags := parseFlags()
	if flags.version {
		entryLog.Info(fmt.Sprintf("%s %s", progname, version))
		os.Exit(0)
	}

	// setup configuration
	cfg := config.NewConfig()
	var cmNamed types.NamespacedName
	{
		cmParts := strings.Split(flags.configMap, "/")
		cmNamed = types.NamespacedName{cmParts[0], cmParts[1]}
	}

	// webhook server tls settings
	// this is matching the default from the webhook server
	certDir := filepath.Join(os.TempDir(), "k8s-webhook-server", "serving-certs")
	certName := "tls.crt"
	keyName := "tls.key"

	// Setup a Manager
	entryLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{CertDir: certDir})
	if err != nil {
		entryLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	// initial configuration parsing
	if err := configuration.EnsureFirstConfigurationLoad(
		cmNamed, mgr.GetAPIReader(), cfg,
		reconfilers.ConfigurationConfigMapKey); err != nil {
		entryLog.Error(err, "unable to set up initial configuration")
		os.Exit(1)
	}

	// setup reconcilers to keep configuration up-to-date
	builder.
		ControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(predicates.GetConfigPredicates(cmNamed)).
		Complete(reconcilers.NewConfigurationReconciler(
			log.WithName("configurationReconciler"),
			cfg))

	// setup all TLS and webhook configuration related stuff
	if len(flags.webhookCertificate) > 0 {
		needCert, err := utilstls.EnsureWeNeedCertificateByCertManager(certDir, certName, keyName)
		if err != nil {
			log.Error(err, "could not detect if cert-manager is needed")
		}
		if err == nil && needCert {
			utilstls.SetupScheme(scheme)
			// we won't have a certificate mounted in the pod, we will reconcile that cert-manager.io Certificate
			// writing the generate secret data in files used by the webhook server, we need a controller/reconciler
			entryLog.Info("Setting up Webhook certificate controllers")
			certParts := strings.Split(flags.webhookCertificate, "/")
			caParts := strings.Split(flags.webhookCAIssuer, "/")
			svcParts := strings.Split(flags.serviceName, "/")
			caIssuerNamed := types.NamespacedName{caParts[0], caParts[1]}
			certNamed := types.NamespacedName{certParts[0], certParts[1]}
			secNamed := types.NamespacedName{certParts[0], certParts[1]}
			svcNamed := types.NamespacedName{svcParts[0], svcParts[1]}

			log.Info("Gonna setup cert", "certDir", certDir, "certName", certName, "keyName", keyName)

			if err := utilstls.SetupCertificateByCertManager(
				log.WithName("utilstls"), mgr.GetAPIReader(), mgr.GetClient(),
				svcNamed, caIssuerNamed, certNamed, secNamed,
				certDir, certName, keyName); err != nil {
				log.Error(err, "Failed to setup certificate for cert-manager")
				os.Exit(1)
			}

			if err := utilstls.SetupWebhookTlsSecretControllersOrDie(
				secNamed, mgr, reconcilers.NewCertificateSecretReconciler(
					log.WithName("certificateSecretReconciler"),
					certDir, certName, keyName)); err != nil {
				log.Error(err, "Failed to setup Webhook TLS Secret Controller")
				os.Exit(1)
			}
		}
	}

	// Setup webhooks
	entryLog.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()

	// ensure validating/mutating webhook configuration for the webhook server is in place
	if err := utilswebhook.EnsureWebhookConfigurations(
		flags.serviceName, flags.webhookCertificate,
		flags.validatingWebhookConfiguration, "",
		flags.enableValidatingWebhook, false,
		mgr.GetAPIReader(), mgr.GetClient()); err != nil {
		entryLog.Error(err, "unable to ensure webhook configurations")
		os.Exit(1)
	}

	entryLog.Info("registering webhooks to the webhook server")
	if flags.enableValidatingWebhook {
		entryLog.Info("setting up genericValidator")
		// hookServer.Register(utilswebhook.ValidatingPath, &webhook.Admission{Handler: &namespaceValidator{Client: mgr.GetClient()}})
		hookServer.Register(utilswebhook.ValidatingPath, &webhook.Admission{
			Handler: webhooks.NewGenericValidator(mgr.GetClient(), log.WithName("genericValidator"), cfg),
		})
	}

	entryLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "unable to run manager")
		os.Exit(1)
	}

}
