package mk

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileCheck(t *testing.T) {
	d, err := ioutil.TempDir("", "go-make")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(d) }()
	ft := &FileTarget{Dir: d, Path: "test.yaml"}
	status, err := ft.Check("")
	require.NoError(t, err)
	assert.False(t, status.Exists)
	assert.False(t, status.UpToDate)
	assert.Equal(t, "", status.CurrentDigest)
	require.NoError(t, ioutil.WriteFile(filepath.Join(d, ft.Path), []byte("Hello World"), 0700))
	status, err = ft.Check("")
	require.NoError(t, err)
	assert.True(t, status.Exists)
	assert.True(t, status.UpToDate)
	digest := status.CurrentDigest
	status, err = ft.Check(digest)
	require.NoError(t, err)
	assert.True(t, status.Exists)
	assert.True(t, status.UpToDate)
	assert.Equal(t, digest, status.CurrentDigest)
	status, err = ft.Check(digest + "foo")
	require.NoError(t, err)
	assert.True(t, status.Exists)
	assert.False(t, status.UpToDate)
	assert.Equal(t, digest, status.CurrentDigest)
}
