package app

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/casualjim/go-app/logging"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

func TestApplication_Constructor(t *testing.T) {
	appi, _ := New("")
	app := appi.(*defaultApplication)

	if assert.NotNil(t, app.appInfo) {
		info := app.appInfo
		assert.NotEmpty(t, info.Version)
		assert.Equal(t, "dev", info.Version)
		assert.NotEmpty(t, info.Name)
		assert.Equal(t, "go-app.test", info.Name)
		assert.Equal(t, "/", info.BasePath)
	}

	Version = "0.1.0"
	appi2, _ := New("the-app")
	app2 := appi2.(*defaultApplication)
	Version = ""

	if assert.NotNil(t, app2.appInfo) {
		info := app2.appInfo
		assert.NotEmpty(t, info.Version)
		assert.Equal(t, "0.1.0", info.Version)
		assert.NotEmpty(t, info.Name)
		assert.Equal(t, "the-app", info.Name)
		assert.Equal(t, "/", info.BasePath)
	}

	assert.NotNil(t, app.rootLogger)
	assert.NotNil(t, app.Tracer())
	assert.NotNil(t, app.Config())
	assert.NotNil(t, app.registry)
	assert.NotNil(t, app.regLock)
}

func TestApplication_InvalidConfigFile(t *testing.T) {
	err := ioutil.WriteFile("config.json", []byte(`{]}`), 0644)
	defer os.Remove("config.json")

	if assert.NoError(t, err) {
		_, err := New("")
		if assert.Error(t, err) {
			t.Log(err)
		}
	}
}

func TestApplication_WatchFile(t *testing.T) {
	err := ioutil.WriteFile("config.json", []byte(`{"name": "some value"}`), 0644)
	defer os.Remove("config.json")

	if assert.NoError(t, err) {
		latch := make(chan struct{})
		reloadCallback = func(_ fsnotify.Event) { latch <- struct{}{} }
		app, err := New("")
		if assert.NoError(t, err) {
			assert.Equal(t, "some value", app.Config().GetString("name"))
			go func() {
				<-time.After(1 * time.Second)
				err := ioutil.WriteFile("config.json", []byte(`{"name": "other value"}`), 0644)
				if err != nil {
					t.Log(err)
				}

			}()
			<-latch
			assert.Equal(t, "other value", app.Config().GetString("name"))
		}
	}
}

func TestApplication_ExeNameFallback(t *testing.T) {
	oldExefn := execName
	defer func() { execName = oldExefn }()

	execName = func() (string, error) { return "app1", nil }
	app1, _ := New("")
	assert.Equal(t, "app1", app1.Info().Name)

	execName = func() (string, error) { return "github.com/some/package/app2", nil }
	app2, _ := New("")
	assert.Equal(t, "app2", app2.Info().Name)

	execName = func() (string, error) { return "", errors.New("expected") }
	_, err := New("")
	assert.EqualError(t, err, "expected")
}

func TestApplication_GetModule(t *testing.T) {
	app, _ := New("GetModuleTest")
	const orig = "original"

	fm := new(firstModule)
	fm.arb = orig
	fm.Init(app)

	var fm2 firstModule
	fm2.arb = "second"
	if assert.NoError(t, app.Get(firstModuleKey, &fm2)) {
		assert.Equal(t, orig, fm2.arb)
		assert.Equal(t, fm.arb, fm2.arb)
	}

	var fm3 firstModule
	assert.EqualError(t, app.Get(ModuleKey(300), &fm3), ErrModuleUnknown.Error())

	var om otherModule
	err := app.Get(firstModuleKey, &om)
	assert.EqualError(t, err, "can't assign app.firstModule to app.otherModule")

	var np someModule
	err2 := app.Get(someModuleKey, np)
	assert.EqualError(t, err2, "expected module app.someModule to be a pointer")
}

func TestApplication_SetModule(t *testing.T) {
	appi, _ := New("SetModuleTest")
	app := appi.(*defaultApplication)

	fm := new(firstModule)
	fm.arb = "original"
	fm.Init(app)

	assert.Contains(t, app.registry, firstModuleKey)
}

func TestApplication_Logger(t *testing.T) {
	app, _ := New("LoggerTest")
	assert.NotNil(t, app.Logger())
	assert.Implements(t, (*logging.Logger)(nil), app.Logger())

	child := app.NewLogger("appModule", logrus.Fields{"extra": "data"})
	if assert.NotNil(t, child) && assert.Implements(t, (*logging.Logger)(nil), child) {
		data := child.(logging.Logger).Fields()
		assert.Equal(t, "appModule", data["module"])
		assert.Equal(t, "data", data["extra"])
		assert.Equal(t, "LoggerTest", data["app"])
	}
}

const (
	firstModuleKey ModuleKey = "firstModule"
	otherModuleKey           = "otherModule"
	someModuleKey            = "someModule"
)

type otherModule struct {
	blah int
}

func (f *otherModule) Init(app Application) error {
	app.Set(otherModuleKey, f)
	return nil
}

func (f *otherModule) Start(_ Application) error {
	return nil
}

func (f *otherModule) Stop(_ Application) error {
	return nil
}

type firstModule struct {
	arb string
}

func (f *firstModule) Init(app Application) error {
	app.Set(firstModuleKey, f)
	return nil
}

func (f *firstModule) Start(_ Application) error {
	return nil
}

func (f *firstModule) Stop(_ Application) error {
	return nil
}

type someModule struct {
	arb string
}

func (f someModule) Init(app Application) error {
	app.Set(someModuleKey, f)
	return nil
}

func (f someModule) Start(_ Application) error {
	return nil
}

func (f someModule) Stop(_ Application) error {
	return nil
}
