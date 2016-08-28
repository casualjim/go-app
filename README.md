# Go App [![Build Status](https://ci.vmware.run/api/badges/casualjim/go-app/status.svg)](https://ci.vmware.run/casualjim/go-app) [![Coverage](https://coverage.vmware.run/badges/casualjim/go-app/coverage.svg)](https://coverage.vmware.run/casualjim/go-app)

A library to provide application level context, config reloading and log configuration.
This is a companion to golangs context.  
It also tries to provide an extensible way for adding log hooks without requiring to download all of github.

This package is one of those tools you won't always need, but when you need it you'll know you need it.

## Depends on

* [logrus](https://github.com/Sirupsen/logrus)
* [viper](https://github.com/spf13/viper)
* [go-metrics](github.com/rcrowley/go-metrics)

## Includes 

* [tiny tracer](#tracer)
* [modular initialization](#modular-initialization)
* [logging config through viper](#logger-configuration)
* watching of configuration file, for online reconfiguration of loggers and modules
* watching of remote configuration, for online reconfiguration of loggers and modules

## Tracer

Using the tracer requires that you put a line a the top of a method:

```go
var tracer = NewTracer("", nil, nil)

func TraceThis() {
  defer tracer.Trace()()

  // do work here
}

func FunctionWithUglyName() {
  defer tracer.Trace("PrettyName")()

  // do work here
}
```

You will then be able to get information about timings for methods. When you don't specify a key, the package
will walk the stack to find out the method name you want to trace. If you think this is dirty, you can just pass a name to the trace method
which will make you not incur that cost.

When used with the github.com/casualjim/middlewares package you can get a JSON document
with the report from $baseurl/audit/metrics.

## Modular initialization

Implements a very simple application context that does allows for modular initialization with a deterministic init order.

A module has a simple 4 phase lifecycle: Init, Start, Reload and Stop. You can enable or disable a feature in the config.
This hooks into the watching infrastructure, so you can also enable or disable modules by just editing config or changing a remote value.

Name | Description
-----|------------
Init | Called on initial creation of the module
Start | Called when the module is started, or enabled at runtime
Reload | Called when the config has changed and the module needs to reconfigure itself
Stop | Called when the module is stopped, or disabled at runtime

Each module is identified by a unique name, this defaults to its package name, 

### Usage

To use it, a package that serves as a module needs to export a method or variable that implements the Module interface.

```go
package orders

var Module app.Module
func init() {
  Module = new(module)
}

type module struct {

}

func (m *module) Init(app app.Application) error {
  orders := new(ordersService)
  app.Set("ordersService", orders)
  orders.app = app 
  return nil
}

func (m *module) Start(app app.Application) error {
  // perform any potentially needed cleanup at start of process
}

func (m *module) Stop(app app.Application) error {
  // perform any potentially needed cleanup before shutdown of process
}

func (m *module) Reload(app app.Application) error {
  // peform any necessary configuration of the active modules, potentially enable/disable new services etc
}

type Order struct {
  ID      int64
  Product int64
}

type odersService struct {
  app app.Application
}

func (o *ordersService) Create(o *Order) error {
  var db OrdersStore
  o.app.Get("ordersDb", &db)  
  return db.Save(o)
}
```

Ideally the module struct is completely stateless which makes using it from concurrent tests easier.

## Logger Configuration

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
