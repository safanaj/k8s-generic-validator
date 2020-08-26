package predicates

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	// "github.com/cshivashankar/namespace-configurator/pkg/product"
	// "github.com/cshivashankar/namespace-configurator/pkg/utils/configuration"
)

// func GetProductPredicates(pc *product.Config) predicate.Funcs {
// 	return predicate.Funcs{
// 		CreateFunc: func(e event.CreateEvent) bool {
// 			labels := e.Meta.GetLabels()
// 			return pc.HasPredicates(labels, e.Meta.GetName())
// 		},

// 		UpdateFunc: func(e event.UpdateEvent) bool {
// 			labels := e.MetaOld.GetLabels()
// 			return (pc.HasPredicates(labels, e.MetaNew.GetName()) && (e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()))
// 		},
// 		DeleteFunc: func(e event.DeleteEvent) bool { return false },
// 	}
// }

// func GetConfigPredicates() predicate.Funcs {
// 	return predicate.Funcs{
// 		CreateFunc: func(e event.CreateEvent) bool {
// 			name := e.Meta.GetName()
// 			namespace := e.Meta.GetNamespace()
// 			return (namespace == configuration.GetConfigurationNamespacedName().Namespace &&
// 				name == configuration.GetConfigurationNamespacedName().Name)
// 		},

// 		UpdateFunc: func(e event.UpdateEvent) bool {
// 			name := e.MetaNew.GetName()
// 			namespace := e.MetaNew.GetNamespace()
// 			return (namespace == configuration.GetConfigurationNamespacedName().Namespace &&
// 				name == configuration.GetConfigurationNamespacedName().Name)
// 		},
// 		DeleteFunc: func(e event.DeleteEvent) bool { return false },
// 	}
// }

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
