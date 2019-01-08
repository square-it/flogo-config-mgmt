package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"strings"
)

var (
	kv     *api.KV
	logger = log.ChildLogger(log.RootLogger(), "consul-resolver")
)

func init() {
	logger := log.RootLogger()

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		logger.Error("Unable to initialize Consul property resolver for configuration management.")
	}

	kv = client.KV()

	logger.Debug("Registering Consul resolver")

	err = property.RegisterExternalResolver("env", &SimpleConsulKVValueResolver{})

	if err != nil {
		logger.Error("Unable to register Consul property resolver for configuration management.")
	}
}

// Resolve property value from a Consul KV Store
type SimpleConsulKVValueResolver struct {
}

func (resolver *SimpleConsulKVValueResolver) LookupValue(key string) (interface{}, bool) {
	key = strings.Replace(key, ".", "/", -1)

	consul_key := "flogo/" + engine.GetAppName() + "/" + key

	pair, _, err := kv.Get(consul_key, nil)

	var value interface{}

	if err != nil || pair == nil {
		logger.Warnf("Property '%s' is not found in Consul.", key)

		value = nil // will use default value
	} else {
		value = string(pair.Value)
	}

	return value, true
}
