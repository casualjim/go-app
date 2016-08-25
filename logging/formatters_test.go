package logging

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLogging_ParseFormatter(t *testing.T) {
	c := viper.New()
	addLoggingDefaults(c)
	assert.IsType(t, &logrus.JSONFormatter{}, parseFormatter("json", c))
	assert.IsType(t, &logrus.JSONFormatter{}, parseFormatter("Json", c))
	assert.IsType(t, &logrus.JSONFormatter{}, parseFormatter("JSON", c))

	assert.IsType(t, &logrus.TextFormatter{}, parseFormatter("text", c))

	assert.IsType(t, &logrus.TextFormatter{}, parseFormatter("anything", c))
}

type NullFormatter struct {
}

func (n *NullFormatter) Format(_ *logrus.Entry) ([]byte, error) {
	return nil, nil
}

func TestLogging_RegisterFormatter(t *testing.T) {
	assert.NotContains(t, knownFormatters, "null")
	RegisterFormatter("null", func(_ *viper.Viper) logrus.Formatter {
		return new(NullFormatter)
	})
	assert.Contains(t, knownFormatters, "null")
}
