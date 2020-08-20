package yamlfe

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobiash/go-make/pkg/mk"
	"github.com/tobiash/go-make/pkg/mk/shell"
	"golang.org/x/mod/sumdb/storage"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationShellMake(t *testing.T) {
	d, err := ioutil.TempDir("", "go-make")
	require.NoError(t, err)
	ctx := log.Logger.WithContext(context.TODO())
	defer func() { _ = os.RemoveAll(d) }()
	se := &shell.ShellExecutor{
		Dir: d,
	}
	mkFileYaml := `
rules:
- pattern: "a.(?P<ext>.+)"
  prerequisites:
  - "b.{{ .Matches.ext }}"
  - "c.{{ .Matches.ext }}"
  recipe:
  - "touch {{ .Target.Path }}"
  - "{{range .Prerequisites }}cat {{ .Path }} >> {{ $.Target.Path }};{{ end }}"
- pattern: "b.(?P<ext>.+)"
  recipe:
  - "echo \"b\" > {{ .Target.Path }}"
- pattern: "c.(?P<ext>.+)"
  recipe:
  - "echo \"c\" > {{ .Target.Path }}"
`
	mkFile := &Makefile{}
	require.NoError(t, mkFile.Parse(strings.NewReader(mkFileYaml)))
	rules, err := mkFile.BuildRules()
	require.NoError(t, err)
	assert.NoError(t, (&mk.Make{Rules: rules, Sum: &storage.Mem{}}).Make(se, ctx, &mk.FileTarget{Path: "a.foo"}))
	assert.FileExists(t, filepath.Join(d, "a.foo"))
	assert.FileExists(t, filepath.Join(d, "c.foo"))
	assert.FileExists(t, filepath.Join(d, "b.foo"))
	aContent, err := ioutil.ReadFile(filepath.Join(d, "a.foo"))
	require.NoError(t, err)
	assert.Equal(t, "b\nc\n", string(aContent))
}
