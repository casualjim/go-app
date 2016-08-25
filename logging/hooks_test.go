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
