package consul

import (
	"github.com/TIBCOSoftware/flogo-lib/app"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/hashicorp/consul/api"
)

var kv *api.KV

func init() {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		logger.Error("Unable to initialize Consul resolver for configuration management")
	}

	kv = client.KV()

	app.RegisterPropertyValueResolver("consul", &SimpleConsulKVValueResolver{})
}

// Resolve property value from a Consul KV Store
type SimpleConsulKVValueResolver struct {
}

func (resolver *SimpleConsulKVValueResolver) ResolveValue(toResolve string) (interface{}, error) {
	key := "flogo/" + app.GetName() + "/" + toResolve

	pair, _, err := kv.Get(key, nil)

	var value interface{}

	if err != nil || pair == nil {
		logger.Warnf("Variable '%s' is not found in Consul", toResolve)

		value = "$" + toResolve
	} else {
		value = string(pair.Value)
	}

	return value, nil
}
