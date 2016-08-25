package logging

import (
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// CreateHook creates a hook based on a viper config
type CreateHook func(*viper.Viper) logrus.Hook

var (
	knownHooks map[string]CreateHook
	hooksLock  *sync.Mutex
)

func init() { // using init avoids a race
	hooksLock = new(sync.Mutex)
	knownHooks = make(map[string]CreateHook, 50)
}

// RegisterHook for use through configuration system
func RegisterHook(name string, factory CreateHook) {
	hooksLock.Lock()
	knownHooks[strings.ToLower(name)] = factory
	hooksLock.Unlock()
}
