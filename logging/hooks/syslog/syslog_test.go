package syslog

import (
	"testing"

	"github.com/casualjim/go-app/logging"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSyslogHook(t *testing.T) {
	if assert.Contains(t, logging.KnownHooks(), "syslog") {
		v := viper.New()
		v.Set("hooks", map[interface{}]interface{}{
			"name": "syslog",
		})
		logging.New(nil, v)
	}
}

func TestSyslogHookPanics(t *testing.T) {
	prioMap["invalid"] = 2993
	if assert.Contains(t, logging.KnownHooks(), "syslog") {
		v := viper.New()
		v.Set("hooks", map[interface{}]interface{}{
			"name":     "syslog",
			"facility": "invalid",
		})
		assert.Panics(t, func() { logging.New(nil, v) })
	}
}

func TestSyslogHookWithPrio(t *testing.T) {
	if assert.Contains(t, logging.KnownHooks(), "syslog") {
		v := viper.New()
		v.Set("hooks", map[interface{}]interface{}{
			"name":     "syslog",
			"severity": "error",
		})
		logging.New(nil, v)
	}
}
