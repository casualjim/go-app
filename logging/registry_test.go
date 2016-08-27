package logging

import (
	"bytes"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	rc1 = []byte(`---
logging:
  root:
    level: debug
    format: json
`)
	rc2 = []byte(`---
root:
  level: debug
  format: json
`)

	rc3 = []byte(`---
root:
  level: debug
  format: json
alerts:
  name: PagerDuty
  format: json
`)
)

func TestLogging_NewRegistry(t *testing.T) {
	v1 := viper.New()
	v1.SetConfigType("YAML")
	if assert.NoError(t, v1.ReadConfig(bytes.NewBuffer(rc1))) {
		r1 := NewRegistry(v1)
		assert.Equal(t, v1.Sub("logging"), r1.config)
		l1, ok1 := r1.store["root"]
		if assert.True(t, ok1) {
			l := l1.(*defaultLogger)
			assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
			assert.IsType(t, &logrus.JSONFormatter{}, l.Logger.Formatter)
		}
	}

	v2 := viper.New()
	v2.SetConfigType("YAML")
	if assert.NoError(t, v2.ReadConfig(bytes.NewBuffer(rc2))) {
		r2 := NewRegistry(v2)
		assert.Equal(t, v2, r2.config)
		l2, ok2 := r2.store["root"]

		if assert.True(t, ok2) {
			l := l2.(*defaultLogger)
			assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
			assert.IsType(t, &logrus.JSONFormatter{}, l.Logger.Formatter)
		}
	}

	v3 := viper.New()
	v3.SetConfigType("YAML")
	if assert.NoError(t, v3.ReadConfig(bytes.NewBuffer(rc3))) {
		r3 := NewRegistry(v3)
		assert.Equal(t, v3, r3.config)
		l3, ok3 := r3.store["alerts"]

		if assert.True(t, ok3) {
			l := l3.(*defaultLogger)
			assert.Equal(t, logrus.InfoLevel, l.Entry.Logger.Level)
			assert.IsType(t, &logrus.JSONFormatter{}, l.Logger.Formatter)
			assert.Equal(t, "PagerDuty", l.Data["module"])
		}
	}

}

func TestLogging_Registry_GetOK(t *testing.T) {
	v1 := viper.New()
	v1.SetConfigType("YAML")
	if assert.NoError(t, v1.ReadConfig(bytes.NewBuffer(rc1))) {
		r1 := NewRegistry(v1)
		assert.Equal(t, v1.Sub("logging"), r1.config)
		l1, ok1 := r1.GetOK("root")
		if assert.True(t, ok1) {
			l := l1.(*defaultLogger)
			assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
			assert.IsType(t, &logrus.JSONFormatter{}, l.Logger.Formatter)
		}

		_, ok := r1.GetOK("NotThere")
		assert.False(t, ok)
	}
}

func TestLogging_Registry_Get(t *testing.T) {
	v1 := viper.New()
	v1.SetConfigType("YAML")
	if assert.NoError(t, v1.ReadConfig(bytes.NewBuffer(rc1))) {
		r1 := NewRegistry(v1)
		assert.Equal(t, v1.Sub("logging"), r1.config)
		l1 := r1.Get("root")
		if assert.NotNil(t, l1) {
			l := l1.(*defaultLogger)
			assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
			assert.IsType(t, &logrus.JSONFormatter{}, l.Logger.Formatter)
		}

		assert.Nil(t, r1.Get("NotThere"))
	}
}

func TestLogging_Registry_Root(t *testing.T) {
	v1 := viper.New()
	v1.SetConfigType("YAML")
	if assert.NoError(t, v1.ReadConfig(bytes.NewBuffer(rc1))) {
		r1 := NewRegistry(v1)
		assert.Equal(t, v1.Sub("logging"), r1.config)
		l1 := r1.Root()
		if assert.NotNil(t, l1) {
			l := l1.(*defaultLogger)
			assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
			assert.IsType(t, &logrus.JSONFormatter{}, l.Logger.Formatter)
		}
	}
}

func TestLogging_Defaults(t *testing.T) {
	v1 := viper.New()
	r1 := NewRegistry(v1)
	assert.Len(t, r1.store, 1)
	assert.IsType(t, &defaultLogger{}, r1.Root())
}
