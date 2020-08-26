package reconcilers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	utilstls "github.com/safanaj/k8s-generic-validator/pkg/utils/tls"
)

type certSecReconciler struct {
	client.Client
	log                        logr.Logger
	certDir, certName, keyName string
}

func NewCertificateSecretReconciler(log logr.Logger, certDir, certName, keyName string) reconcile.Reconciler {
	return &certSecReconciler{log: log, certDir: certDir, certName: certName, keyName: keyName}
}

func (r *certSecReconciler) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &certSecReconciler{}

func (r *certSecReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("request", request)

	// Fetch the ConfigMap from the cache
	sec := &corev1.Secret{}
	err := r.Get(context.TODO(), request.NamespacedName, sec)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not find Secret")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch Secret: %+v", err)
	}

	err = utilstls.WriteSecretForCertificateByCertManager(r.log, sec, r.certDir, r.certName, r.keyName)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not sync Secret to files: %+v", err)
	}
	return reconcile.Result{}, nil
}
