/*Package tracing implements a super simple tracer/profiler based on go-metrics.

  var tracer = NewTracer("", nil, nil)

  func TraceThis() {
      defer tracer.Trace()()
      // do work here
  }

  func FunctionWithUglyName() {
      defer tracer.Trace("PrettyName")()
      // do work here
  }

You will then be able to get information about timings for methods.

When used with the github.com/casualjim/middlewares package you can get a JSON document
with the report from /audit/metrics.
*/
package tracing
