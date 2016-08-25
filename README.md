# Go App [![Build Status](https://ci.vmware.run/api/badges/casualjim/go-app/status.svg)](https://ci.vmware.run/casualjim/go-app)

A library to provide application level context, config reloading and log configuration.
This is a companion to golangs context.

## Depends on

* logrus
* viper
* go-metrics

## Includes 

* tiny tracer
* logging config through viper
* watching of configuration file
* watching of remote configuration

## configuration

The configuration can be expressed in JSON, YAML, TOML or HCL.

example: 

```yaml
tracing:
  log: true
logging:
  level: Debug
  hooks:
    - journald
  context:
    env: dev
  child1:
    level: Info
    hooks:
      - file
    context:
      env: dev  
```

