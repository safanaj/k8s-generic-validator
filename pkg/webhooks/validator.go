package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	_ "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/go-logr/logr"

	"github.com/safanaj/k8s-generic-validator/pkg/config"
)

// validates entry of namespaces
type genericValidator struct {
	Client  client.Client
	decoder *admission.Decoder
	log     logr.Logger
	cfg     *config.Config
}

func NewGenericValidator(c client.Client, log logr.Logger, cfg *config.Config) admission.Handler {
	return &genericValidator{Client: c, log: log, cfg: cfg}
}

var _ admission.Handler = &genericValidator{}

// genericValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *genericValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

// genericValidator implements admission.Handler
func (v *genericValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	u := &unstructured.Unstructured{}

	v.log.Info("Handle", "req", req)

	err := v.decoder.Decode(req, u)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	v.log.Info("Handle req is ok", "req", req, "obj", u.Object, "userinfo", req.UserInfo)
	// check user info
	if isClusterAdmin(req.UserInfo, v.cfg.GetAdminGroups()) {
		v.log.Info("Handle Allow cluster admin")
		return admission.Allowed("")
	}

	if rules := v.cfg.GetRulesForKind(u.GetKind()); len(rules) > 0 {
		for _, rule := range rules {
			if ok, err := v.verify(u.Object, rule); !ok || err != nil {
				var denyMsg string
				if err != nil {
					denyMsg = fmt.Sprintf("The error %v occurred verifing the rule: %v", err, rule)
				} else {
					denyMsg = fmt.Sprintf("Rule: %v violated", rule)
				}
				return admission.Denied(denyMsg)
			}
		}
	}

	v.log.Info("Handle Allow")
	return admission.Allowed("")
}

// is cluster admin is checking for the user is member of specific groups
func isClusterAdmin(userInfo authv1.UserInfo, adminGroups []string) bool {
	userGroups := sets.NewString(userInfo.Groups...)
	return userGroups.HasAny(adminGroups...)
}

// Logic for validation is implemented in verify method
func (v *genericValidator) verify(obj map[string]interface{}, rule config.Rule) (bool, error) {
	fieldPathParts := strings.Split(rule.Field, ".")
	switch rule.Type {
	case ValueTypeString:
		{
			var checkValues []string
			checkValue, ok := rule.Value.(string)
			if !ok {
				if rule.Op != OperatorIn && rule.Op != OperatorNotIn {
					return false, fmt.Errorf(
						"Value (of type %T) in rule is not of type: %s with Operator %s",
						rule.Value, rule.Type, rule.Op)
				} else {
					checkValues, ok = rule.Value.([]string)
					if !ok {
						return false, fmt.Errorf(
							"Value (of type %T) in rule is not of type: []%s with Operator %s",
							rule.Value, rule.Type, rule.Op)
					}
				}
			}
			val, ok, err := unstructured.NestedString(obj, fieldPathParts...)
			if !ok {
				return false, fmt.Errorf(
					"Field not found at %s into %+v", rule.Field, obj)
			}
			if err != nil {
				return false, err
			}
			switch rule.Op {
			case OperatorIsNot:
				return val != checkValue, nil
			case OperatorIs:
				return val == checkValue, nil
			case OperatorIn:
				return sets.NewString(checkValues...).Has(val), nil
			case OperatorNotIn:
				return !sets.NewString(checkValues...).Has(val), nil
			}
		}
	case ValueTypeBool:
		{
			checkValue, ok := rule.Value.(bool)
			if !ok {
				return false, fmt.Errorf(
					"Value (of type %T) in rule is not of type: %s",
					rule.Value, rule.Type)
			}
			val, ok, err := unstructured.NestedBool(obj, fieldPathParts...)
			if !ok {
				return false, fmt.Errorf(
					"Field not found at %s into %+v", rule.Field, obj)
			}
			if err != nil {
				return false, err
			}
			switch rule.Op {
			case OperatorIsNot:
				return val != checkValue, nil
			case OperatorIs:
				return val == checkValue, nil
			}
		}
	case ValueTypeInt, ValueTypeInt64:
		{
			checkIntValue, ok := rule.Value.(int)
			if !ok {
				return false, fmt.Errorf(
					"Value (of type %T) in rule is not of type: %s",
					rule.Value, rule.Type)
			}
			checkValue := int64(checkIntValue)
			val, ok, err := unstructured.NestedInt64(obj, fieldPathParts...)
			if !ok {
				return false, fmt.Errorf(
					"Field not found at %s into %+v", rule.Field, obj)
			}
			if err != nil {
				return false, err
			}
			switch rule.Op {
			case OperatorIsNot:
				return val != checkValue, nil
			case OperatorIs:
				return val == checkValue, nil
			case OperatorGreaterThan, OperatorMoreThan:
				return val > checkValue, nil
			case OperatorSmallerThan, OperatorLessThan:
				return val < checkValue, nil
			case OperatorEqualOrMoreThan, OperatorEqualOrGreaterThan:
				return val >= checkValue, nil
			case OperatorEqualOrLessThan, OperatorEqualOrSmallerThan:
				return val <= checkValue, nil
			}
		}
	case ValueTypeFloat, ValueTypeFloat64:
		{
			checkValue, ok := rule.Value.(float64)
			if !ok {
				return false, fmt.Errorf(
					"Value (of type %T) in rule is not of type: %s",
					rule.Value, rule.Type)
			}
			val, ok, err := unstructured.NestedFloat64(obj, fieldPathParts...)
			if !ok {
				return false, fmt.Errorf(
					"Field not found at %s into %+v", rule.Field, obj)
			}
			if err != nil {
				return false, err
			}
			switch rule.Op {
			case OperatorIsNot:
				return val != checkValue, nil
			case OperatorIs:
				return val == checkValue, nil
			case OperatorGreaterThan, OperatorMoreThan:
				return val > checkValue, nil
			case OperatorSmallerThan, OperatorLessThan:
				return val < checkValue, nil
			case OperatorEqualOrMoreThan, OperatorEqualOrGreaterThan:
				return val >= checkValue, nil
			case OperatorEqualOrLessThan, OperatorEqualOrSmallerThan:
				return val <= checkValue, nil
			}
		}
	}
	return false, fmt.Errorf("unknonw type in rule: %v", rule)
}
