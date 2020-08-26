package webhooks

import (
	"context"
	"fmt"
	"net/http"

	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/go-logr/logr"
)

// validates entry of namespaces
type serviceValidator struct {
	Client  client.Client
	decoder *admission.Decoder
	log     logr.Logger
}

func NewServiceValidator(c client.Client, log logr.Logger) admission.Handler {
	return &serviceValidator{Client: c, log: log}
}

var _ admission.Handler = &serviceValidator{}

// namespaceValidator admits a pod iff a specific annotation exists.
func (v *serviceValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	svc := &corev1.Service{}

	v.log.Info("Handle", "req", req)

	err := v.decoder.Decode(req, svc)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	v.log.Info("Handle req is ok", "req", req, "svc", svc, "userinfo", req.UserInfo)
	// check user info
	if isClusterAdmin(req.UserInfo) {
		v.log.Info("Handle Allow cluster admin")
		return admission.Allowed("")
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer || svc.Spec.Type == corev1.ServiceTypeNodePort {
		v.log.Info("serviceValidator.Handle Deny")
		denyMsg := fmt.Sprintf(
			"Only service type ClusterIP or ExternalName can be used, %s service type is denied", string(svc.Spec.Type))
		return admission.Denied(denyMsg)
	}

	v.log.Info("Handle Allow")
	return admission.Allowed("")
}

// namespaceValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *serviceValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

// is cluster admin is checking for the user is member of specific groups
var AdminGroups []string = []string{"system:masters"}

func isClusterAdmin(userInfo authv1.UserInfo) bool {
	userGroups := sets.NewString(userInfo.Groups...)
	return userGroups.HasAny(AdminGroups...)
}
