package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"
	"sync"
)

type Rule struct {
	Field string      `yaml:"field"`
	Type  string      `yaml:"type"`
	Op    string      `yaml:"op"`
	Value interface{} `yaml:"value"`
}

type ForKindRules struct {
	ApiVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind"`
	Rules      []Rule `yaml:"rules"`
}

type Config struct {
	sync.RWMutex
	cache         map[string]Rule // this map is using the Kind (and api version? (gvk)?) as key
	ForKindsRules []ForKindRules  `yaml:"forKindsRules,omitempty"`
	AdminGroups   []string        `yaml:"adminGroups,omitempty"`
}

func NewConfig() *Config { return &Config{} }

func (cfg *Config) ParseYaml(data []byte) error {
	cfg.Lock()
	defer cfg.Unlock()

	err := yaml.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("Error parsing yaml: %v", err)
	}

	// set default of AdminGroups is not defined
	if len(cfg.AdminGroups) == 0 {
		cfg.AdminGroups = []string{"system:masters"}
	}

	// build cache
	cfg.cache = make(map[string][]Rule)
	for _, k := range cfg.ForKindsRules {
		key := k.Kind
		if _, found := cfg.cache[key]; found {
			cfg.cache[key] = append(cfg.cache[key], k.Rules...)
		} else {
			cfg.cache[key] = append([]Rule{}, k.Rules...)
		}
	}
	return nil
}

func (cfg *Config) GetRulesForKind(kind string) []Rule {
	cfg.Lock()
	defer cfg.Unlock()
	rulesForKind := []Rule{}
	rules, found := cfg.cache[kind]
	if found {
		for _, rule := range rules {
			rulesForKind = append(rulesForKind, rule)
		}
	}
	return rulesForKind
}

func (cfg *Config) GetAdminGroups() []string {
	cfg.Lock()
	defer cfg.Unlock()
	groups := []string{}
	for _, group := range cfg.AdminGroups {
		groups = append(groups, group)
	}
	return groups
}
