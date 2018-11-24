# Centralized configuration management for Flogo with a Consul KV Store 

> **WARNING**: This contribution is in an experimental state and uses a patched version of _TIBCOSoftware/flogo-lib_.

## Usage

### Create a Flogo project

1. create a simple Flogo project and enter in its directory
```
flogo create simple-config
cd simple-config
```

2. install this contribution
```
flogo install -v simple-consul-kv github.com/square-it/flogo-config-mgmt/consul
```

3. use patched version of TIBCOSoftware/flogo-lib
```
cat << EOF >> ./src/simple-config/Gopkg.toml                                                    

[[override]]
  name = "github.com/TIBCOSoftware/flogo-lib"
  branch = "external-config-mgmt"
  source = "github.com/square-it/flogo-lib"

EOF
```

4. ensure dependencies
```
flogo ensure
```

5. add a property ```message``` and use it in the log activity
```
cat flogo.json | jq '. + {"properties": [{"name": "message", "type": "string", "value": "Default message"}]}' | jq '.resources[].data.tasks[].activity |= . + {"mappings": {"input": [{"type": "assign", "value": "$property.message", "mapTo": "message"}]}}' | jq '.resources[].data.tasks[].activity.input |= del(.message)' > flogo.json.tmp && mv flogo.json.tmp flogo.json
```

6. build the application
```
flogo build -e
```

### Prepare a Consul KV Store

1. start a single-host development-mode Consul agent
```
docker run -d --name=dev-consul -e CONSUL_BIND_INTERFACE=eth0 -p 8500:8500 consul
```

2. add a value for the property ```message``` in Consul 
```
docker exec -t dev-consul consul kv put flogo/simple-config/message "Consul message"
```

The key is prefixed by ```flogo``` and by the application name (```simple-config```).

### Run & test 

#### without Consul

1. run the application
```
./bin/simple-config
```

2. in another terminal
```
curl http://127.0.0.1:9233/test
```

You should see a log like:
```
INFO   [activity-flogo-log] - Default message
```

#### with Consul 

1. set following environment variables
```
export CONSUL_HTTP_ADDR=127.0.0.1:8500
export FLOGO_APP_PROPS_OVERRIDE=consul
export FLOGO_APP_PROPS_VALUE_RESOLVER=consul
```

2. run the application
```
./bin/simple-config
```

You should see a log like:
```
INFO   [activity-flogo-log] - Consul message
```
