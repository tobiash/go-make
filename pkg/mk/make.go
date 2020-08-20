package mk

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/mod/sumdb/storage"
	"runtime"
)

type MatchQuality int

const (
	NoMatch MatchQuality = iota
	MatchExplicit
	MatchImplicit
)

var ErrTargetNotExists = fmt.Errorf("target does not exist")
var ErrNoRule = fmt.Errorf("no rule to make target")

type Rule interface {
	// Match checks if the rule matches the target and with which prerequisites
	Match(target Target) (match MatchQuality, inv Invocation, err error)
}

type Invocation interface {
	Prerequisites() []Target
	Execute(exec Executor, ctx context.Context) error
}

type Executor interface {
}

type Make struct {
	Sum      storage.Storage
	Rules    []Rule
	nWorkers int
}

type TargetStatus struct {
	UpToDate      bool
	Exists        bool
	CurrentDigest string
}

// Target must be comparable
type Target interface {
	// Name is a unique name of the target, even across target types, e.g. a URL
	Name() string

	// Update check gets called with the comparison digest to determine if the target is up-to-date
	// a new digest value to be recorded can be returned in either non-error case
	Check(digest string) (TargetStatus, error)
}

// Make makes a target
func (m *Make) Make(executor Executor, ctx context.Context, targets ...Target) error {
	dag, rules, err := m.dag(zerolog.Ctx(ctx), targets...)
	if err != nil {
		return err
	}
	nWorkers := m.nWorkers
	if nWorkers == 0 {
		nWorkers = runtime.NumCPU()
	}
	return m.Sum.ReadWrite(ctx, func(ctx context.Context, sumTr storage.Transaction) error {
		return dag.WalkUp(ctx, nWorkers, func(ctx context.Context, target Target) error {
			log := zerolog.Ctx(ctx).With().Str("target", target.Name()).Logger()
			digest, err := sumTr.ReadValue(ctx, target.Name())
			if err != nil {
				return errors.Wrapf(err, "error checking previous digest of target '%s'", target.Name())
			}
			status, err := target.Check(digest)
			if err != nil {
				return errors.Wrapf(err, "error checking status of target '%s'", target.Name())
			}
			rule, ruleExists := rules[target]
			switch {
			case !ruleExists && !status.Exists:
				return errors.Wrapf(ErrNoRule, "error making target '%s'", target.Name())
			case status.UpToDate:
				log.Debug().Msg("target is up-to-date")
			default:
				if err := rule.Execute(executor, ctx); err != nil {
					return err
				}
				status, err = target.Check(digest)
				if err != nil {
					return errors.Wrapf(err, "error checking status of target '%s' post-exec", target.Name())
				}
				if err = sumTr.BufferWrites([]storage.Write{{
					Key:   target.Name(),
					Value: status.CurrentDigest,
				}}); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

// dag computes the DAG for the given targets
func (m *Make) dag(log *zerolog.Logger, targets ...Target) (*DAG, map[Target]Invocation, error) {
	dag := NewDAG()
	dag.Logger = log
	next := make([]Target, 0, len(targets))
	next = append(next, targets...)
	rules := make(map[Target]Invocation)

	for len(next) > 0 {
		u := next[0]
		next = next[1:]
		r, inv, err := m.ruleFor(u)
		if err != nil {
			return nil, nil, err
		}
		if r != nil {
			rules[u] = inv
			prereq := inv.Prerequisites()
			for p := range prereq {
				if _, ok := rules[prereq[p]]; !ok {
					next = append(next, prereq[p])
				}
			}
			dag.AddTarget(u, prereq)
		} else {
			dag.AddTarget(u, nil)
		}
	}
	return dag, rules, nil
}

// rileFor searches for the first rule that "best" matches a target
func (m *Make) ruleFor(t Target) (Rule, Invocation, error) {
	var matchQ MatchQuality
	var rule Rule
	var invocation Invocation
	for r := range m.Rules {
		if match, inv, err := m.Rules[r].Match(t); err != nil {
			return nil, nil, err
		} else if match != NoMatch {
			if match > matchQ {
				matchQ = match
				rule = m.Rules[r]
				invocation = inv
			}
		}
	}
	if matchQ != NoMatch {
		return rule, invocation, nil
	}
	return nil, nil, nil
}
