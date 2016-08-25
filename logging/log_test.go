package logging

import (
	"bytes"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLogging_ParseLevel(t *testing.T) {
	valid := map[string]logrus.Level{
		"debug":   logrus.DebugLevel,
		"info":    logrus.InfoLevel,
		"warn":    logrus.WarnLevel,
		"warning": logrus.WarnLevel,
		"error":   logrus.ErrorLevel,
		"fatal":   logrus.FatalLevel,
		"panic":   logrus.PanicLevel,
	}

	for k, lvl := range valid {
		assert.Equal(t, lvl, parseLevel(k))
	}

	prevOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(prevOut)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	assert.Equal(t, logrus.ErrorLevel, parseLevel(""))
	assert.NotEmpty(t, buf.String())

	var buf2 bytes.Buffer
	logrus.SetOutput(&buf2)
	assert.Equal(t, logrus.ErrorLevel, parseLevel("not a level"))
	assert.NotEmpty(t, buf.String())
}

func TestLogging_AddDefaults(t *testing.T) {
	c := viper.New()
	addLoggingDefaults(c)
	assert.Equal(t, "info", c.GetString("level"))
	assert.Equal(t, map[interface{}]interface{}{
		"stderr": nil,
	}, c.Get("writer"))
}

func TestLogging_MergeConfig(t *testing.T) {
	c := viper.New()
	c.Set("level", "debug")
	c.Set("format", "json")
	c.Set("writer", map[interface{}]interface{}{"stdout": nil})

	cc := viper.New()

	mergeConfig(cc, c)
	assert.Equal(t, "debug", cc.GetString("level"))
	cc.Set("level", "warn")
	assert.Equal(t, "warn", cc.GetString("level"))
	assert.Equal(t, "json", cc.GetString("format"))
	assert.Equal(t, map[interface{}]interface{}{"stdout": nil}, cc.Get("writer"))
}

func TestLogging_CreateNamedLogger(t *testing.T) {
	c := viper.New()
	addLoggingDefaults(c)

	l := newNamedLogger("the-name", logrus.Fields{"some": "field"}, c).(*defaultLogger)
	assert.Equal(t, c, l.Config())
	assert.Equal(t, logrus.Fields{"some": "field"}, l.Fields())
	assert.Equal(t, logrus.InfoLevel, l.Entry.Logger.Level)
	assert.IsType(t, &logrus.TextFormatter{}, l.Entry.Logger.Formatter)
}

func TestLogging_New(t *testing.T) {
	c := viper.New()
	l := New(logrus.Fields{"some": "field"}, c).(*defaultLogger)

	assert.Equal(t, c, l.Config())
	assert.Equal(t, logrus.Fields{"module": "root", "some": "field"}, l.Fields())
	assert.Equal(t, logrus.InfoLevel, l.Entry.Logger.Level)
	assert.IsType(t, &logrus.TextFormatter{}, l.Entry.Logger.Formatter)

	l2 := New(nil, c).(*defaultLogger)
	assert.Equal(t, c, l2.Config())
	assert.Equal(t, logrus.Fields{"module": "root"}, l2.Fields())
	assert.Equal(t, logrus.InfoLevel, l2.Entry.Logger.Level)
	assert.IsType(t, &logrus.TextFormatter{}, l2.Entry.Logger.Formatter)
}

func TestLogging_NewChildLogger(t *testing.T) {
	cfgb := []byte(`---
level: debug
formatter: json
someModule:
  level: warn
  writer:
    stderr:
`)

	c := viper.New()
	c.SetConfigType("YAML")
	if assert.NoError(t, c.ReadConfig(bytes.NewBuffer(cfgb))) {

		l := New(logrus.Fields{"some": "field"}, c).(*defaultLogger)
		assert.Equal(t, c, l.Config())
		assert.Equal(t, logrus.Fields{"module": "root", "some": "field"}, l.Fields())
		assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
		assert.IsType(t, &logrus.TextFormatter{}, l.Entry.Logger.Formatter)

		cl := l.New("someModule", logrus.Fields{"other": "value"}).(*defaultLogger)
		assert.Equal(t, mergeConfig(c.Sub("somemodule"), c), cl.Config())
		assert.Equal(t, logrus.Fields{"module": "someModule", "some": "field", "other": "value"}, cl.Fields())
		assert.Equal(t, logrus.WarnLevel, cl.Entry.Logger.Level)
		assert.IsType(t, &logrus.TextFormatter{}, cl.Entry.Logger.Formatter)
	}
}

func TestLogging_SharedChildLogger(t *testing.T) {
	cfgb := []byte(`---
level: debug
formatter: json
someModule:
  level: warn
  writer:
    stderr:
`)

	c := viper.New()
	c.SetConfigType("YAML")
	if assert.NoError(t, c.ReadConfig(bytes.NewBuffer(cfgb))) {

		l := New(logrus.Fields{"some": "field"}, c).(*defaultLogger)
		assert.Equal(t, c, l.Config())
		assert.Equal(t, logrus.Fields{"module": "root", "some": "field"}, l.Fields())
		assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
		assert.IsType(t, &logrus.TextFormatter{}, l.Entry.Logger.Formatter)

		cl := l.New("otherModule", logrus.Fields{"other": "value"}).(*defaultLogger)
		assert.Equal(t, c, cl.Config())
		assert.Equal(t, logrus.Fields{"module": "otherModule", "some": "field", "other": "value"}, cl.Fields())
		assert.Equal(t, logrus.DebugLevel, cl.Entry.Logger.Level)
		assert.IsType(t, &logrus.TextFormatter{}, cl.Entry.Logger.Formatter)
	}
}
