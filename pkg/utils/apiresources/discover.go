package apiresources

import (
	ar "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"strings"
)

type GroupVersionResourceScope struct {
	Group    string
	Version  string
	Resource string
	Scope    *ar.ScopeType
}

type SupportedAPIResourcesMap = map[string]*GroupVersionResourceScope

func SupportedMap(config *rest.Config) (SupportedAPIResourcesMap, error) {
	c, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	rsrcLists, err := c.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	kindGVMap := make(SupportedAPIResourcesMap)
	for _, rsrcList := range rsrcLists {
		//fmt.Printf("GV: %v\n", rsrcList.GroupVersion)
		var g, v string
		gv := strings.Split(rsrcList.GroupVersion, "/")
		if len(gv) == 1 {
			g = ""
			v = gv[0]
		} else {
			g = gv[0]
			v = gv[1]
		}
		for _, r := range rsrcList.APIResources {
			if len(r.Group) > 0 {
				g = r.Group
			}
			if len(r.Version) > 0 {
				v = r.Version
			}
			scope := ar.ClusterScope
			if r.Namespaced {
				scope = ar.NamespacedScope
			}
			kindGVMap[r.Kind] = &GroupVersionResourceScope{g, v, r.Name, &scope}
		}
	}
	return kindGVMap, nil
}
