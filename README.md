# Configuration management for Flogo 

This repository aims at managing the configuration of Flogo applications.

It does so by providing several [custom property *resolvers* for Flogo](#custom-property-resolvers) since most of the
configuration of a Flogo application is [stored in properties](#about-properties-resolution-in-flogo).

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

### Resolvers

Properties values are retrieved by different *resolvers*:

#### Builtin resolvers

* **Environment variables**: simplest method when deploying the application in containers (Docker, Kubernetes, ...).
* **JSON files**: best method to define different *profiles* of properties such as *dev*, *integration*, *prod*.
It is also possible to merge different files (*common.json* + *prod.json* for instance).

#### Custom resolvers
* **Custom _resolver_**: can be any system supported by a Flogo contribution such as Consul, etcd, k8s ConfigMaps...
See the [complete list below](#custom-property-resolvers).

#### Default
* If no *resolver* is configured or if none of them returns a value for a property, the value specified in the
*flogo.json* configuration file of the application for this property is used by default.

### Priorities

The default order of priority for properties resolution is as follows:

> Custom *resolver* **>**  Environment variable *resolver* **>** JSON files *resolver* **>** default value

* This priority order is the default order when a custom *resolver* and the builtin *resolvers* are
[registered](#register) but none of them has been explicitly [enabled](#enable).
* Each *resolver* is called only if it is [registered](#register) and [enabled](#enable)
(see [Configuration](#configuration)).
* Each property uses the value from the first *resolver* succeeding.
* Resolution is performed at the start of the Flogo engine and properties are then cached during the lifetime of the
application.

## Custom property resolvers

* using a [Consul KV Store](./consul/README.md)

## Configuration

To be used during resolution phase, a *resolver* must be [**registered**](#register) and [**enabled**](#enable).

### Register

To register a property *resolver*, the corresponding Go package must be imported in the application.
It is done seamlessly with the ```flogo install``` command, for instance:

```
flogo install github.com/square-it/flogo-config-mgmt/consul@v0.0.1
```

To register the builtin property *resolvers* (from *github.com/project-flogo/core@v0.9.0-alpha.6*):
```
flogo install github.com/project-flogo/core/app/propertyresolver@v0.9.0-alpha.6
```

If only one property *resolver* is registered (with the builtin *resolvers* or not): it will be
**automatically enabled**.

If the builtin property *resolvers* are registered, they will be used as fallback *resolvers* if a property
has no value in the custom property *resolver*.

Otherwise, or if a custom configuration is required, [enable explicitly the resolvers](#enable).

### Enable

Property *resolvers* can be enabled by setting the ```FLOGO_APP_PROP_RESOLVERS``` environment variable.