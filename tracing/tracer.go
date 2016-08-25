package tracing

import (
	"runtime"
	"time"

	logrus "github.com/Sirupsen/logrus"
	metrics "github.com/rcrowley/go-metrics"
)

const (
	noMethodName = "<anonymous>"
	trace        = "trace"
)

// Tracer interface that represents a tracer in golang
type Tracer interface {
	Trace(name ...string) func()
}

// NewTracer creates a new tracer object with the specified configuration
// When the config is nil the tracer will use default values for the config,
// this is equivalent to
//
//      name: global
//
// Usage of the tracer:
//
//      var tracer = NewTracer(&Config{Name:"restapi"})
//
//      func TraceThis() {
//          defer tracer.Trace()()
//          /* do a work */
//      }
//
//      func FunctionWithUglyName() {
//          defer tracer.Trace("PrettyName")()
//      }
//
func NewTracer(name string, logger logrus.FieldLogger, registry metrics.Registry) Tracer {
	nm := name
	if nm == "" {
		nm = trace
	}

	var bl = logger
	if bl == nil {
		bl = logrus.WithField("name", nm)
	}

	reg := registry
	if reg == nil {
		reg = metrics.DefaultRegistry
	}
	return &defaultTracing{logger: bl, registry: reg}
}

type defaultTracing struct {
	logger   logrus.FieldLogger
	registry metrics.Registry
}

func (d *defaultTracing) Trace(methods ...string) func() {
	var method string
	if len(methods) == 0 || methods[0] == "" {
		method = noMethodName
		pc, _, _, ok := runtime.Caller(1)
		if ok {
			fun := runtime.FuncForPC(pc)
			if fun != nil {
				method = fun.Name()
			}
		}
	} else {
		method = methods[0]
	}

	timer := metrics.GetOrRegisterTimer(method, d.registry)
	d.logger.Debugf("Enter %s ", method)

	start := time.Now()
	return func() {
		timer.UpdateSince(start)
		d.logger.Debugf("Leave %s took %v", method, time.Now().Sub(start))
	}
}
