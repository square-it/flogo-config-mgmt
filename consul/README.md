# Centralized configuration management for Flogo with a Consul KV Store 

## Usage

### Create a Flogo project

1.
```
flogo create simple-config
cd simple-config
```

2.
```
flogo install -v simple-consul-kv github.com/square-it/flogo-config-mgmt/consul
```

3.
```
cat << EOF >> ./src/simple-config/Gopkg.toml                                                    

[[override]]
  name = "github.com/TIBCOSoftware/flogo-lib"
  branch = "external-config-mgmt"
  source = "github.com/debovema/flogo-lib"

EOF
```

4.
```
flogo ensure
```

5.
```
cat flogo.json | jq '. + {"properties": [{"name": "message", "type": "string", "value": "Default message"}]}' | jq '.resources[].data.tasks[].activity |= . + {"mappings": {"input": [{"type": "assign", "value": "$property.message", "mapTo": "message"}]}}' | jq '.resources[].data.tasks[].activity.input |= del(.message)' > flogo.json.tmp && mv flogo.json.tmp flogo.json
```

6.
```
flogo build -e
```

### Prepare a Consul KV Store

1. start a single-host development-mode Consul agent
```
docker run -d --name=dev-consul -e CONSUL_BIND_INTERFACE=eth0 -p 8500:8500 consul
```

2. add a property in Consul 
```
docker exec -t dev-consul consul kv put flogo/simple-config/message "Consul message"
```

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
