package consul

import (
	"errors"
	"fmt"
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

// Resolve property value from environment variable
type SimpleConsulKVValueResolver struct {
}

func (resolver *SimpleConsulKVValueResolver) ResolveValue(toResolve string) (interface{}, error) {
	key := "flogo/" + app.GetName() + "/" + toResolve

	pair, _, err := kv.Get(key, nil)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Variable %s is not found in Consul", toResolve))
	}

	value := string(pair.Value)

	return value, nil
}
