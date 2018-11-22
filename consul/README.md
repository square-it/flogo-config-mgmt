# Centralized configuration management for Flogo with a Consul KV Store 

## Usage

### Prepare a Consul KV Store

1. create a configuration file in Consul JSON format:
```
cat << EOF > ./config.json
[
        {
                "key": "flogo/simple-config/message",
                "flags": 0,
                "value": "SGVsbG8="
        }
]
EOF
```

2. start a single-host development-mode Consul agent
```
docker run -d --name=dev-consul -e CONSUL_BIND_INTERFACE=eth0 --mount type=bind,source="$(pwd)"/config.json,target=/config.json -p 8500:8500 consul
```

3. import the configuration in Consul 
```
docker exec -t dev-consul consul kv import @config.json
```

### Install this contribution

```
flogo install github.com/square-it/flogo-config-mgmt/consul
```

### Set Flogo environment variables

```
export FLOGO_APP_PROPS_OVERRIDE=http://127.0.0.1/v1/kv
export FLOGO_APP_PROPS_VALUE_RESOLVER=consul
```