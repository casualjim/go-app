package logging

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLogging_RegisterWriter(t *testing.T) {
	assert.NotContains(t, knownWriters, "another")
	RegisterWriter("another", func(_ *viper.Viper) io.Writer { return nil })
	assert.Contains(t, knownWriters, "another")
}

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

	// fallback to default
	assert.Equal(t, os.Stderr, parseWriter(nil))
	assert.Equal(t, os.Stderr, parseWriter(tcn("not-there")))
}

func TestLogging_KnownWriters(t *testing.T) {
	kw := KnownWriters()
	assert.Len(t, kw, len(knownWriters))
}
