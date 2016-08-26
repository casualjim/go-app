package logging

import (
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// CreateFormatter is a factory for creating formatters configured through viper
type CreateFormatter func(*viper.Viper) logrus.Formatter

var (
	knownFormatters map[string]CreateFormatter
	formattersLock  *sync.Mutex

	// DefaultFormatter the fallback formatter when no registered one matches
	DefaultFormatter CreateFormatter
)

func init() {
	formattersLock = new(sync.Mutex)

	DefaultFormatter = func(c *viper.Viper) logrus.Formatter {
		return &logrus.TextFormatter{}
	}

	knownFormatters = make(map[string]CreateFormatter, 10)
	knownFormatters["json"] = func(c *viper.Viper) logrus.Formatter {
		return &logrus.JSONFormatter{}
	}
	knownFormatters["text"] = func(c *viper.Viper) logrus.Formatter {
		return &logrus.TextFormatter{}
	}
}

func parseFormatter(fmtr string, v *viper.Viper) logrus.Formatter {
	formattersLock.Lock()
	defer formattersLock.Unlock()

	if create, ok := knownFormatters[strings.ToLower(fmtr)]; ok {
		return create(v)
	}

	logrus.Debugf("unknown formatter %q, falling back to default", fmtr)
	return DefaultFormatter(v)
}

// RegisterFormatter registers a formatter for use in config files
func RegisterFormatter(name string, factory CreateFormatter) {
	formattersLock.Lock()
	knownFormatters[strings.ToLower(name)] = factory
	formattersLock.Unlock()
}
