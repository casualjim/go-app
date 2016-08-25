package app

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/casualjim/go-app/logging"
	"github.com/stretchr/testify/assert"
)

func TestApplication_Constructor(t *testing.T) {
	app := New("", "").(*defaultApplication)

	if assert.NotNil(t, app.appInfo) {
		info := app.appInfo
		assert.NotEmpty(t, info.Version)
		assert.Equal(t, "dev", info.Version)
		assert.NotEmpty(t, info.Name)
		assert.Equal(t, "go-app.test", info.Name)
		assert.Equal(t, "/", info.BasePath)
	}

	Version = "0.1.0"
	app2 := New("the-app", "/v1").(*defaultApplication)
	Version = ""

	if assert.NotNil(t, app2.appInfo) {
		info := app2.appInfo
		assert.NotEmpty(t, info.Version)
		assert.Equal(t, "0.1.0", info.Version)
		assert.NotEmpty(t, info.Name)
		assert.Equal(t, "the-app", info.Name)
		assert.Equal(t, "/v1", info.BasePath)
	}

	assert.NotNil(t, app.rootLogger)
	assert.NotNil(t, app.rootTracer)
	assert.NotNil(t, app.config)
	assert.NotNil(t, app.registry)
	assert.NotNil(t, app.regLock)
}

func TestApplication_GetModule(t *testing.T) {
	app := New("GetModuleTest", "/")
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

	var om otherModule
	err := app.Get(firstModuleKey, &om)
	assert.EqualError(t, err, "can't assign app.firstModule to app.otherModule")

	var np someModule
	err2 := app.Get(someModuleKey, np)
	assert.EqualError(t, err2, "expected module app.someModule to be a pointer")
}

func TestApplication_SetModule(t *testing.T) {
	app := New("SetModuleTest", "/").(*defaultApplication)

	fm := new(firstModule)
	fm.arb = "original"
	fm.Init(app)

	assert.Contains(t, app.registry, firstModuleKey)
}

func TestApplication_Logger(t *testing.T) {
	app := New("LoggerTest", "")
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
	firstModuleKey ModuleKey = iota
	otherModuleKey
	someModuleKey
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
