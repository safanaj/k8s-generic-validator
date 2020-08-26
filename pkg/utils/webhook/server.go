package webhook

// import (
// 	"context"
// 	"sigs.k8s.io/controller-runtime/pkg/manager"
// 	"sigs.k8s.io/controller-runtime/pkg/webhook"

// 	corev1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/types"

// 	utilstls "github.com/cshivashankar/namespace-configurator/pkg/utils/tls"
// )

// func NewWebhookServer(mgr manager.Manager, certDir, clientCAPath string) *webhook.Server {
// 	reader := mgr.GetAPIReader()
// 	cm := &corev1.ConfigMap{}

// 	if err := reader.Get(context.TODO(), types.NamespacedName{"kube-system", "extension-apiserver-authentication"}, cm); err != nil {
// 		panic(err)
// 	}

// 	if err := utilstls.WriteClientCACertificate(certDir, cm.Data["client-ca-file"]); err != nil {
// 		panic(err)
// 	}

// 	// if err := utilstls.CopyClientCACertIntoCertDir(); err != nil {
// 	// 	panic(err)
// 	// }

// 	ws := &webhook.Server{
// 		CertDir:      certDir,
// 		ClientCAName: utilstls.GetClientCAName(),
// 	}

// 	if err := mgr.Add(ws); err != nil {
// 		panic("unable to add webhookServer to the controller manager")
// 	}
// 	return ws
// }
