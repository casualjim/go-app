package logging

import (
	"bytes"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/kr/pretty"
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

	rc4 = []byte(`---
root:
  level: debug
  format: json
  child1:
    level: warn
    format: text
    child1child:
      level: error
      format: json
alerts:
  format: json
`)

	rc5 = []byte(`---
root:
  level: warn
  format: text
  child1:
    level: error
    format: json
    child1child:
      level: info
      format: text
alerts:
  level: error
  format: text
`)
)

func TestLogging_NewRegistry(t *testing.T) {
	v1 := viper.New()
	v1.SetConfigType("YAML")
	if assert.NoError(t, v1.ReadConfig(bytes.NewBuffer(rc1))) {
		r1 := NewRegistry(v1, nil)
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
		r2 := NewRegistry(v2, nil)
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
		r3 := NewRegistry(v3, nil)
		assert.Equal(t, v3, r3.config)
		assert.Equal(t, "PagerDuty", r3.config.GetString("alerts.name"))
		l3, ok3 := r3.store["alerts"]

		if assert.True(t, ok3) {
			l33 := l3.(*defaultLogger)
			assert.Equal(t, logrus.InfoLevel, l33.Entry.Logger.Level)
			assert.IsType(t, &logrus.JSONFormatter{}, l33.Logger.Formatter)

			if !assert.Equal(t, "PagerDuty", l33.Entry.Data["module"], pretty.Sprint(l33.config.AllSettings())) {
				pretty.Println(v3.AllSettings())
				pretty.Println(l33.path)
			}
			assert.Equal(t, "PagerDuty", l33.config.GetString("name"))
		}
	}

}

func TestLogging_Registry_GetOK(t *testing.T) {
	v1 := viper.New()
	v1.SetConfigType("YAML")
	if assert.NoError(t, v1.ReadConfig(bytes.NewBuffer(rc1))) {
		r1 := NewRegistry(v1, nil)
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
		r1 := NewRegistry(v1, nil)
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
		r1 := NewRegistry(v1, logrus.Fields{"some": "field"})
		assert.Equal(t, v1.Sub("logging"), r1.config)
		l1 := r1.Root()
		if assert.NotNil(t, l1) {
			l := l1.(*defaultLogger)
			assert.Equal(t, logrus.DebugLevel, l.Entry.Logger.Level)
			assert.IsType(t, &logrus.JSONFormatter{}, l.Logger.Formatter)
			assert.Equal(t, logrus.Fields{"module": "root", "some": "field"}, l.Data)
		}
		assert.NotNil(t, r1.Writer())
	}
}

func TestLogging_Defaults(t *testing.T) {
	v1 := viper.New()
	r1 := NewRegistry(v1, nil)
	assert.Len(t, r1.store, 1)
	assert.IsType(t, &defaultLogger{}, r1.Root())
	assert.Equal(t, logrus.Fields{"module": "root"}, r1.Root().Fields())

	r2 := NewRegistry(nil, nil)
	assert.Len(t, r2.store, 1)
	assert.IsType(t, &defaultLogger{}, r2.Root())
	assert.Equal(t, logrus.Fields{"module": "root"}, r2.Root().Fields())

	r3 := NewRegistry(nil, logrus.Fields{"some": "field"})
	assert.Len(t, r3.store, 1)
	assert.IsType(t, &defaultLogger{}, r3.Root())
	assert.Equal(t, logrus.Fields{"module": "root", "some": "field"}, r3.Root().Fields())
}

func TestLogging_Reload(t *testing.T) {
	assert := assert.New(t)
	v1 := viper.New()
	v1.SetConfigType("yaml")
	if assert.NoError(v1.ReadConfig(bytes.NewBuffer(rc4))) {
		// create registry
		reg := NewRegistry(v1, nil)
		root := reg.Root().(*defaultLogger)
		alerts := reg.Get("alerts").(*defaultLogger)
		// create nested child loggers
		child1 := root.New("child1", logrus.Fields{"mode": "dev"}).(*defaultLogger)
		child2 := root.New("child2", logrus.Fields{"mode": "dev"}).(*defaultLogger)
		child1child := child1.New("child1child", logrus.Fields{"some": "field"}).(*defaultLogger)
		child2child := child2.New("child2child", logrus.Fields{"other": "field"}).(*defaultLogger)
		// verify configuration
		assert.Equal(logrus.DebugLevel, root.Logger.Level)
		assert.IsType(&logrus.JSONFormatter{}, root.Logger.Formatter)
		assert.Equal(logrus.InfoLevel, alerts.Logger.Level)
		assert.IsType(&logrus.JSONFormatter{}, alerts.Logger.Formatter)
		assert.Equal(root.Logger, child2.Logger)
		assert.Equal(root.Logger, child2child.Logger)
		assert.Equal(logrus.WarnLevel, child1.Logger.Level)
		assert.IsType(&logrus.TextFormatter{}, child1.Logger.Formatter)
		assert.Equal(logrus.ErrorLevel, child1child.Logger.Level)
		assert.IsType(&logrus.JSONFormatter{}, child1child.Logger.Formatter)

		// upate the config for all loggers
		if assert.NoError(v1.ReadConfig(bytes.NewBuffer(rc5))) {
			// call reload
			reg.Reload()
			// verify new configuration
			assert.Equal(logrus.WarnLevel, root.Logger.Level)
			assert.IsType(&logrus.TextFormatter{}, root.Logger.Formatter)
			assert.Equal(logrus.ErrorLevel, alerts.Logger.Level)
			assert.IsType(&logrus.TextFormatter{}, alerts.Logger.Formatter)
			assert.Equal(root.Logger, child2.Logger)
			assert.Equal(root.Logger, child2child.Logger)
			assert.Equal(logrus.ErrorLevel, child1.Logger.Level)
			assert.IsType(&logrus.JSONFormatter{}, child1.Logger.Formatter)
			assert.Equal(logrus.InfoLevel, child1child.Logger.Level)
			assert.IsType(&logrus.TextFormatter{}, child1child.Logger.Formatter)
		}
	}
}

func TestLogging_LongestMatchingPath(t *testing.T) {
	cs := `---
root:
  value: rootLogger
  child1:
    value: childLogger
    child1child:
      value: child1childLogger
alerts:
  value: alertsLogger
`
	v1 := viper.New()
	v1.SetConfigType("yaml")
	v1.ReadConfig(bytes.NewBufferString(cs))

	expected := []struct {
		Value string
		Path  string
	}{
		{"rootLogger", "root"},
		{"alertsLogger", "alerts"},
		{"childLogger", "root.child1"},
		{"child1childLogger", "root.child1.child1child"},
		{"rootLogger", "root.child2"},
		{"rootLogger", "root.child2.child2child"},
	}

	for _, v := range expected {
		f := findLongestMatchingPath(v.Path, v1)
		if assert.NotNil(t, f, "%q can't be nil", v.Path) {
			assert.Equal(t, v.Value, f.GetString("value"))
		}
	}

	assert.Nil(t, findLongestMatchingPath("not-there", v1))
}
