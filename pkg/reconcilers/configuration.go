package reconcilers

import (
	"context"
	"fmt"

	"crypto/md5"
	"sort"
	"strings"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/safanaj/k8s-generic-validator/pkg/config"
	"github.com/safanaj/k8s-generic-validator/pkg/utils/apiresources"
)

const ConfigurationConfigMapKey string = "config.yml"

type configurationReconciler struct {
	client.Client
	clientCfg *rest.Config
	log       logr.Logger
	cfg       *config.Config

	kindsChecksum [md5.Size]byte
}

func NewConfigurationReconciler(log logr.Logger, cfg *config.Config) reconcile.Reconciler {
	return &configurationReconciler{log: log, cfg: cfg}
}

func (r *configurationReconciler) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}

func (r *configurationReconciler) InjectConfig(cliCfg *rest.Config) error {
	r.clientCfg = cliCfg
	return nil
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &configurationReconciler{}

func (r *configurationReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// set up a convenient log object so we don't have to type request over and over again
	log := r.log.WithValues("request", request)

	// Fetch the ConfigMap from the cache
	cm := &corev1.ConfigMap{}
	err := r.Get(context.TODO(), request.NamespacedName, cm)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not find ConfigMap")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch ConfigMap: %+v", err)
	}

	log = log.WithValues("configmap", request.NamespacedName)
	log.Info("Reconciling Configuration")

	data, found := cm.Data[ConfigurationConfigMapKey]
	if !found {
		return reconcile.Result{}, fmt.Errorf("ConfigMap is missing required key: %s", ConfigurationConfigMapKey)
	}

	if err := r.cfg.ParseYaml([]byte(data)); err != nil {
		return reconcile.Result{}, fmt.Errorf("ConfigMap is not well formatted: %+v", err)
	}

	// store kinds checksum and eventually update webhook configuration
	oldsum := r.kindsChecksum
	kinds := r.cfg.GetKinds()
	sort.Strings(kinds)
	kindsStr := strings.Join(kinds, "")
	r.kindsChecksum = md5.Sum([]byte(kindsStr))
	if oldsum[0] != 0 && oldsum != r.kindsChecksum {
		// TODO: notify/update the webhook configuration
		if supportedMap, err := apiresources.SupportedMap(r.clientCfg); err == nil {
			// TODO: refresh rules in Webhook configuration (utilswebhook.EnsureWebhookConfigurations() ??)
			_ = oldsum
			_ = supportedMap
		} else {
			// getting error, enable to refresh webhook configuration
			return reconcile.Result{}, fmt.Errorf("Unable to fetch supported api-resources: %+v", err)
		}
	}

	return reconcile.Result{}, nil
}
