package webhook

import (
	"context"
	"strings"

	ar "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	port int32 = 443

	// needs refactoring to build this at runtime based on configuration
	ValidatingPath string = "/validate"
	MutatingPath   string = "/mutate"
)

// needs refactoring to build this at runtime based on configuration
func getRules() []ar.RuleWithOperations {
	scope := ar.NamespacedScope
	return []ar.RuleWithOperations{
		{
			Operations: []ar.OperationType{
				ar.Create, ar.Update,
			},
			Rule: ar.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"services"},
				Scope:       &scope,
			},
		},
	}
}

func getServiceNamespacedName(name string) types.NamespacedName {
	parts := strings.Split(name, "/")
	return types.NamespacedName{parts[0], parts[1]}
}

func EnsureWebhookConfigurations(
	serviceName, webhookCertificate, validating, mutating string,
	enableValidating, enableMutating bool,
	r client.Reader, c client.Client) error {
	svcNamed := getServiceNamespacedName(serviceName)
	vPath := ValidatingPath
	mPath := MutatingPath
	port_ := port
	sideEffects := ar.SideEffectClassNone
	matchPolicy := ar.Equivalent

	if enableValidating && len(validating) > 0 {
		whc := &ar.ValidatingWebhookConfiguration{}
		if err := r.Get(context.TODO(), types.NamespacedName{"", validating}, whc); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}

			// needs to be created
			whc.ObjectMeta.Name = validating
			whc.ObjectMeta.Annotations = map[string]string{
				"cert-manager.io/inject-ca-from": webhookCertificate,
			}
			whc.Webhooks = []ar.ValidatingWebhook{
				{
					Name: strings.Join([]string{"validate", validating, "aureacentral", "com"}, "."),
					ClientConfig: ar.WebhookClientConfig{
						Service: &ar.ServiceReference{
							Namespace: svcNamed.Namespace,
							Name:      svcNamed.Name,
							Path:      &vPath,
							Port:      &port_,
						},
					},
					Rules:       getRules(),
					MatchPolicy: &matchPolicy,
					SideEffects: &sideEffects,
				},
			}

			if err := c.Create(context.TODO(), whc); err != nil {
				return err
			}
		}
		// already exists, keep it untouched

		whc.ObjectMeta.Annotations = map[string]string{
			"cert-manager.io/inject-ca-from": webhookCertificate,
		}
		whc.Webhooks = []ar.ValidatingWebhook{
			{
				Name: strings.Join([]string{"validate", validating, "aureacentral", "com"}, "."),
				ClientConfig: ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: svcNamed.Namespace,
						Name:      svcNamed.Name,
						Path:      &vPath,
						Port:      &port_,
					},
				},
				Rules:       getRules(),
				MatchPolicy: &matchPolicy,
				SideEffects: &sideEffects,
			},
		}

		if err := c.Update(context.TODO(), whc); err != nil {
			return err
		}
	}

	if enableMutating && len(mutating) > 0 {
		whc := &ar.MutatingWebhookConfiguration{}
		if err := r.Get(context.TODO(), types.NamespacedName{"", mutating}, whc); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}

			// needs to be created
			whc.ObjectMeta.Name = mutating
			whc.ObjectMeta.Annotations = map[string]string{
				"cert-manager.io/inject-ca-from": webhookCertificate,
			}
			whc.Webhooks = []ar.MutatingWebhook{
				{
					Name: strings.Join([]string{"mutate", mutating, "aureacentral", "com"}, "."),
					ClientConfig: ar.WebhookClientConfig{
						Service: &ar.ServiceReference{
							Namespace: svcNamed.Namespace,
							Name:      svcNamed.Name,
							Path:      &mPath,
							Port:      &port_,
						},
					},
					Rules:       getRules(),
					MatchPolicy: &matchPolicy,
					SideEffects: &sideEffects,
				},
			}

			if err := c.Create(context.TODO(), whc); err != nil {
				return err
			}
		}

		whc.ObjectMeta.Annotations = map[string]string{
			"cert-manager.io/inject-ca-from": webhookCertificate,
		}
		whc.Webhooks = []ar.MutatingWebhook{
			{
				Name: strings.Join([]string{"mutate", mutating, "aureacentral", "com"}, "."),
				ClientConfig: ar.WebhookClientConfig{
					Service: &ar.ServiceReference{
						Namespace: svcNamed.Namespace,
						Name:      svcNamed.Name,
						Path:      &mPath,
						Port:      &port_,
					},
				},
				Rules:       getRules(),
				MatchPolicy: &matchPolicy,
				SideEffects: &sideEffects,
			},
		}

		if err := c.Update(context.TODO(), whc); err != nil {
			return err
		}
	}

	return nil
}
