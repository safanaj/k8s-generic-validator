package reconcilers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/safanaj/k8s-generic-validator/pkg/config"
)

const ConfigurationConfigMapKey string = "config.yml"

type configurationReconciler struct {
	client.Client
	log logr.Logger
	cfg *config.Config
}

func NewConfigurationReconciler(log logr.Logger, cfg *config.Config) reconcile.Reconciler {
	return &configurationReconciler{log: log, cfg: cfg}
}

func (r *configurationReconciler) InjectClient(c client.Client) error {
	r.Client = c
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

	return reconcile.Result{}, nil
}
