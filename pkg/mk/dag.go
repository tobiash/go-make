package mk

import (
	"context"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type DAG struct {
	graph  map[Target]map[Target]bool
	Logger *zerolog.Logger
}

func NewDAG() *DAG {
	return &DAG{map[Target]map[Target]bool{}, nil}
}

func (g *DAG) AddTarget(t Target, prereqs []Target) {
	if g.graph[t] == nil {
		g.graph[t] = map[Target]bool{}
	}

	for _, p := range prereqs {
		if g.graph[p] == nil {
			g.graph[p] = map[Target]bool{}
		}
		g.graph[t][p] = true
	}
}

func (g *DAG) reverse() *DAG {
	n := NewDAG()
	n.Logger = g.Logger
	for target, prereq := range g.graph {
		for p := range prereq {
			n.AddTarget(p, []Target{target})
		}
		if len(prereq) == 0 {
			n.AddTarget(target, nil)
		}
	}
	return n
}

func (g *DAG) TopographicalSort() []Target {
	var linearOrder []Target

	inDegree := map[Target]int{}
	for n := range g.graph {
		inDegree[n] = 0
	}

	for _, adjacent := range g.graph {
		for v := range adjacent {
			inDegree[v]++
		}
	}

	var next []Target
	for u, v := range inDegree {
		if v != 0 {
			continue
		}
		next = append(next, u)
	}

	for len(next) > 0 {
		u := next[0]
		next = next[1:]

		linearOrder = append(linearOrder, u)

		for v := range g.graph[u] {
			inDegree[v]--

			if inDegree[v] == 0 {
				next = append(next, v)
			}
		}
	}

	return linearOrder
}

func (g *DAG) WalkUp(ctx context.Context, nWorkers int, fn func(context.Context, Target) error) error {
	return g.reverse().WalkDown(ctx, nWorkers, fn)
}

func (g *DAG) WalkDown(ctx context.Context, nWorkers int, fn func(context.Context, Target) error) error {
	inDegree := map[Target]int{}
	for n := range g.graph {
		inDegree[n] = 0
	}
	for _, adjacent := range g.graph {
		for v := range adjacent {
			inDegree[v]++
		}
	}

	gr, ctx := errgroup.WithContext(ctx)
	next := make(chan Target, len(g.graph))
	complete := make(chan Target)

	gr.Go(func() error {
		for u, v := range inDegree {
			if v != 0 {
				continue
			}
			select {
			case next <- u:
				g.Logger.Debug().Str("target", u.Name()).Msg("queuing initial target")
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		g.Logger.Debug().Msg("initial targets queued")

		for {
			// Check if done
			done := true
			for k := range inDegree {
				if inDegree[k] != 0 {
					done = false
				}
			}
			if done {
				g.Logger.Debug().Msg("closing next")
				close(next)
				break
			}

			select {
			case u := <-complete:
				g.Logger.Debug().Str("target", u.Name()).Msg("target complete")
				for v := range g.graph[u] {
					inDegree[v]--
					g.Logger.Debug().Str("prevTarget", u.Name()).Str("target", v.Name()).Int("degree", inDegree[v]).Msg("reduce degree")
					if inDegree[v] == 0 {
						select {
						case <-ctx.Done():
							return ctx.Err()
						case next <- v:
							g.Logger.Debug().Str("prevTarget", u.Name()).Str("target", v.Name()).Msg("queuing next")
						}
					}
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		for _ = range complete {
			// Consume remaining completions (of root nodes) to exit gracefully
		}
		return nil
	})

	wg, wctx := errgroup.WithContext(ctx)
	for i := 0; i < nWorkers; i++ {
		wl := g.Logger.With().Int("worker", i).Logger()
		wg.Go(func() error {
			for {
				select {
				case <-wctx.Done():
					return wctx.Err()
				case l, ok := <-next:
					if !ok {
						wl.Debug().Msg("worker exiting")
						return nil
					}
					wl.Debug().Str("target", l.Name()).Msg("processing target")
					if err := fn(wctx, l); err != nil {
						wl.Err(err).Msg("dag walk function error")
						return err
					}
					wl.Debug().Str("target", l.Name()).Msg("target completed")
					complete <- l
				}
			}
		})
	}

	gr.Go(func() error {
		g.Logger.Debug().Msg("all workers done")
		defer close(complete)
		return wg.Wait()
	})
	return gr.Wait()
}
