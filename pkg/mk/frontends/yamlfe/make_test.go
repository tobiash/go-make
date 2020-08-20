package yamlfe

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestParseMakeFile(t *testing.T) {
	testYaml := `
shell: [ "/usr/bin/env", "bash", "-c" ]
rules:
  - type: regexp
    pattern: "foo.bar"
    prerequisites:
      - "bar"
    recipe:
      - "touch {{ .Target }}"
`
	var mkFile Makefile
	require.NoError(t, yaml.NewDecoder(strings.NewReader(testYaml)).Decode(&mkFile))
	assert.Equal(t, []string { "/usr/bin/env", "bash", "-c"}, mkFile.Shell)
	assert.Len(t, mkFile.Rules, 1)
	rules, err := mkFile.BuildRules()
	require.NoError(t, err)
	assert.NotNil(t, rules)
	require.Len(t, rules, 1)
	assert.IsType(t, &regexRule{}, rules[0])
}
