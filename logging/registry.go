package logging

import (
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// DefaultRegistry for loggers
var DefaultRegistry *LoggerRegistry

// RootName of the root logger, defaults to root
var RootName string

func init() {
	RootName = "root"
	DefaultRegistry = NewRegistry(viper.GetViper())
}

// LoggerRegistry represents a registry for known loggers
type LoggerRegistry struct {
	config *viper.Viper
	store  map[string]Logger
	lock   *sync.Mutex
}

// NewRegistry creates a new logger registry
func NewRegistry(cfg *viper.Viper) *LoggerRegistry {
	c := cfg
	if c.InConfig("logging") {
		c = cfg.Sub("logging")
	}

	keys := c.AllKeys()
	store := make(map[string]Logger, len(keys))
	reg := &LoggerRegistry{
		store:  store,
		config: c,
		lock:   new(sync.Mutex),
	}

	for _, k := range keys {
		v := c
		if c.InConfig(k) {
			v = c.Sub(k)
		}

		addLoggingDefaults(v)

		nm := k
		if v.IsSet("name") {
			nm = v.GetString("name")
		}

		l := newNamedLogger(k, logrus.Fields{"module": nm}, v, nil)
		l.reg = reg
		reg.Register(k, l)
	}
	if len(keys) == 0 {
		l := newNamedLogger(RootName, logrus.Fields{"module": RootName}, c, nil)
		l.reg = reg
		reg.Register(RootName, l)
	}

	return reg
}

// Get a logger by name, returns nil when logger doesn't exist.
// GetOK is the safe method to use.
func (r *LoggerRegistry) Get(name string) Logger {
	l, ok := r.GetOK(name)
	if !ok {
		return nil
	}
	return l
}

// GetOK a logger by name, boolean is true when a logger was found
func (r *LoggerRegistry) GetOK(name string) (Logger, bool) {
	r.lock.Lock()
	res, ok := r.store[strings.ToLower(name)]
	r.lock.Unlock()
	return res, ok
}

// Register a logger in this registry, overrides existing keys
func (r *LoggerRegistry) Register(path string, logger Logger) {
	r.lock.Lock()
	r.store[strings.ToLower(path)] = logger
	r.lock.Unlock()
}

// Root returns the root logger, the name is configurable through the RootName variable
func (r *LoggerRegistry) Root() Logger {
	return r.Get(RootName)
}

// Reload all the loggers with the new config
func (r *LoggerRegistry) Reload(cfg *viper.Viper) error {
	prev := r.config
	_ = prev
	r.config = cfg
	return nil
}
