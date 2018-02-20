package xc

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(td, ".gopass.yml")))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	passphrase := "test"
	xc, err := New(td, &fakeAgent{passphrase})
	assert.NoError(t, err)
	assert.NotNil(t, xc)
	assert.NoError(t, xc.Initialized(ctx))
	assert.Equal(t, "xc", xc.Name())
	assert.Equal(t, semver.Version{Patch: 1}, xc.Version(ctx))
	assert.Equal(t, "xc", xc.Ext())
	assert.Equal(t, ".xc-ids", xc.IDFile())
}
