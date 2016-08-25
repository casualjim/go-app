package logging

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

func addLoggingDefaults(cfg *viper.Viper) {
	cfg.SetDefault("level", "info")
	cfg.SetDefault("writer", map[interface{}]interface{}{
		"stderr": nil,
	})
}

func parseLevel(level string) logrus.Level {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.Warnf("%v, falling back to default of error", err)
		return logrus.ErrorLevel
	}
	return lvl
}

func mergeConfig(child, parent *viper.Viper) *viper.Viper {
	// This merge is only a partial merge
	// the remaining keys are not used to configure a logger but
	// indicate children of the current logger
	child.SetDefault("format", parent.GetString("format"))
	child.SetDefault("level", parent.GetString("level"))
	child.SetDefault("writer", parent.Get("writer"))

	return child
}

func mergeFields(child, parent logrus.Fields) logrus.Fields {
	data := make(logrus.Fields, len(parent)+len(child))
	for k, v := range parent {
		data[k] = v
	}
	for k, v := range child {
		data[k] = v
	}
	return data
}

func newNamedLogger(name string, fields logrus.Fields, cfg *viper.Viper) Logger {
	logger := logrus.New()
	logger.Formatter = parseFormatter(cfg.GetString("format"), cfg)
	logger.Level = parseLevel(cfg.GetString("level"))

	// logger.Hooks = cfg.Hooks

	return &defaultLogger{
		Entry: logrus.Entry{
			Logger: logger,
			Data:   fields,
		},
		config: cfg,
	}
}

// Logger is the interface that application use to log against
type Logger interface {
	logrus.FieldLogger

	Reload() error
	Config() *viper.Viper

	New(string, logrus.Fields) Logger
	Fields() logrus.Fields
}

type defaultLogger struct {
	logrus.Entry

	config *viper.Viper
	path   string
}

// New logger for the given config, if config is nil the default config will be used
func New(fields logrus.Fields, v *viper.Viper) Logger {
	const name = "root"

	addLoggingDefaults(v)

	if fields == nil {
		fields = make(logrus.Fields, 1)
	}
	fields["module"] = name

	return newNamedLogger(name, fields, v)
}

func (d *defaultLogger) New(name string, fields logrus.Fields) Logger {
	data := mergeFields(fields, d.Entry.Data)
	nme := strings.ToLower(name)
	data["module"] = name

	if d.config.InConfig(nme) {
		// new config, so make a new logger
		cfg := mergeConfig(d.config.Sub(nme), d.config)
		return newNamedLogger(name, data, cfg)
	}

	// Share the logger with the parent, same config
	return &defaultLogger{
		Entry: logrus.Entry{
			Logger: d.Entry.Logger,
			Data:   data,
		},
		config: d.config,
	}
}

func (d *defaultLogger) Reload() error {
	return nil
}

func (d *defaultLogger) Config() *viper.Viper {
	return d.config
}

func (d *defaultLogger) Fields() logrus.Fields {
	return d.Entry.Data
}
