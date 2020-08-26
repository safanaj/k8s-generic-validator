package tls

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"

	certmanagerapiv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"

	"github.com/safanaj/k8s-generic-validator/pkg/utils/predicates"
)

func EnsureWeNeedCertificateByCertManager(certDir, certName, keyName string) (bool, error) {
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		os.MkdirAll(certDir, os.FileMode(0700))
		return true, nil
	}

	cPath := filepath.Join(certDir, certName)
	kPath := filepath.Join(certDir, keyName)

	_, cErr := os.Stat(cPath)
	_, kErr := os.Stat(kPath)

	if os.IsNotExist(cErr) && os.IsNotExist(kErr) {
		return true, nil
	}

	if cErr == nil && kErr == nil {
		return false, nil
	}

	if (os.IsNotExist(cErr) && !os.IsNotExist(kErr)) || (!os.IsNotExist(cErr) && os.IsNotExist(kErr)) {
		return false, fmt.Errorf("Inconsistent TLS setup")
	}

	return false, fmt.Errorf("Inconsistent TLS setup: Unknown Errors: cert: %v - key: %v", cErr, kErr)
}

func SetupScheme(scheme *runtime.Scheme) {
	certmanagerapiv1alpha2.AddToScheme(scheme)
}

func SetupCertificateByCertManager(
	log logr.Logger,
	r client.Reader,
	c client.Client,
	svcNamed, caIssuerNamed, certNamed, secretNamed types.NamespacedName,
	certDir, certName, keyName string,
) error {
	var err error
	caIssuer := &certmanagerapiv1alpha2.Issuer{}
	if err := r.Get(context.TODO(), caIssuerNamed, caIssuer); err != nil {
		return err
	}
	caIssuerRef := cmmeta.ObjectReference{
		Name:  caIssuer.ObjectMeta.Name,
		Kind:  caIssuer.GetObjectKind().GroupVersionKind().Kind,
		Group: caIssuer.GetObjectKind().GroupVersionKind().Group,
	}

	d, _ := time.ParseDuration("50000h")
	rd, _ := time.ParseDuration("50h")
	cert := &certmanagerapiv1alpha2.Certificate{}

	opFn := func(ctx context.Context, obj runtime.Object) error { return c.Update(ctx, obj) }

	if err := r.Get(context.TODO(), certNamed, cert); err != nil {
		if !errors.IsNotFound(err) {
			panic(err)
		}

		cert.ObjectMeta.Name = certNamed.Name
		cert.ObjectMeta.Namespace = certNamed.Namespace
		opFn = func(ctx context.Context, obj runtime.Object) error { return c.Create(ctx, obj) }

	}

	dnsShortName := strings.Join([]string{svcNamed.Name, svcNamed.Namespace, "svc"}, ".")
	dnsFqdn := strings.Join([]string{dnsShortName, "cluster", "local"}, ".")
	cert.Spec = certmanagerapiv1alpha2.CertificateSpec{
		CommonName:   dnsShortName,
		Organization: []string{},
		DNSNames:     []string{dnsShortName, dnsFqdn},
		Usages: []certmanagerapiv1alpha2.KeyUsage{
			certmanagerapiv1alpha2.UsageDigitalSignature,
			certmanagerapiv1alpha2.UsageKeyEncipherment,
			certmanagerapiv1alpha2.UsageKeyAgreement,
			certmanagerapiv1alpha2.UsageServerAuth,
			certmanagerapiv1alpha2.UsageClientAuth,
		},
		Duration:    &metav1.Duration{d},
		RenewBefore: &metav1.Duration{rd},
		SecretName:  secretNamed.Name,
		KeySize:     4096,
		IssuerRef:   caIssuerRef,
	}

	err = opFn(context.TODO(), cert)

	if err != nil {
		return err
	}

	until, _ := time.ParseDuration("10m")
	err = waitForSecret(r, secretNamed, until)
	log.Info("wait for secret done", "err", err)

	if err != nil {
		return err
	}

	sec := &corev1.Secret{}
	if err := r.Get(context.TODO(), secretNamed, sec); err != nil {
		return err
	}
	log.Info("Gonna write files", "certDir", certDir, "certName", certName, "keyName", keyName)
	return WriteSecretForCertificateByCertManager(log, sec, certDir, certName, keyName)
}

func waitForSecret(c client.Reader, secNamed types.NamespacedName, until time.Duration) error {
	backoff := 2
	startedAt := time.Now()
	sec := &corev1.Secret{}
	for {
		if err := c.Get(context.TODO(), secNamed, sec); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
			if time.Since(startedAt) > until {
				return fmt.Errorf("Timed out waiting for Secret %s", secNamed.String())
			}
			time.Sleep(time.Duration(backoff*10) * time.Second)
			backoff++
			continue
		}
		return nil
	}
}

func WriteSecretForCertificateByCertManager(log logr.Logger, sec *corev1.Secret, certDir, certName, keyName string) error {
	cPath := filepath.Join(certDir, certName)
	kPath := filepath.Join(certDir, keyName)

	var cFile, kFile *os.File
	var err error

	log.Info("Writing cert", "file", cPath)
	// write cert
	cFile, err = os.OpenFile(cPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		goto cleanupAndError
	}

	_, err = cFile.WriteAt([]byte(sec.Data["tls.crt"]), 0)
	if err != nil {
		goto cleanupAndError
	}

	if err = cFile.Close(); err != nil {
		goto cleanupAndError
	}

	log.Info("Wrote cert", "file", cPath)
	log.Info("Writing key", "file", kPath)
	// write key
	kFile, err = os.OpenFile(kPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		goto cleanupAndError
	}

	_, err = kFile.WriteAt([]byte(sec.Data["tls.key"]), 0)
	if err != nil {
		goto cleanupAndError
	}

	if err = kFile.Close(); err != nil {
		goto cleanupAndError
	}

	log.Info("Wrote key", "file", kPath)
	return nil

cleanupAndError:
	if cFile != nil {
		cFile.Close()
	}
	if kFile != nil {
		kFile.Close()
	}
	os.Remove(cPath)
	os.Remove(kPath)
	return err
}

func SetupWebhookTlsSecretControllersOrDie(secNamed types.NamespacedName, mgr manager.Manager, reconciler reconcile.Reconciler) error {
	return builder.
		ControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(predicates.GetCertificateSecretPredicates(secNamed)).
		Complete(reconciler)
}

// unused
func GetClientCAName() string { return "clientCA.crt" }
func WriteClientCACertificate(certDir, data string) error {
	dst := filepath.Join(certDir, GetClientCAName())
	dFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer dFile.Close()
	_, err = dFile.WriteAt([]byte(data), 0)
	if err != nil {
		return err
	}
	return nil
}
func CopyClientCACertIntoCertDir(clientCAPath, certDir string) error {
	src := clientCAPath
	dst := filepath.Join(certDir, GetClientCAName())

	sFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sFile.Close()

	dFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer dFile.Close()
	_, err = io.Copy(dFile, sFile)
	return err
}
