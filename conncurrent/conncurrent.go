package conncurrent

import (
	"context"
	"sync"
)

type Group struct {
	ctx    context.Context
	cancel func()

	wg sync.WaitGroup

	limitChan chan struct{}
	errOnce   sync.Once
	err       error
}

func WithContext(ctx context.Context, limit int) (*Group, context.Context) {
	ctxTask, cancel := context.WithCancel(ctx)
	res := &Group{cancel: cancel, ctx: ctxTask}
	if limit > 0 {
		res.limitChan = make(chan struct{}, limit)
		for i := limit; i > 0; i-- {
			res.limitChan <- struct{}{}
		}
	}
	return res, ctx
}

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	if g.limitChan != nil {
		close(g.limitChan)
	}
	return g.err
}

func (g *Group) Go(f func() error) {
	if g.limitChan != nil {
		select {
		case _, ok := <-g.limitChan:
			if !ok {
				return
			}
		case <-g.ctx.Done():
			return
		}
	}
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
		g.limitChan <- struct{}{}
	}()
}
