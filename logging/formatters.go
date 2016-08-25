package logging

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// CreateFormatter is a factory for creating formatters configured through viper
type CreateFormatter func(*viper.Viper) logrus.Formatter

var (
	knownFormatters map[string]logrus.Formatter
	formattersLock  *sync.Mutex
)

func init() {
	formattersLock = new(sync.Mutex)
	knownFormatters = make(map[string]logrus.Formatter, 10)
}
