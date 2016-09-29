package app

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/casualjim/go-app/logging"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/xordataexchange/crypt/backend/consul"
	"github.com/xordataexchange/crypt/backend/etcd"
	"github.com/xordataexchange/crypt/encoding/secconf"
)

const (
	conjson = `{
  "name": "go-app.test",
  "location": "a wonderful, magical place among the stars",
  "count": 1
}`
	conjson2 = `{
  "name": "go-app.test",
  "location": "a wonderful, magical place among the stars",
  "count": 3
}`
	conyaml2 = `---
name: go-app.test
location: "a wonderful, magical place among the stars"
count: 3
`
)

func TestApplication_RemoteErrors(t *testing.T) {
	defer os.Unsetenv("CONFIG_REMOTE_URL")
	defer os.Unsetenv("CONFIG_KEYRING")

	// invalid url
	os.Setenv("CONFIG_REMOTE_URL", "etcd://[/")
	_, err := New("")
	assert.Error(t, err)

	// invalid scheme
	os.Setenv("CONFIG_REMOTE_URL", "zookeeper://127.0.0.1:2379/etcdenc/config.json")
	assert.EqualError(t, addViperRemoteConfig(viper.New()), "Unsupported Remote Provider Type \"zookeeper\"")
	os.Setenv("CONFIG_KEYRING", ".pubring.gpg")
	assert.EqualError(t, addViperRemoteConfig(viper.New()), "Unsupported Remote Provider Type \"zookeeper\"")

	// invalid type
	os.Setenv("CONFIG_REMOTE_URL", "etcd://127.0.0.1:2379/etcdenc/config.unknown")
	assert.EqualError(t, addViperRemoteConfig(viper.New()), "config is invalid as unknown")
}

func encrypt(d []byte) ([]byte, error) {
	kr, err := os.Open(".secring.gpg")
	if err != nil {
		return nil, err
	}
	defer kr.Close()

	return secconf.Encode(d, kr)
}

func TestApplication_EtcdEncrypted(t *testing.T) {
	if _, err := os.Stat(".secring.gpg"); os.IsNotExist(err) {
		t.Skip("skipping, no keyring.")
	}
	defer os.Unsetenv("CONFIG_KEYRING")
	defer os.Unsetenv("CONFIG_REMOTE_URL")
	os.Setenv("CONFIG_KEYRING", ".secring.gpg")
	os.Setenv("CONFIG_REMOTE_URL", "etcd://127.0.0.1:2379/etcdenc/config.json")

	etcdc, err := etcd.New([]string{"http://127.0.0.1:2379"})
	if assert.NoError(t, err) {
		encrypted, err := encrypt([]byte(conjson))
		if assert.NoError(t, err) {
			if assert.NoError(t, etcdc.Set("/etcdenc/config.json", encrypted)) {
				b, err := etcdc.Get("/etcdenc/config.json")
				if assert.NoError(t, err) {
					assert.Equal(t, string(encrypted), string(b))

					v := viper.New()
					if assert.NoError(t, addViperRemoteConfig(v)) {
						assert.Equal(t, "go-app.test", v.GetString("name"))
						assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
						assert.Equal(t, 1, v.GetInt("count"))
					}
				}
			}
		}
	}
}

func TestApplication_EtcdUnencrypted(t *testing.T) {
	defer os.Unsetenv("CONFIG_REMOTE_URL")
	os.Unsetenv("CONFIG_KEYRING")
	os.Setenv("CONFIG_REMOTE_URL", "etcd://127.0.0.1:2379/etcdplain/config.json")

	etcdc, err := etcd.New([]string{"http://127.0.0.1:2379"})
	if assert.NoError(t, err) {
		err = etcdc.Set("/etcdplain/config.json", []byte(conjson))
		if err != nil {
			t.Skip("skipping, no etcd.")
		} else {
			b, err := etcdc.Get("/etcdplain/config.json")
			if assert.NoError(t, err) {
				assert.Equal(t, conjson, string(b))

				v := viper.New()
				if assert.NoError(t, addViperRemoteConfig(v)) {
					assert.Equal(t, "go-app.test", v.GetString("name"))
					assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
					assert.Equal(t, 1, v.GetInt("count"))
				}
			}
		}

		os.Setenv("CONFIG_REMOTE_URL", "etcd://127.0.0.1:2379/etcdplain/config")
		if assert.NoError(t, etcdc.Set("/etcdplain/config", []byte(conjson))) {
			b, err := etcdc.Get("/etcdplain/config")
			if assert.NoError(t, err) {
				assert.Equal(t, conjson, string(b))

				v := viper.New()
				if assert.NoError(t, addViperRemoteConfig(v)) {
					assert.Equal(t, "go-app.test", v.GetString("name"))
					assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
					assert.Equal(t, 1, v.GetInt("count"))
				}
			}
		}
	}
}

func TestApplication_WatchEtcd(t *testing.T) {
	if _, err := os.Stat(".secring.gpg"); os.IsNotExist(err) {
		t.Skip("skipping, no keyring.")
	}
	defer os.Unsetenv("CONFIG_KEYRING")
	defer os.Unsetenv("CONFIG_REMOTE_URL")
	os.Setenv("CONFIG_KEYRING", ".secring.gpg")
	os.Setenv("CONFIG_REMOTE_URL", "etcd://127.0.0.1:2379/etcdenc/config.json")

	etcdc, err := etcd.New([]string{"http://127.0.0.1:2379"})
	if assert.NoError(t, err) {
		encrypted, err := encrypt([]byte(conjson))
		if assert.NoError(t, err) {
			if assert.NoError(t, etcdc.Set("/etcdenc/config.json", encrypted)) {
				b, err := etcdc.Get("/etcdenc/config.json")
				if assert.NoError(t, err) {
					assert.Equal(t, string(encrypted), string(b))

					latch := make(chan struct{})
					app, err := newWithCallback("", "", func(_ fsnotify.Event) { latch <- struct{}{} })
					if assert.NoError(t, err) {
						v := app.Config()
						assert.Equal(t, "go-app.test", v.GetString("name"))
						assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
						assert.Equal(t, 1, v.GetInt("count"))
						encrypted2, err := encrypt([]byte(conyaml2))

						if assert.NoError(t, err) {
							go func() {
								<-time.After(1 * time.Second)
								err = etcdc.Set("/etcdenc/config.json", encrypted2)
								if err != nil {
									t.Log(err)
								}
							}()

							select {
							case <-latch:
								b, err := etcdc.Get("/etcdenc/config.json")
								if assert.NoError(t, err) {
									assert.Equal(t, string(encrypted2), string(b))

									assert.Equal(t, "go-app.test", v.GetString("name"))
									assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
									assert.Equal(t, 3, v.GetInt("count"))
								}
							case <-time.After(2 * time.Second):
								t.Log("watch timed out, expected")
							}
						}
					}
				}
			}
		}
	}
}

func TestApplication_WatchEtcdError(t *testing.T) {
	if _, err := os.Stat(".secring.gpg"); os.IsNotExist(err) {
		t.Skip("skipping, no keyring.")
	}
	defer os.Unsetenv("CONFIG_KEYRING")
	defer os.Unsetenv("CONFIG_REMOTE_URL")
	os.Setenv("CONFIG_KEYRING", ".secring.gpg")
	os.Setenv("CONFIG_REMOTE_URL", "etcd://127.0.0.1:2379/etcdenc/config.json")

	etcdc, err := etcd.New([]string{"http://127.0.0.1:2379"})
	if assert.NoError(t, err) {
		encrypted, err := encrypt([]byte(conjson))
		if assert.NoError(t, err) {
			if assert.NoError(t, etcdc.Set("/etcdenc/config.json", encrypted)) {
				b, err := etcdc.Get("/etcdenc/config.json")
				if assert.NoError(t, err) {
					assert.Equal(t, string(encrypted), string(b))

					latch := make(chan struct{})
					app, err := newWithCallback("", "", func(_ fsnotify.Event) { latch <- struct{}{} })
					if assert.NoError(t, err) {
						v := app.Config()
						assert.Equal(t, "go-app.test", v.GetString("name"))
						assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
						assert.Equal(t, 1, v.GetInt("count"))
						encrypted2, err := encrypt([]byte(conjson2))

						if assert.NoError(t, err) {
							go func() {
								<-time.After(1 * time.Second)
								err = etcdc.Set("/etcdenc/config.json", encrypted2)
								if err != nil {
									t.Log(err)
								}
							}()

							select {
							case <-latch:
								b, err := etcdc.Get("/etcdenc/config.json")
								if assert.NoError(t, err) {
									assert.Equal(t, string(encrypted2), string(b))

									assert.Equal(t, "go-app.test", v.GetString("name"))
									assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
									assert.Equal(t, 3, v.GetInt("count"))
								}
							case <-time.After(10 * time.Second):
								t.Error("watch timed out")
							}
						}
					}
				}
			}
		}
	}
}

func TestApplication_ConsulEncrypted(t *testing.T) {
	if _, err := os.Stat(".secring.gpg"); os.IsNotExist(err) {
		t.Skip("skipping, no keyring.")
	}
	defer os.Unsetenv("CONFIG_KEYRING")
	defer os.Unsetenv("CONFIG_REMOTE_URL")
	os.Setenv("CONFIG_KEYRING", ".secring.gpg")
	os.Setenv("CONFIG_REMOTE_URL", "consul://127.0.0.1:8500/consulenc/config.json")

	consulc, err := consul.New([]string{"127.0.0.1:8500"})
	if assert.NoError(t, err) {
		encrypted, err := encrypt([]byte(conjson))
		if assert.NoError(t, err) {
			if assert.NoError(t, consulc.Set("/consulenc/config.json", encrypted)) {
				b, err := consulc.Get("/consulenc/config.json")
				if assert.NoError(t, err) {
					assert.Equal(t, string(encrypted), string(b))

					v := viper.New()
					if assert.NoError(t, addViperRemoteConfig(v)) {
						assert.Equal(t, "go-app.test", v.GetString("name"))
						assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
						assert.Equal(t, 1, v.GetInt("count"))
					}
				}
			}
		}
	}

}

func TestApplication_ConsulUnencrypted(t *testing.T) {
	defer os.Unsetenv("CONFIG_REMOTE_URL")
	os.Unsetenv("CONFIG_KEYRING")
	os.Setenv("CONFIG_REMOTE_URL", "consul://127.0.0.1:8500/consulplain/config.json")

	consulc, err := consul.New([]string{"127.0.0.1:8500"})
	if assert.NoError(t, err) {
		err = consulc.Set("/consulplain/config.json", []byte(conjson))
		if err != nil {
			t.Skip("skipping, no consul.")
		} else {
			b, err := consulc.Get("/consulplain/config.json")
			if assert.NoError(t, err) {
				assert.Equal(t, conjson, string(b))

				v := viper.New()
				if assert.NoError(t, addViperRemoteConfig(v)) {
					assert.Equal(t, "go-app.test", v.GetString("name"))
					assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
					assert.Equal(t, 1, v.GetInt("count"))
				}
			}
		}
	}
}

func TestApplication_WatchConsul(t *testing.T) {
	t.Skip("until fixed")
	defer os.Unsetenv("CONFIG_KEYRING")
	defer os.Unsetenv("CONFIG_REMOTE_URL")
	os.Setenv("CONFIG_KEYRING", ".secring.gpg")
	os.Setenv("CONFIG_REMOTE_URL", "consul://127.0.0.1:8500/consulenc/config.json")

	consulc, err := consul.New([]string{"127.0.0.1:8500"})
	if assert.NoError(t, err) {
		encrypted, err := encrypt([]byte(conjson))
		if assert.NoError(t, err) {
			if assert.NoError(t, consulc.Set("/consulenc/config.json", encrypted)) {
				b, err := consulc.Get("/consulenc/config.json")
				if assert.NoError(t, err) {
					assert.Equal(t, string(encrypted), string(b))

					latch := make(chan struct{})
					app, err := newWithCallback("", "", func(_ fsnotify.Event) { latch <- struct{}{} })
					if assert.NoError(t, err) {
						v := app.Config()
						assert.Equal(t, "go-app.test", v.GetString("name"))
						assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
						assert.Equal(t, 1, v.GetInt("count"))

						encrypted2, err := encrypt([]byte(conjson2))
						if assert.NoError(t, err) {
							go func() {
								<-time.After(1 * time.Second)
								err = consulc.Set("/consulenc/config.json", encrypted2)
								if err != nil {
									t.Log(err)
								}
							}()

							select {
							case <-latch:
								b, err := consulc.Get("/consulenc/config.json")
								if assert.NoError(t, err) {
									assert.Equal(t, string(encrypted2), string(b))

									assert.Equal(t, "go-app.test", v.GetString("name"))
									assert.Equal(t, "a wonderful, magical place among the stars", v.Get("location"))
									assert.Equal(t, 3, v.GetInt("count"))
								}
							case <-time.After(10 * time.Second):
								t.Error("watch timed out")
							}
						}
					}
				}
			}
		}
	}
}

func TestApplication_Constructor(t *testing.T) {
	appi, err := New("")
	if assert.NoError(t, err) {
		app := appi.(*defaultApplication)

		if assert.NotNil(t, app.appInfo) {
			info := app.appInfo
			assert.NotEmpty(t, info.Version)
			assert.Equal(t, "dev", info.Version)
			assert.NotEmpty(t, info.Name)
			assert.Equal(t, "go-app.test", info.Name)
			assert.Equal(t, "/", info.BasePath)
		}

		assert.NotNil(t, app.Logger())
		assert.NotNil(t, app.Tracer())
		assert.NotNil(t, app.Config())
		assert.NotNil(t, app.registry)
		assert.NotNil(t, app.regLock)

	}

	Version = "0.1.0"
	appi2, err := New("the-app")
	if assert.NoError(t, err) {
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
	}
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
		app, err := newWithCallback("", "", func(_ fsnotify.Event) { latch <- struct{}{} })
		if assert.NoError(t, err) {
			app.Add(MakeModule(Reload(func(_ Application) error { return errors.New("expected") })))
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

func TestApplication_EnvConfigPath(t *testing.T) {
	oldCP := os.Getenv("CONFIG_PATH")
	defer os.Setenv("CONFIG_PATH", oldCP)

	dir, err := ioutil.TempDir("", "go-app")
	if assert.NoError(t, err) {
		cpath := filepath.Join(dir, "configs")
		tpar := filepath.Join(os.TempDir(), "configs")
		os.Setenv("CONFIG_PATH", tpar+":"+cpath)

		// save in last dir
		if assert.NoError(t, os.MkdirAll(cpath, 0755)) {
			defer os.RemoveAll(cpath)
			fpath := filepath.Join(cpath, "config.json")
			content := []byte(`{"name":"some-config"}`)
			if assert.NoError(t, ioutil.WriteFile(fpath, content, 0644)) {
				v, err := createViper("test3", "")
				if assert.NoError(t, err) {
					assert.Equal(t, "some-config", v.GetString("name"))
				}
			}
		}

		// save in first dir
		if assert.NoError(t, os.MkdirAll(tpar, 0755)) {
			defer os.RemoveAll(tpar)
			fpath := filepath.Join(tpar, "config.json")
			content := []byte(`{"name":"other-config"}`)
			if assert.NoError(t, ioutil.WriteFile(fpath, content, 0644)) {
				v, err := createViper("test4", "")
				if assert.NoError(t, err) {
					assert.Equal(t, "other-config", v.GetString("name"))
				}
			}
		}
	}
}

func TestApplication_ConstructorWithConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "go-app")
	if assert.NoError(t, err) {
		fpath := filepath.Join(dir, "myconfig.json")
		content := []byte(`{"name":"other-config"}`)
		if assert.NoError(t, ioutil.WriteFile(fpath, content, 0644)) {
			appi, err := NewWithConfig("", fpath)
			if assert.NoError(t, err) {
				app := appi.(*defaultApplication)
				v := app.Config()
				assert.Equal(t, "other-config", v.GetString("name"))
			}
		}
		// should work with or without file extension
		fpath = filepath.Join(dir, "myconfig")
		content = []byte(`{"name":"other-config"}`)
		if assert.NoError(t, ioutil.WriteFile(fpath, content, 0644)) {
			appi, err := NewWithConfig("", fpath)
			if assert.NoError(t, err) {
				app := appi.(*defaultApplication)
				v := app.Config()
				assert.Equal(t, "other-config", v.GetString("name"))
			}
		}
	}
}

func TestApplication_ConstructorWithInvalidConfig(t *testing.T) {
	_, err := NewWithConfig("", "/some/nonexistent/path.json")
	assert.Error(t, err)
	dir, err := ioutil.TempDir("", "go-app")
	if assert.NoError(t, err) {
		fpath := filepath.Join(dir, "myconfig.json")
		content := []byte(`{"name":"other-config"}`)
		if assert.NoError(t, ioutil.WriteFile(fpath, content, 0644)) {
			appi, err := NewWithConfig("", fpath)
			if assert.NoError(t, err) {
				app := appi.(*defaultApplication)
				v := app.Config()
				assert.Equal(t, "other-config", v.GetString("name"))
			}
		}
		// invalid json
		fpath = filepath.Join(dir, "myconfig.json")
		content = []byte(`{name:"other-config"}`)
		if assert.NoError(t, ioutil.WriteFile(fpath, content, 0644)) {
			_, err = NewWithConfig("", fpath)
			assert.Error(t, err)
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

func TestApplication_GetOKModule(t *testing.T) {
	app, _ := New("GetOKModuleTest")
	const orig = "original"

	fm := new(firstModule)
	fm.arb = orig
	fm.Init(app)

	var fm2 firstModule
	fm2.arb = "second"
	iface, ok := app.GetOK(firstModuleKey)
	if assert.True(t, ok) {
		fm2 := iface.(*firstModule)
		assert.Equal(t, orig, fm2.arb)
		assert.Equal(t, fm.arb, fm2.arb)
	}

	fm3, ok := app.GetOK(Key("300"))
	assert.False(t, ok)
	assert.Nil(t, fm3)
}

func TestApplication_GetModule(t *testing.T) {
	app, _ := New("GetModuleTest")
	const orig = "original"

	fm := new(firstModule)
	fm.arb = orig
	fm.Init(app)

	var fm2 firstModule
	fm2.arb = "second"
	iface := app.Get(firstModuleKey)
	if assert.NotNil(t, iface) {
		fm2 := iface.(*firstModule)
		assert.Equal(t, orig, fm2.arb)
		assert.Equal(t, fm.arb, fm2.arb)
	}

	fm3 := app.Get(Key("300"))
	assert.Nil(t, fm3)
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
	firstModuleKey Key = "firstModule"
	otherModuleKey     = "otherModule"
	someModuleKey      = "someModule"
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
