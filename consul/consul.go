package consul

import (
	"github.com/TIBCOSoftware/flogo-lib/app"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/hashicorp/consul/api"
)

var (
	kv  *api.KV
	log = logger.GetLogger("config-mgmt-consul-kv")
)

func init() {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Error("Unable to initialize Consul property resolver for configuration management.")
	}

	kv = client.KV()

	err = app.RegisterPropertyValueResolver("consul", &SimpleConsulKVValueResolver{})
	if err != nil {
		log.Error("Unable to register Consul property resolver for configuration management.")
	}
}

// Resolve property value from a Consul KV Store
type SimpleConsulKVValueResolver struct {
}

func (resolver *SimpleConsulKVValueResolver) ResolveValue(toResolve string) (interface{}, error) {
	key := "flogo/" + app.GetName() + "/" + toResolve

	pair, _, err := kv.Get(key, nil)

	var value interface{}

	if err != nil || pair == nil {
		log.Warnf("Property '%s' is not found in Consul.", toResolve)

		value = nil // will use default value
	} else {
		value = string(pair.Value)
	}

	return value, nil
}
