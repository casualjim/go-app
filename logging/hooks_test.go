package logging

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLogging_RegisterHook(t *testing.T) {
	assert.NotContains(t, knownHooks, "null")
	RegisterHook("null", func(_ *viper.Viper) logrus.Hook { return nil })
	assert.Contains(t, knownHooks, "null")
}

func TestLogging_ParseHook(t *testing.T) {
	v := viper.New()
	assert.Empty(t, v.GetString("not there"))
}

func TestLogging_ParseHooks(t *testing.T) {

}
