package yamlfe

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobiash/go-make/pkg/mk"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestRuleMatch(t *testing.T) {
	testYaml := `
pattern: "foo.yaml"
prerequisites:
- "bar.yaml"
recipe:
- "touch {{ .Target.Path }}"
`
	var r ruleWrapper
	assert.NoError(t, yaml.NewDecoder(strings.NewReader(testYaml)).Decode(&r))
}

func TestRulePatternMatch(t *testing.T) {
	testYaml := `
pattern: "(?P<first>[^/]+)/bla.yaml"
prerequisites:
- "{{ .Matches.first }}.json"
recipe:
- "echo {{ index .Matches \"first\" }}; touch {{ .Target.Path }}"
`
	var r ruleWrapper
	require.NoError(t, yaml.NewDecoder(strings.NewReader(testYaml)).Decode(&r))
	rule, err := r.build(&Makefile{
		Rules: nil,
	})
	require.NoError(t, err)
	q, inv, err := rule.Match(&mk.FileTarget{
		Path: "foo/bla.yaml",
	})
	require.NoError(t, err)
	assert.Equal(t, mk.MatchImplicit, q)
	assert.Len(t, inv.Prerequisites(), 1)
	assert.Equal(t, &mk.FileTarget{Path: "foo.json"}, inv.Prerequisites()[0])
}
