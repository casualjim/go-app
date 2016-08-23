package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/casualjim/go-app/tracing"
	cjm "github.com/casualjim/middlewares"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"
)

var (
	// ErrModuleUnknown returned when no module can be found for the specified key
	ErrModuleUnknown = errors.New("unknown module")

	// Version of the application
	Version string
)

// A Module is a component that has a specific lifecycle
type Module interface {
	Init(Application) error
	Start(Application) error
	Stop(Application) error
}

// A ModuleKey represents a key for a module.
// Users of this package can define their own keys, this is just the type definition.
type ModuleKey uint16

// Application is an application level context package
// It can be used as a kind of dependency injection container
type Application interface {
	// Get the module at the specified key, thread-safe
	Get(ModuleKey, Module) error

	// Set the module at the specified key, this should be safe across multiple threads
	Set(ModuleKey, Module) error

	// Logger gets the root logger for this application
	Logger() logrus.FieldLogger

	// NewLogger creates a new named logger for this application
	NewLogger(string, map[string]interface{}) logrus.FieldLogger

	// Tracer returns the root
	Tracer() tracing.Tracer

	// NewTracer creates a new named tracer for this application
	NewTracer(string) tracing.Tracer

	// Config returns the viper config for this application
	Config() *viper.Viper

	// Info returns the app info object for this application
	Info() cjm.AppInfo
}

// New application with the specified name, at the specified basepath
func New(name, basePath string) Application {
	// configure version defaults
	version := "dev"
	if Version != "" {
		version = Version
	}

	// configure name defaults
	if name == "" {
		exe, err := osext.Executable()
		if err != nil {
			logrus.Fatalln(err)
		}
		name = exe
	}

	// configure basepath
	if basePath == "" {
		basePath = "/"
	}

	appInfo := cjm.AppInfo{
		Name:     filepath.Base(name),
		BasePath: basePath,
		Version:  version,
		Pid:      os.Getpid(),
	}

	// TODO: read config
	cfg := viper.GetViper()

	rootLogger := logrus.WithFields(logrus.Fields{
		"app":  name,
		"name": "root",
	})

	return &defaultApplication{
		appInfo:    appInfo,
		rootLogger: rootLogger,
		rootTracer: tracing.NewTracer("root", rootLogger.WithField("name", "tracer"), nil),
		config:     cfg,
		registry:   make(map[ModuleKey]reflect.Value, 100),
		regLock:    new(sync.Mutex),
	}
}

type defaultApplication struct {
	appInfo    cjm.AppInfo
	rootLogger logrus.FieldLogger
	rootTracer tracing.Tracer
	config     *viper.Viper

	registry map[ModuleKey]reflect.Value
	regLock  *sync.Mutex
}

// Get the module at the specified key, return a not found error when the module can't be found
func (d *defaultApplication) Get(key ModuleKey, module Module) error {
	mv := reflect.ValueOf(module)
	if mv.Kind() != reflect.Ptr {
		return fmt.Errorf("expected module %T to be a pointer", module)
	}

	d.regLock.Lock()
	mod, ok := d.registry[key]
	if !ok {
		d.regLock.Unlock()
		return ErrModuleUnknown
	}

	av := reflect.Indirect(mv)
	if !mod.Type().AssignableTo(av.Type()) {
		d.regLock.Unlock()
		return fmt.Errorf("can't assign %T to %T", mod.Interface(), av.Interface())
	}

	av.Set(mod)
	d.regLock.Unlock()
	return nil
}

func (d *defaultApplication) Set(key ModuleKey, module Module) error {
	d.regLock.Lock()
	d.registry[key] = reflect.Indirect(reflect.ValueOf(module))
	d.regLock.Unlock()
	return nil
}

func (d *defaultApplication) Logger() logrus.FieldLogger {
	return d.rootLogger
}

func (d *defaultApplication) NewLogger(name string, ctx map[string]interface{}) logrus.FieldLogger {
	if ctx == nil {
		ctx = map[string]interface{}{}
	}
	ctx["name"] = name
	return d.rootLogger.WithFields(logrus.Fields(ctx))
}

func (d *defaultApplication) Tracer() tracing.Tracer {
	return d.rootTracer
}

func (d *defaultApplication) NewTracer(name string) tracing.Tracer {
	return tracing.NewTracer(name, d.rootLogger.WithFields(logrus.Fields{"name": name + "-tracer"}), nil)
}

func (d *defaultApplication) Config() *viper.Viper {
	return d.config
}

func (d *defaultApplication) Info() cjm.AppInfo {
	return d.appInfo
}
