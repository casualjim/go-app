package tracing

import (
	"bytes"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testHandler struct {
	cnt int
}

func (t *testHandler) Log(r *logrus.Entry) error {
	t.cnt++
	return nil
}

func TestNewTracer(t *testing.T) {
	assert := assert.New(t)
	tracer := NewTracer("testModule", nil, nil).(*defaultTracing)
	if assert.NotNil(tracer) {
		assert.NotNil(tracer.logger)
	}
}

func TestTracerLog(t *testing.T) {
	assert := assert.New(t)

	var buf countingWriter
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.DebugLevel)
	defer logrus.SetOutput(os.Stderr)

	tracer1 := NewTracer("testModule", nil, nil).(*defaultTracing)
	if assert.NotNil(tracer1) {
		testFunc(tracer1)
		tracer2 := NewTracer("testModule", nil, nil).(*defaultTracing)
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
