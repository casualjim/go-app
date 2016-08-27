# Go App [![Build Status](https://ci.vmware.run/api/badges/casualjim/go-app/status.svg)](https://ci.vmware.run/casualjim/go-app) [![Coverage](https://coverage.vmware.run/badges/casualjim/go-app/coverage.svg)](https://coverage.vmware.run/casualjim/go-app)

A library to provide application level context, config reloading and log configuration.
This is a companion to golangs context.

This package is one of those tools you won't always need, but when you need it you'll know you need it.

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

```hcl
logging {
  root {
    level = "debug"
    hooks = [
      { name = "journald" }
    ]
  
    child1 {
      level = "info"
      hooks = [
        { 
          name = "file"
          path = "./app.log"
        },
        {
          name     = "syslog"
          network  = "udp"
          host     = "localhost:514"
          priority = "info"
        }
      ]
    }
  }

  alerts {
    level  = "error"
    writer = "stderr"
  }
}
```

or the more concise yaml:

```yaml
logging:
  root:
    level: Debug
    hooks:
      - name: journald
    writer: stderr
    child1:
      level: Info
      hooks:
        - name: file
          path: ./app.log
        - name: syslog
          network: udp
          host: localhost:514
          priority: info
  alerts:
    level: error
    writer: stderr
 ```
