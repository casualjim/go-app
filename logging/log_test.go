package logging

import (
	"bytes"
	"os"
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
	origDebug := os.Getenv("DEBUG")
	defer os.Setenv("DEBUG", origDebug)

	os.Setenv("DEBUG", "1")
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	assert.Equal(t, logrus.ErrorLevel, parseLevel(""))
	assert.NotEmpty(t, buf.String())

	var buf2 bytes.Buffer
	logrus.SetOutput(&buf2)
	assert.Equal(t, logrus.ErrorLevel, parseLevel("not a level"))
	assert.NotEmpty(t, buf2.String())

	os.Setenv("DEBUG", "")
	var buf3 bytes.Buffer
	logrus.SetOutput(&buf3)
	assert.Equal(t, logrus.ErrorLevel, parseLevel(""))
	assert.Empty(t, buf3.String())

	var buf4 bytes.Buffer
	logrus.SetOutput(&buf4)
	assert.Equal(t, logrus.ErrorLevel, parseLevel("not a level"))
	assert.Empty(t, buf4.String())
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
	c.Set("writer", "stdout")
	c.Set("hooks", map[interface{}]interface{}{"name": "other", "host": "blah", "port": 3939, "replace": true})

	cc := viper.New()

	mergeConfig(cc, c)
	assert.Equal(t, "debug", cc.GetString("level"))
	cc.Set("level", "warn")
	assert.Equal(t, "warn", cc.GetString("level"))
	assert.Equal(t, "json", cc.GetString("format"))
	assert.Equal(t, "stdout", cc.Get("writer"))
}

func TestLogging_CreateNamedLogger(t *testing.T) {
	c := viper.New()
	addLoggingDefaults(c)

	l := newNamedLogger("the-name", logrus.Fields{"some": "field"}, c, nil)
	assert.Equal(t, c, l.config)
	assert.Equal(t, logrus.Fields{"some": "field"}, l.Entry.Data)
	assert.Equal(t, logrus.InfoLevel, l.Entry.Logger.Level)
	assert.IsType(t, &logrus.TextFormatter{}, l.Entry.Logger.Formatter)
}

func TestLogging_CreateNamedLoggerWithHooks(t *testing.T) {
	c := viper.New()
	c.Set("level", "debug")
	c.Set("format", "json")
	c.Set("writer", "stdout")
	c.Set("hooks", map[interface{}]interface{}{"name": "other", "host": "blah", "port": 3939, "replace": true})
	addLoggingDefaults(c)

	l := newNamedLogger("the-name", logrus.Fields{"some": "field"}, c, nil)
	assert.Equal(t, c, l.config)
	assert.Equal(t, logrus.Fields{"some": "field"}, l.Entry.Data)
	assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
	assert.Equal(t, os.Stdout, l.Entry.Logger.Out)
	assert.IsType(t, &logrus.JSONFormatter{}, l.Entry.Logger.Formatter)
	assert.NotEmpty(t, l.Logger.Hooks)
}

func TestLogging_NewChildLogger(t *testing.T) {
	cfgb := []byte(`---
root:
  level: debug
  formatter: json
  somemodule:
    level: warn
    writer:
      stderr:
`)

	c := viper.New()
	c.SetConfigType("YAML")
	if assert.NoError(t, c.ReadConfig(bytes.NewBuffer(cfgb))) {
		reg := NewRegistry(c, logrus.Fields{"some": "field"})
		// _ = reg
		lr := reg.Root()
		rc := c.Sub(RootName)
		addLoggingDefaults(rc)

		if assert.NotNil(t, lr) {
			l := lr.(*defaultLogger)
			assert.Equal(t, rc, l.config)
			assert.Equal(t, logrus.Fields{"module": "root", "some": "field"}, l.Entry.Data)
			assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
			assert.IsType(t, &logrus.TextFormatter{}, l.Entry.Logger.Formatter)
		}

		cl := lr.New("someModule", logrus.Fields{"other": "value"}).(*defaultLogger)
		assert.Equal(t, mergeConfig(rc.Sub("someModule"), rc), cl.config)
		assert.Equal(t, logrus.Fields{"module": "someModule", "some": "field", "other": "value"}, cl.Entry.Data)
		assert.Equal(t, logrus.WarnLevel, cl.Entry.Logger.Level)
		assert.IsType(t, &logrus.TextFormatter{}, cl.Entry.Logger.Formatter)

		assert.Len(t, reg.store, 2)
	}
}

func TestLogging_SharedChildLogger(t *testing.T) {
	cfgb := []byte(`---
root:
  level: debug
  formatter: json
  somemodule:
    level: warn
    writer:
      stderr:
`)

	c := viper.New()
	c.SetConfigType("YAML")
	if assert.NoError(t, c.ReadConfig(bytes.NewBuffer(cfgb))) {

		reg := NewRegistry(c, nil)
		l := reg.Root().(*defaultLogger)
		rc := c.Sub(RootName)
		addLoggingDefaults(rc)
		assert.Equal(t, rc, l.config)
		assert.Equal(t, logrus.Fields{"module": "root"}, l.Entry.Data)
		assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
		assert.IsType(t, &logrus.TextFormatter{}, l.Entry.Logger.Formatter)

		cl := l.New("otherModule", logrus.Fields{"other": "value"}).(*defaultLogger)
		assert.Equal(t, rc, cl.config)
		assert.Equal(t, logrus.Fields{"module": "otherModule", "other": "value"}, cl.Entry.Data)
		assert.Equal(t, logrus.DebugLevel, cl.Entry.Logger.Level)
		assert.IsType(t, &logrus.TextFormatter{}, cl.Entry.Logger.Formatter)

		assert.Len(t, reg.store, 2)
	}
}

func TestLogging_ChildLoggerFromCache(t *testing.T) {
	cfgb := []byte(`---
root:
  level: debug
  formatter: json
  somemodule:
    level: warn
    writer:
      stderr:
`)

	c := viper.New()
	c.SetConfigType("YAML")
	if assert.NoError(t, c.ReadConfig(bytes.NewBuffer(cfgb))) {

		reg := NewRegistry(c, nil)
		l := reg.Root()
		cl := l.New("otherModule", logrus.Fields{"other": "value"})
		cl2 := l.New("otherModule", logrus.Fields{"other": "value"})
		assert.Equal(t, cl, cl2)
		assert.Len(t, reg.store, 2)
	}
}

func TestLogging_LoggerConfigure(t *testing.T) {
	c := viper.New()
	c.Set("level", "debug")
	c.Set("format", "json")
	c.Set("writer", "stdout")
	c.Set("hooks", map[interface{}]interface{}{"name": "other", "host": "blah", "port": 3939, "replace": true})

	l := newNamedLogger("alerts", logrus.Fields{"mode": "dev"}, c, nil)
	assert.Equal(t, c, l.config)
	assert.Equal(t, logrus.Fields{"mode": "dev"}, l.Fields())
	assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
	assert.Equal(t, os.Stdout, l.Entry.Logger.Out)
	assert.IsType(t, &logrus.JSONFormatter{}, l.Entry.Logger.Formatter)
	assert.NotEmpty(t, l.Logger.Hooks)
	h := l.Logger.Hooks[logrus.InfoLevel]
	assert.Len(t, h, 1)
	assert.IsType(t, &otherHook{}, h[0])
	hh := h[0].(*otherHook)
	assert.Equal(t, "blah", hh.Host)
	assert.Equal(t, 3939, hh.Port)

	c = viper.New()
	c.Set("level", "warn")
	c.Set("format", "text")
	c.Set("writer", "stderr")
	c.Set("hooks", map[interface{}]interface{}{"name": "other", "host": "example", "port": 4444, "replace": true})
	l.Configure(c)
	assert.Equal(t, c, l.config)
	assert.Equal(t, logrus.Fields{"mode": "dev"}, l.Fields())
	assert.Equal(t, logrus.WarnLevel, l.Entry.Logger.Level)
	assert.Equal(t, os.Stderr, l.Entry.Logger.Out)
	assert.IsType(t, &logrus.TextFormatter{}, l.Entry.Logger.Formatter)
	assert.NotEmpty(t, l.Logger.Hooks)
	h = l.Logger.Hooks[logrus.InfoLevel]
	assert.Len(t, h, 1)
	assert.IsType(t, &otherHook{}, h[0])
	hh = h[0].(*otherHook)
	assert.Equal(t, "example", hh.Host)
	assert.Equal(t, 4444, hh.Port)
}
