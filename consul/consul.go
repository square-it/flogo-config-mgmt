package consul

import (
	"github.com/TIBCOSoftware/flogo-lib/app"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/hashicorp/consul/api"
	"strings"
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

func (resolver *SimpleConsulKVValueResolver) ResolveValue(key string) (interface{}, error) {
	key = strings.Replace(key, ".", "/", -1)
	consul_key := "flogo/" + app.GetName() + "/" + key

	pair, _, err := kv.Get(consul_key, nil)

	var value interface{}

	if err != nil || pair == nil {
		log.Warnf("Property '%s' is not found in Consul.", key)

		value = nil // will use default value
	} else {
		value = string(pair.Value)
	}

	return value, nil
}
