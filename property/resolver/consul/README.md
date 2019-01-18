# Centralized configuration management for Flogo with a Consul key/value store 

> **WARNING**: This contribution is in an experimental state and uses latest version of _project-flogo/core_.

## About properties resolution in Flogo

Properties are defined in the *flogo.json* configuration file in the *properties* array. For instance:
```json 
  "properties": [
    {
      "name": "log.message",
      "type": "string",
      "value": "Default message"
    }
  ]
```

Properties values can be overridden by different *resolvers*:
* **Environment variables**: this [built-in resolver](https://github.com/project-flogo/core/blob/master/app/propertyresolver/env.go)
is the simplest method when deploying the application in containers (Docker, Kubernetes, ...)
* **JSON files**: this [built-in resolver](https://github.com/project-flogo/core/blob/master/app/propertyresolver/json.go)
allows to load properties from one or several JSON files (common.json, dev.json/prod.json, ...) 
* **Others**: other *resolvers* can be provided by Flogo contributions (as this one for Consul).
These contributions are installed in the application using standard ``` flogo install``` command.

These *resolvers* can be enabled and chained using the ```FLOGO_APP_PROPS_RESOLVERS``` environment variable.
It means that defined resolvers will be used to resolve properties as follows:

* The priority order is following the order of declaration in ```FLOGO_APP_PROPS_RESOLVERS```.
* Each property uses the value from the first *resolver* succeeding (independently of other properties).
* If no *resolver* returns a value for a property, its default value from *flogo.json* is used.
* Resolution is performed at the start of the Flogo engine and properties are then cached during the lifetime of the
application.

## Requirements

* Go >= 1.11 (with [modules support enabled](https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support))
* Flogo >= v0.9.0-alpha4 (use [new CLI](https://github.com/project-flogo/cli))
* Docker, for testing purpose only
* jq, to simplify the creation of the project only

## Usage

Let's demonstrate the resolution of properties with this Consul-based contribution (and with built-in methods too). 

### Create a Flogo project

1. create a simple Flogo project and enter in its directory
```
flogo create simple-config
cd simple-config
```

2. install this contribution
```
flogo install github.com/square-it/flogo-config-mgmt/property/resolver/consul
``` 

3. add a property ```log.message``` and use it in the log activity
```
cat flogo.json | \
jq '. + {"properties": [{"name": "log.message", "type": "string", "value": "Default message"}]}' | \
jq '.resources[].data.tasks[].activity.input.message |= "=$property[log.message]"' \
> flogo.json.tmp && mv flogo.json.tmp flogo.json
```

4. build the application
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

1. run the application
```
./bin/simple-config
```

2. in another terminal
```
curl http://127.0.0.1:8888/test/test
```

You should see a log like:
```
INFO	[flogo.activity.log] -	Default message
```

#### with Consul 

1. set following environment variables
```
export CONSUL_HTTP_ADDR=127.0.0.1:8500          # optional as it is the default value
export FLOGO_APP_PROPS_RESOLVERS=consul          # enable the Consul property resolver
```

2. run the application
```
./bin/simple-config
```

3. in another terminal
```
curl http://127.0.0.1:8888/test/test
```

You should see a log like:
```
INFO	[flogo.activity.log] -	Consul message
```

#### with Consul, overriding with a profile file

1. create a JSON profile file with an overridden value for ```log.message```
```
echo '{"log.message": "Profile file message"}' | jq '.' > profile.json
```

2. add the ```github.com/project-flogo/core/app/propertyresolver``` package to ```imports.go``` file
(if it does not exist yet) and rebuild the application:
```
sed -i 's|^)$|        _ "github.com/project-flogo/core/app/propertyresolver"\n)|' ./src/imports.go
flogo build
```

3. set following environment variables
```
export CONSUL_HTTP_ADDR=127.0.0.1:8500          # optional as it is the default value
export FLOGO_APP_PROPS_RESOLVERS=json,consul     # enable the built-in JSON property resolver and Consul property resolver, in this order
export FLOGO_APP_PROPS_JSON=profile.json        # override with a property value in a JSON profile file
```

4. run the application
```
./bin/simple-config
```

5. in another terminal
```
curl http://127.0.0.1:8888/test/test
```

You should see a log like:
```
INFO	[flogo.activity.log] -	Profile file message
```

The property value in the profile file has overridden the Consul value.

The Consul API is not even called since the value of the property was resolved successfully by a resolver with a higher
priority.

However the Consul resolver is still configured and active for potential other properties to resolve.

#### with Consul, overriding with an environment variable

1. add the ```github.com/project-flogo/core/app/propertyresolver``` package to ```imports.go``` file
(if it does not exist yet) and rebuild the application:
```
sed -i 's|^)$|        _ "github.com/project-flogo/core/app/propertyresolver"\n)|' ./src/imports.go
flogo build
```

2. set following environment variables
```
export CONSUL_HTTP_ADDR=127.0.0.1:8500          # optional as it is the default value
export FLOGO_APP_PROPS_RESOLVERS=env,json,consul # enable the Consul external property resolver
export FLOGO_APP_PROPS_JSON=profile.json        # override with a property value in a JSON profile file
export LOG_MESSAGE="Env var message"            # override with a (canonical) environment variable
```

3. run the application
```
./bin/simple-config
```

4. in another terminal
```
curl http://127.0.0.1:8888/test/test
```

You should see a log like:
```
INFO	[flogo.activity.log] -	Env var message
```

The environment variable has overridden the Consul value **and** the profile file value.

The Consul API is not even called since the value of the property was resolved successfully by a resolver with a higher
priority.

However the Consul resolver **and** the profile file resolver are still configured and active for potential other
properties to resolve.

### Teardown

```
docker rm -f dev-consul

unset CONSUL_HTTP_ADDR
unset FLOGO_APP_PROPS_RESOLVERS
unset FLOGO_APP_PROPS_JSON
unset LOG_MESSAGE
```

## Configuration

The configuration of this Consul property resolver is set by putting "consul" in the comma-separated
```FLOGO_APP_PROPS_RESOLVERS``` environment variable and configuring the Consul backend with its own environment
variables.

### Flogo level

| Name                       | Default value | Required                      | Example value                                 |
|----------------------------|---------------|-------------------------------|-----------------------------------------------|
| FLOGO_APP_PROPS_RESOLVERS   | ø             | yes                           | "consul", "consul,env", "env,consul,json" ... |

### Consul level

| Name              | Default value   | Required | Example value           |
|-------------------|-----------------|----------|-------------------------|
| CONSUL_HTTP_ADDR  | 127.0.0.1:8500  | no       | consul.company.com:8500 |

All other [environment variables supported](https://godoc.org/github.com/hashicorp/consul/api#pkg-constants) supported
by the [Go Consul API client](https://github.com/hashicorp/consul/tree/master/api) are supported (for instance.
```CONSUL_HTTP_AUTH```, ```CONSUL_HTTP_SSL```, ...).
