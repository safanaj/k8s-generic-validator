package configuration

import (
	"context"
	"fmt"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/safanaj/k8s-generic-validator/pkg/config"
)

var firstConfigLoad sync.Once
var configuration types.NamespacedName

func GetConfigurationNamespacedName() types.NamespacedName { return configuration }

func EnsureFirstConfigurationLoad(namespacedConfigMap string, c client.Reader, cfg *config.Config, configurationMapKey string) error {
	var err error
	onceDo := func() {
		parts := strings.Split(namespacedConfigMap, "/")
		configuration = types.NamespacedName{parts[0], parts[1]}
		cm := &corev1.ConfigMap{}
		if err = c.Get(context.TODO(), configuration, cm); err != nil {
			return
		}

		data, found := cm.Data[configurationMapKey]
		if !found {
			err = fmt.Errorf("ConfigMap %s is missing required key: %s", configuration.String(),
				configurationMapKey)
			return
		}

		err = cfg.ParseYaml([]byte(data))
	}
	firstConfigLoad.Do(onceDo)
	return err
}
