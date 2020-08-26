package main

import (
	goflag "flag"
	flag "github.com/spf13/pflag"
)

type Flags struct {
	version bool

	webhookCAIssuer                string
	webhookCertificate             string
	serviceName                    string
	validatingWebhookConfiguration string
	enableValidatingWebhook        bool

	// unused
	mutatingWebhookConfiguration string
	enableMutatingWebhook        bool
}

func parseFlags() *Flags {
	flags := &Flags{}
	flag.BoolVar(&flags.version, "version", false, "Print version and exit")
	flag.StringVar(&flags.webhookCAIssuer, "webhook-ca-issuer", "kube-system/central-root-ca-for-webhooks", "Namespaced cert-manager.io issuer to look for")
	flag.StringVar(&flags.webhookCertificate, "webhook-certificate", "", "Namespaced cert-manager.io certificate to look for (or create) certificate/key pair")
	flag.StringVar(&flags.serviceName, "service-name", "kube-system/k8s-generic-validator", "Namespaced Service Name")
	flag.StringVar(&flags.validatingWebhookConfiguration, "validating-webhook-configuration", "", "ValidatingWebhookConfiguration to create")
	flag.BoolVar(&flags.enableValidatingWebhook, "enable-validating-webhook", true, "Execute the validating webhook service")

	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
	return flags
}
