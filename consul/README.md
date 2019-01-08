# Centralized configuration management for Flogo with a Consul key/value store 

Before proceeding, read the
[introduction of properties resolution in Flogo](../README.md#about-properties-resolution-in-flogo).

## Requirements

* [Go language](https://golang.org) 1.11+ (with modules support enabled)
* [Flogo CLI](https://github.com/project-flogo/cli)
* [Docker](https://www.docker.com/get-started), for testing purpose only
* [jq](https://stedolan.github.io/jq/), to simplify the creation of the project only

## Usage

Let's demonstrate the resolution of properties with this Consul-based contribution, in collaboration (or not) with
built-in *resolvers*. 

### Create a Flogo project

1. create a simple Flogo project and enter in its directory
```
flogo create simple-config
cd simple-config
```

2. install this contribution
```
flogo install github.com/square-it/flogo-config-mgmt/consul
```

3. install the builtin resolvers
```
flogo install github.com/project-flogo/core/app/propertyresolver
```

4. add a property ```log.message``` and use it in the log activity
```
cat flogo.json | \
jq '. + {"properties": [{"name": "log.message", "type": "string", "value": "Default message"}]}' | \
jq '.resources[].data.tasks[].activity |= . + {"mappings": {"input": [{"type": "assign", "value": "$property[log.message]", "mapTo": "message"}]}}' | \
jq '.resources[].data.tasks[].activity.input |= del(.message)' \
> flogo.json.tmp && mv flogo.json.tmp flogo.json
```

5. build the application
```
flogo build
```

### Prepare a Consul key-value store

1. start a single-host development-mode Consul agent with Docker (for development purpose only)
```
docker run -d --name=dev-consul -e CONSUL_BIND_INTERFACE=eth0 -p 8500:8500 consul
```

2. add a value for the property ```log.message``` in Consul
```
docker exec -t dev-consul consul kv put flogo/simple-config/log/message "Consul message"
```

* The key is prefixed by ```flogo``` and by the application name (```simple-config```).

* Notice that the property ```log.message``` created above in *flogo.json* is using the
[standard way of grouping properties](https://tibcosoftware.github.io/flogo/development/flows/property-bag/#grouping-of-properties)
with a dot separator. All dots in properties names are replaced by forward slashes before being resolved in Consul,
leveraging the hierarchical paradigm of the Consul key/value store.

### Run & test 

#### without Consul

1. disable all property resolvers
```
export FLOGO_APP_PROP_RESOLVERS=disabled         # ensure none of the registered property resolvers is enabled
```

2. run the application
```
./bin/simple-config
```

3. in another terminal
```
curl http://127.0.0.1:8888/test/hello
```

You should see a log like:
```
INFO    [flogo.activity.log] -  Default message
```

#### with Consul 

1. set following environment variables
```
export CONSUL_HTTP_ADDR=127.0.0.1:8500           # optional as it is the default value

export FLOGO_APP_PROP_RESOLVERS=consul           # enable only the Consul property resolver

# or simply unset FLOGO_APP_PROP_RESOLVERS to use default behaviour
unset FLOGO_APP_PROP_RESOLVERS                   # enable all resolvers & the Consul one will have the highest priority
```

2. run the application
```
./bin/simple-config
```

3. in another terminal
```
curl http://127.0.0.1:8888/test/hello
```

You should see a log like:
```
INFO    [flogo.activity.log] -  Consul message
```

#### with Consul, overriding with a JSON file

1. create a JSON file with an overridden value for ```log.message```
```
echo '{"log.message": "JSON file message"}' | jq '.' > config.json
```

2. set following environment variables
```
export CONSUL_HTTP_ADDR=127.0.0.1:8500           # optional as it is the default value
export FLOGO_APP_PROP_RESOLVERS=json,consul      # enable the JSON files resolver & the Consul property resolver 
export FLOGO_APP_PROPS_JSON=config.json          # override with a property value in a JSON file
```

3. run the application
```
./bin/simple-config
```

4. in another terminal
```
curl http://127.0.0.1:8888/test/hello
```

You should see a log like:
```
INFO    [flogo.activity.log] -  JSON file message
```

The property value in the JSON config file has overridden the Consul value.

The Consul API is not even called since the value of the property was resolved successfully by a resolver with a higher
priority.

However the Consul resolver is still configured and active for potential other properties to resolve.

#### with Consul, overriding with an environment variable

1. set following environment variables
```
export CONSUL_HTTP_ADDR=127.0.0.1:8500           # optional as it is the default value
export FLOGO_APP_PROP_RESOLVERS=env,json,consul  # enable the environment variable resolver & the other resolvers
export FLOGO_APP_PROPS_JSON=config.json          # override with a property value in a JSON file 
export LOG_MESSAGE="Env var message"             # override with a (canonical) environment variable
```

2. run the application
```
./bin/simple-config
```

3. in another terminal
```
curl http://127.0.0.1:8888/test/hello
```

You should see a log like:
```
INFO    [flogo.activity.log] -  Env var message
```

The environment variable has overridden the Consul value **and** the JSON file value.

The Consul API is not even called since the value of the property was resolved successfully by a resolver with a higher
priority.

However the Consul resolver **and** the JSON file resolver are still configured and active for potential other
properties to resolve.

### Teardown

To cleanup the environment after the tests:

```
docker rm -f dev-consul

unset CONSUL_HTTP_ADDR
unset FLOGO_APP_PROP_RESOLVERS
unset FLOGO_APP_PROPS_JSON
unset LOG_MESSAGE
```

## Configuration

The configuration is set by a very limited set of environment variables. All of them have default values.

### Flogo level

| Name                       | Required                                              | Default value     | Example value                                              |
|----------------------------|-------------------------------------------------------|-------------------|------------------------------------------------------------|
| FLOGO_APP_PROP_RESOLVERS   | only when two or more custom resolvers are registered | "consul,env,json" | "consul,env", "consul", "json,env", "consul,etcd,env", ... |
| FLOGO_APP_PROPS_JSON       | for "json" resolver only                              | ø                 | "default.json", "common.json,prod.json"                    |

> Default value of ```FLOGO_APP_PROP_RESOLVERS``` is computed when only a single custom *resolver* is registered.
Otherwise, it must be set explicitly.

### Consul level

| Name             | Required | Default value    | Example value           |
|------------------|----------|------------------|-------------------------|
| CONSUL_HTTP_ADDR | no       | 127.0.0.1:8500   | consul.company.com:8500 |
