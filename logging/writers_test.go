package logging

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func tcn(n string) *viper.Viper {
	v := viper.New()
	v.Set("name", n)
	return v
}

var defaultWriters = []struct {
	Name     string
	Expected io.Writer
}{
	{"discard", ioutil.Discard},
	{"stdout", os.Stdout},
	{"stderr", os.Stderr},
}

type SomeWriter struct {
	Padding int32
	Reverse bool
}

func (s *SomeWriter) Write(p []byte) (n int, err error) {
	return
}

func TestLogging_RegisterWriter(t *testing.T) {
	assert.NotContains(t, knownWriters, "another")
	RegisterWriter("another", func(_ *viper.Viper) io.Writer { return nil })
	assert.Contains(t, knownWriters, "another")
}

func TestLogging_DefaultWriters(t *testing.T) {
	for _, sut := range defaultWriters {
		assert.Contains(t, knownWriters, sut.Name)
		assert.Equal(t, sut.Expected, knownWriters[sut.Name](nil))
	}
}

func TestLogging_ParseWriter(t *testing.T) {
	for _, sut := range defaultWriters {
		assert.Equal(t, sut.Expected, parseWriter(tcn(sut.Name)), "failed to create writer for %s", sut.Name)
	}

	// config is a string value
	v := viper.New()
	v.Set("writer", "stderr")
	l := newNamedLogger("root", nil, v).(*defaultLogger)
	assert.Equal(t, l.Entry.Logger.Out, os.Stderr)

	// fallback to default
	assert.Equal(t, os.Stderr, parseWriter(nil))
	assert.Equal(t, os.Stderr, parseWriter(tcn("not-there")))

	// complexer writer with config
	RegisterWriter("some_writer", func(c *viper.Viper) io.Writer {
		var w SomeWriter
		if err := c.Unmarshal(&w); err != nil {
			panic(err)
		}
		return &w
	})

	cc := viper.New()
	cc.Set("name", "some_writer")
	cc.Set("padding", 10)
	cc.Set("reverse", true)
	ww := parseWriter(cc)

	assert.IsType(t, &SomeWriter{}, ww)
	sw := ww.(*SomeWriter)
	assert.EqualValues(t, 10, sw.Padding)
	assert.True(t, sw.Reverse)
}

func TestLogging_KnownWriters(t *testing.T) {
	kw := KnownWriters()
	assert.Len(t, kw, len(knownWriters))
}
