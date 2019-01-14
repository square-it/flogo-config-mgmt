package consulresolver

import (
	"github.com/hashicorp/consul/api"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"strings"
)

var (
	logger = log.ChildLogger(log.RootLogger(), "consul-resolver")
)

func init() {
	logger := log.RootLogger()

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		logger.Error("Unable to initialize Consul property resolver for configuration management.")
		return
	}

	simpleConsulKVValueResolver := SimpleConsulKVValueResolver{consulKVClient: client.KV()}

	logger.Debug("Registering Consul resolver")
	err = property.RegisterExternalResolver("consul", &simpleConsulKVValueResolver)
	if err != nil {
		logger.Error("Unable to register Consul property resolver for configuration management.")
	}
}

// Resolve property value from a Consul KV Store
type SimpleConsulKVValueResolver struct {
	consulKVClient *api.KV
}

func (resolver *SimpleConsulKVValueResolver) LookupValue(key string) (interface{}, bool) {
	key = strings.Replace(key, ".", "/", -1)

	consulKey := "flogo/" + engine.GetAppName() + "/" + key

	pair, _, err := resolver.consulKVClient.Get(consulKey, nil)

	var value interface{}

	if err != nil || pair == nil {
		logger.Warnf("Property '%s' is not found in Consul.", key)

		return nil, false // will use value of next resolver, or fail if no more resolver in the list
	} else {
		value = string(pair.Value)
	}

	return value, true
}
