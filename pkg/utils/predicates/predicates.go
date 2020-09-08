package predicates

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func GetConfigPredicates(cmNamed types.NamespacedName) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			name := e.Meta.GetName()
			namespace := e.Meta.GetNamespace()
			return (namespace == cmNamed.Namespace &&
				name == cmNamed.Name)
		},

		UpdateFunc: func(e event.UpdateEvent) bool {
			name := e.MetaNew.GetName()
			namespace := e.MetaNew.GetNamespace()
			return (namespace == cmNamed.Namespace &&
				name == cmNamed.Name)
		},
		DeleteFunc: func(e event.DeleteEvent) bool { return false },
	}
}

func GetCertificateSecretPredicates(secNamed types.NamespacedName) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			name := e.Meta.GetName()
			namespace := e.Meta.GetNamespace()
			return namespace == secNamed.Namespace && name == secNamed.Name
		},

		UpdateFunc: func(e event.UpdateEvent) bool {
			name := e.MetaNew.GetName()
			namespace := e.MetaNew.GetNamespace()
			return namespace == secNamed.Namespace && name == secNamed.Name
		},
		DeleteFunc: func(e event.DeleteEvent) bool { return false },
	}
}
