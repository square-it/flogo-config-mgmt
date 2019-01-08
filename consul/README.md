# Centralized configuration management for Flogo with a Consul key/value store 

> **WARNING**: This contribution is in an experimental state and uses a patched version of _project-flogo/core_.

## About properties resolution in Flogo

> **WARNING**: This description is true *only in this experimentation* until
[core/pull#9](https://github.com/project-flogo/core/pull/9) is merged.

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

Properties values can be resolved by different methods:
* **Environment variables**: simplest method when deploying the application in containers (Docker, Kubernetes, ...)
* **Profile files**: best method to define different *profiles* of properties such as *dev*, *integration*, *prod*
* **External resolver**: can be any system supported by a Flogo contribution (as this one for Consul).
The contribution must be installed in the application using standard ``` flogo install``` command
* **Default**: use the value specified in the *flogo.json* configuration file of the application

The defined order for properties resolution is as follows:

> Environment variable **>** profile file (JSON) **>** external (Consul) **>** default value

* This priority order is statically defined.
* Each method is called only if it is enabled (see [Configuration](#configuration)).
* Each property uses the value from the first method succeeding.
* Resolution is performed at the start of the Flogo engine and properties are then cached during the lifetime of the
application.

## Requirements

* Go >= 1.11 (with modules support enabled)
* Flogo, at least v0.9.0-alpha3
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
flogo install github.com/square-it/flogo-config-mgmt/consul
```

3. use patched version of project-core/flow
```
sed -i 's|github.com/project-flogo/core .*|github.com/project-flogo/core external_prop_resolve|' ./src/go.mod
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
export CONSUL_HTTP_ADDR=127.0.0.1:8500         # optional as it is the default value
export FLOGO_APP_CONFIG_EXTERNAL=consul        # enable the Consul external property resolver
```

2. run the application
```
./bin/simple-config
```

You should see a log like:
```
INFO   [activity-flogo-log] - Consul message
```

#### with Consul, overriding with a profile file

1. create a JSON profile file with an overridden value for ```log.message```
```
echo '{"log.message": "Profile file message"}' | jq '.' > profile.json
```

2. set following environment variables
```
export CONSUL_HTTP_ADDR=127.0.0.1:8500         # optional as it is the default value
export FLOGO_APP_CONFIG_EXTERNAL=consul        # enable the Consul external property resolver
export FLOGO_APP_CONFIG_PROFILES=profile.json  # override with a property value in a JSON profile file
```

3. run the application
```
./bin/simple-config
```

You should see a log like:
```
INFO   [activity-flogo-log] - Profile file message
```

The property value in the profile file has overridden the Consul value.

The Consul API is not even called since the value of the property was resolved successfully by a resolver with a higher
priority.

However the Consul resolver is still configured and active for potential other properties to resolve.

#### with Consul, overriding with an environment variable

1. set following environment variables
```
export CONSUL_HTTP_ADDR=127.0.0.1:8500         # optional as it is the default value
export FLOGO_APP_CONFIG_EXTERNAL=consul        # enable the Consul external property resolver
export FLOGO_APP_CONFIG_PROFILES=profile.json  # override with a property value in a JSON profile file
export LOG_MESSAGE="Env var message"           # override with a (canonical) environment variable
```

2. run the application
```
./bin/simple-config
```

You should see a log like:
```
INFO   [activity-flogo-log] - Env var message
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
unset FLOGO_APP_CONFIG_ENV_VARS
unset FLOGO_APP_CONFIG_EXTERNAL
unset LOG_MESSAGE
```

## Configuration

The configuration is set by a very limited set of environment variables. Some (three) are builtin in the Flogo engine,
the others are defined by external resolvers such as this current one.

### Flogo level

| Name                       | Default value | Required                      | Example value                                      |
|----------------------------|---------------|-------------------------------|----------------------------------------------------|
| FLOGO_APP_CONFIG_ENV_VARS  | true          | no                            | false, to disable environment variables resolution |
| FLOGO_APP_CONFIG_PROFILES  | ø             | for "file" resolver only      | "default.json", "common.json,prod.json"            |
| FLOGO_APP_CONFIG_EXTERNAL  | ø             | for "external" resolvers only | "consul", "etcd", ...                              |

### Consul level

| Name              | Default value   | Required | Example value           |
|-------------------|-----------------|----------|-------------------------|
| CONSUL_HTTP_ADDR  | 127.0.0.1:8500  | no       | consul.company.com:8500 |
