package tracing

import (
	"bytes"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewTracer(t *testing.T) {
	assert := assert.New(t)
	tracer := New("testModule", nil, nil).(*defaultTracing)
	if assert.NotNil(tracer) {
		assert.NotNil(tracer.logger)
		assert.NotNil(tracer.registry)
	}

	tr2 := New("", nil, nil).(*defaultTracing)
	if assert.NotNil(tr2) {
		assert.NotNil(tr2.logger)
		assert.NotNil(tr2.registry)
		assert.Equal(trace, tr2.logger.(*logrus.Entry).Data["name"])
	}
}

func TestTracerLog(t *testing.T) {

	assert := assert.New(t)

	prevLevel := logrus.GetLevel()
	prevOut := logrus.StandardLogger().Out
	var buf countingWriter
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.DebugLevel)
	defer logrus.SetOutput(prevOut)
	defer logrus.SetLevel(prevLevel)

	tracer1 := New("testModule", nil, nil).(*defaultTracing)
	if assert.NotNil(tracer1) {
		testFunc(tracer1)
		tracer2 := New("testModule", nil, nil).(*defaultTracing)
		if assert.NotNil(tracer2) {
			tracer2.Trace("myMethod")()
		}
	}
	assert.Equal(4, buf.count)
	t.Log(buf.String())
}

type countingWriter struct {
	count int
	buf   bytes.Buffer
}

func (w *countingWriter) Write(data []byte) (int, error) {
	w.count++
	return w.buf.Write(data)
}

func (w *countingWriter) String() string {
	return w.buf.String()
}

func testFunc(tracer Tracer) {
	defer tracer.Trace()()
}
