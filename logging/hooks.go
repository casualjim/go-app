package logging

import (
	"sort"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cast"
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

// KnownHooks returns the list of keys for the registered hooks
func KnownHooks() []string {
	var hooks []string
	for k := range knownHooks {
		hooks = append(hooks, k)
	}
	sort.Strings(hooks)
	return hooks
}

func parseHooks(v *viper.Viper) []logrus.Hook {
	if v.IsSet("hooks") {
		hs := v.Get("hooks")
		var res []logrus.Hook
		switch ele := hs.(type) {
		case []interface{}:
			for _, v := range ele {
				mm, err := cast.ToStringMapE(v)
				if err != nil {
					continue
				}
				h := parseHook(mm)
				if h == nil {
					continue
				}
				res = append(res, h)
			}
			return res
		case map[interface{}]interface{}:
			h := parseHook(v.GetStringMap("hooks"))
			if h == nil {
				return nil
			}
			res = append(res, h)
			return res
		}
	}
	return nil
}

func parseHook(v map[string]interface{}) logrus.Hook {
	if nme, ok := v["name"]; ok {
		name, err := cast.ToStringE(nme)
		if err != nil {
			return nil
		}

		hooksLock.Lock()
		if create, ok := knownHooks[strings.ToLower(name)]; ok {
			vv := viper.New()
			vv.Set("nested", v)
			h := create(vv.Sub("nested"))
			hooksLock.Unlock()
			return h
		}
		hooksLock.Unlock()
	}
	return nil
}
