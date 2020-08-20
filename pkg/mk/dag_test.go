package mk

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type ttarget struct {
	name string
}

func (tt *ttarget) Check(digest string) (TargetStatus, error) {
	panic("implement me")
}

func (tt *ttarget) Name() string {
	return tt.name
}

func (tt *ttarget) Equal(t Target) bool {
	return t.Name() == tt.name
}

func TestDAGSimpleWalk(t *testing.T) {
	dag := NewDAG()
	a := ttarget{"a"}
	b := ttarget{"b"}
	c := ttarget{"c"}
	d := ttarget{"d"}

	dag.AddTarget(&a, []Target{&b, &c})
	dag.AddTarget(&b, []Target{&d})
	dag.AddTarget(&c, nil)
	dag.AddTarget(&d, nil)

	l := sync.Mutex{}
	var rs []Target
	assert.NoError(t, dag.WalkUp(context.TODO(), 4, func(ctx context.Context, target Target) error {
		l.Lock()
		defer l.Unlock()
		rs = append(rs, target)
		return nil
	}))
	assert.True(t, targetListEqual(rs, []Target{&d, &c, &b, &a}) || targetListEqual(rs, []Target{&c, &d, &b, &a}))
}

func targetListEqual(a, b []Target) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name() != b[i].Name() {
			return false
		}
	}
	return true
}
