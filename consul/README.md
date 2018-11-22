# Centralized configuration management for Flogo with a Consul KV Store 

## Usage

### Prepare a Consul KV Store

1. start a single-host development-mode Consul agent
```
docker run -d --name=dev-consul -e CONSUL_BIND_INTERFACE=eth0 --mount type=bind,source="$(pwd)"/config.json,target=/config.json -p 8500:8500 consul
```

2. add a property in Consul 
```
docker exec -t dev-consul consul kv put flogo/simple-config/message "Hello world"
```

### Install this contribution

```
flogo install github.com/square-it/flogo-config-mgmt/consul
```

### Set Flogo environment variables

```
export FLOGO_APP_PROPS_OVERRIDE=consul
export FLOGO_APP_PROPS_VALUE_RESOLVER=consul
export CONSUL_HTTP_ADDR=127.0.0.1:8500
```
