package yamlfe

import (
	"github.com/tobiash/go-make/pkg/mk"
	"gopkg.in/yaml.v3"
	"io"
	"regexp"
	"text/template"
)

type Makefile struct {
	Shell []string      `yaml:"shell"`
	Rules []ruleWrapper `yaml:"rules"`
}

// ruleWrapper is a helper for yaml unmarshalling that wraps different rule types
type ruleWrapper struct {
	rule
}

type rule interface {
	build(f *Makefile) (mk.Rule, error)
}

type regexpRuleRaw struct {
	Pattern       string   `yaml:"pattern"`
	Prerequisites []string `yaml:"prerequisites"`
	Recipe        []string `yaml:"recipe"`
}

func (f *Makefile) Parse(r io.Reader) error {
	return yaml.NewDecoder(r).Decode(f)
}

func (f *Makefile) BuildRules() ([]mk.Rule, error) {
	rs := make([]mk.Rule, len(f.Rules))
	for i := range f.Rules {
		r, err := f.Rules[i].build(f)
		if err != nil {
			return nil, err
		}
		rs[i] = r
	}
	return rs, nil
}

func (r *regexpRuleRaw) build(f *Makefile) (mk.Rule, error) {
	mr, err := regexp.Compile(r.Pattern)
	if err != nil {
		return nil, err
	}
	rec := make([]*template.Template, len(r.Recipe))
	for i, r := range r.Recipe {
		rtpl, err := template.New("").Parse(r)
		if err != nil {
			return nil, err
		}
		rec[i] = rtpl
	}
	ps := make([]*template.Template, len(r.Prerequisites))
	for k, p := range r.Prerequisites {
		ptpl, err := template.New("").Parse(p)
		if err != nil {
			return nil, err
		}
		ps[k] = ptpl
	}
	return &regexRule{
		mkfile:        f,
		regexp:        mr,
		prerequisites: ps,
		recipe:        rec,
	}, nil
}

func (r *ruleWrapper) UnmarshalYAML(value *yaml.Node) error {
	fields := make(map[string]interface{})
	if err := value.Decode(&fields); err != nil {
		return err
	}
	switch fields["type"] {
	default:
		var raw regexpRuleRaw
		if err := value.Decode(&raw); err != nil {
			return err
		}
		r.rule = &raw
	}
	return nil
}
