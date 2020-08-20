package yamlfe

import (
	"bytes"
	"context"
	"github.com/rs/zerolog"
	"github.com/tobiash/go-make/pkg/mk"
	"regexp"
	"text/template"
)

type tplContext struct {
	Target mk.Target
	Matches map[string]string
	Prerequisites []mk.Target
}

type regexRule struct {
	mkfile *Makefile
	regexp *regexp.Regexp
	prerequisites []*template.Template
	recipe []*template.Template
}

type invocation struct {
	rule *regexRule
	target mk.Target
	prereqs []mk.Target
	matches map[string]string
}

func (i *invocation) Prerequisites() []mk.Target {
	return i.prereqs
}

func (i *invocation) Execute(exec mk.Executor, ctx context.Context) error {
	se, ok := exec.(interface{ RunShell(ctx context.Context, cmd string) error })
	if !ok {
		panic("rule needs shell execution")
	}
	log := zerolog.Ctx(ctx)
	tplCtx := &tplContext{
		Target: i.target,
		Prerequisites: i.prereqs,
		Matches: i.matches,
	}
	for k := range i.rule.recipe {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		buf := new(bytes.Buffer)
		if err := i.rule.recipe[k].Execute(buf, &tplCtx); err != nil {
			return err
		}
		log.Info().Str("cmd", buf.String()).Msg("executing recipe")
		if err := se.RunShell(ctx, buf.String()); err != nil {
			return err
		}
	}
	return nil
}

func (r *regexRule) Match(target mk.Target) (mk.MatchQuality, mk.Invocation, error) {
	match := r.regexp.FindStringSubmatch(target.Name())
	if match == nil {
		return mk.NoMatch, nil, nil
	}
	submatches := make(map[string]string)
	for i, n := range r.regexp.SubexpNames() {
		if n != "" {
			submatches[n] = match[i]
		}
	}
	ctx := tplContext{
		Target: target,
		Matches: submatches,
	}
	prereqs := make([]mk.Target, len(r.prerequisites))
	for i, ps := range r.prerequisites {
		buf := new(bytes.Buffer)
		if err := ps.Execute(buf, ctx); err != nil {
			return mk.MatchImplicit, nil, err
		}
		prereqs[i] = &mk.FileTarget{Path: buf.String()}
	}
	return mk.MatchImplicit, &invocation{
		rule:    r,
		target:  target,
		prereqs: prereqs,
		matches: submatches,
	}, nil
}
