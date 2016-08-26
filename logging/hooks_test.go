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
	h := parseHook(map[string]interface{}{"name": "simple", "host": "localhost:3928"})
	if assert.NotNil(t, h) {
		assert.IsType(t, &simpleHook{}, h)
		assert.Equal(t, "localhost:3928", h.(*simpleHook).Host)
	}

	hh := parseHook(map[string]interface{}{
		"name":    "other",
		"host":    "example.com",
		"port":    3939,
		"replace": true,
	})
	if assert.NotNil(t, hh) {
		assert.IsType(t, &otherHook{}, hh)
		o, _ := hh.(*otherHook)
		assert.Equal(t, "example.com", o.Host)
		assert.Equal(t, 3939, o.Port)
		assert.True(t, o.Replace)
	}

	assert.Nil(t, parseHook(map[string]interface{}{"name": []int{}}))
	assert.Nil(t, parseHook(map[string]interface{}{"name": "not-there"}))
}

func TestLogging_ParseHooks(t *testing.T) {
	v := viper.New()
	v.Set("hooks", []interface{}{
		map[string]interface{}{"name": "simple", "host": "blah"},
		map[string]interface{}{"name": "other", "host": "blah", "port": 3939, "replace": true},
		struct{ Name string }{"invalid"},
		map[string]interface{}{"name": []int{}},
	})

	assert.Len(t, parseHooks(v), 2)

	vv := viper.New()
	vv.Set("hooks", map[interface{}]interface{}{"name": "other", "host": "blah", "port": 3939, "replace": true})
	assert.Len(t, parseHooks(vv), 1)

	vvv := viper.New()
	vvv.Set("hooks", struct{ Name string }{"invalid"})
	assert.Empty(t, parseHooks(vvv))

	vvvv := viper.New()
	vvvv.Set("hooks", map[interface{}]interface{}{"name": []int{}})
	assert.Empty(t, parseHooks(vvvv))
}

func TestLogging_KnownHooks(t *testing.T) {
	kw := KnownHooks()
	assert.Len(t, kw, len(knownHooks))
}

func init() {
	RegisterHook("simple", func(c *viper.Viper) logrus.Hook {
		return &simpleHook{
			Host: c.GetString("host"),
		}
	})

	RegisterHook("other", func(c *viper.Viper) logrus.Hook {
		var ch otherHook
		if err := c.Unmarshal(&ch); err != nil {
			return nil
		}
		return &ch
	})
}

type simpleHook struct {
	Host string
}

func (s *simpleHook) Levels() []logrus.Level { return logrus.AllLevels }
func (s *simpleHook) Fire(entry *logrus.Entry) error {
	return nil
}

type otherHook struct {
	Host    string
	Port    int
	Replace bool
}

func (s *otherHook) Levels() []logrus.Level { return logrus.AllLevels }
func (s *otherHook) Fire(entry *logrus.Entry) error {
	return nil
}
