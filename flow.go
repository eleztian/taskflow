package taskflow

import (
	"context"
	"github.com/pkg/errors"

	"github.com/eleztian/taskflow/conncurrent"
)

type Flow interface {
	Task
	Sequential(ts ...Task) Flow
	Parallel(limit int64, ts ...Task) Flow
	Success(flow Task) Flow
	Failed(flow Task) Flow
}

type flow struct {
	name    string
	process Runner
	recover bool
	logger  Logger
}

type OptionFunc func(flow *flow)

func NewFlow(name string, ops ...OptionFunc) Flow {
	res := &flow{
		name:   name,
		logger: NoopLogger{},
	}
	for _, op := range ops {
		op(res)
	}

	return res
}

func (f *flow) Name() string {
	return f.name
}

func (f *flow) Sequential(ts ...Task) Flow {
	process := f.process

	f.process = FuncRunner(func(ctx context.Context) (res interface{}, err error) {
		if process != nil {
			res, err = process.Run(ctx)
			if err != nil {
				return nil, err
			}
			ctx = setRes(ctx, res)
		}
		if f.recover {
			defer func() {
				if e := recover(); e != nil {
					err = errors.Errorf("Panic:%v", e)
				}
			}()
		}
		for _, t := range ts {
			f.logger.Printf("Start to exec %s", t.Name())
			res, err := t.Run(ctx)
			if err != nil {
				f.logger.Printf("Failed to exec %s %v", t.Name(), err)
				return nil, err
			}
			ctx = setRes(ctx, res)
			f.logger.Printf("Success to exec %s %#v", t.Name(), res)
		}
		return getRes(ctx), nil
	})
	return f
}

func (f *flow) Parallel(limit int64, ts ...Task) Flow {
	process := f.process

	f.process = FuncRunner(func(ctx context.Context) (interface{}, error) {
		if process != nil {
			res, err := process.Run(ctx)
			if err != nil {
				return nil, err
			}
			ctx = setRes(ctx, res)
		}
		wg, ctxTask := conncurrent.WithContext(ctx, int(limit))
		for _, t := range ts {
			task := t
			wg.Go(func() (err error) {
				if f.recover {
					defer func() {
						if e := recover(); e != nil {
							err = errors.Wrap(errors.Errorf("Panic:%v", e), task.Name())
						}
					}()
				}
				f.logger.Printf("Start to exec %s", task.Name())
				_, err = task.Run(ctxTask)
				if err != nil {
					f.logger.Printf("Failed to exec %s %v", task.Name(), err)
					return errors.Wrap(err, task.Name())
				}
				f.logger.Printf("Success to exec %s", task.Name())
				return nil
			})
		}
		err := wg.Wait()
		if err != nil {
			return nil, err
		}

		return getRes(ctx), nil
	})
	return f
}

func (f *flow) Success(task Task) Flow {
	if f.process != nil {
		process := f.process
		f.process = FuncRunner(func(ctx context.Context) (res interface{}, err error) {
			res, err = process.Run(ctx)
			if err != nil {
				return nil, err
			}

			if f.recover {
				defer func() {
					if e := recover(); e != nil {
						err = errors.Wrap(errors.Errorf("Panic:%v", e), task.Name())
					}
				}()
			}

			f.logger.Printf("Start to exec %s", task.Name())
			res, err = task.Run(setRes(ctx, res))
			if err != nil {
				f.logger.Printf("Failed to exec %s %v", task.Name(), err)
				return nil, err
			}
			f.logger.Printf("Success to exec %s %#v", task.Name(), res)

			return res, nil
		})
	} else {
		f.process = task
	}

	return f
}

func (f *flow) Failed(task Task) Flow {
	if f.process != nil {
		process := f.process
		f.process = FuncRunner(func(ctx context.Context) (interface{}, error) {
			res, err := process.Run(ctx)
			if err != nil {
				if f.recover {
					defer func() {
						if e := recover(); e != nil {
							err = errors.Wrap(errors.Errorf("Panic:%v", e), task.Name())
						}
					}()
				}
				res, err = task.Run(ctx)
				if err != nil {
					f.logger.Printf("Failed to exec %s %v", task.Name(), err)
					return nil, err
				}
				f.logger.Printf("Success to exec %s %#v", task.Name(), res)
				return res, nil
			}
			return res, nil
		})
	}

	return f
}

func (f *flow) Run(ctx context.Context) (res interface{}, err error) {
	if f.process != nil {
		return f.process.Run(ctx)
	}

	return
}

var resKey = struct{}{}

func setRes(ctx context.Context, value interface{}) context.Context {
	return context.WithValue(ctx, resKey, value)
}

func getRes(ctx context.Context) interface{} {
	return ctx.Value(resKey)
}

func WithLogger(l Logger) OptionFunc {
	return func(flow *flow) {
		flow.logger = l
	}
}

func WithRecover() OptionFunc {
	return func(flow *flow) {
		flow.recover = true
	}
}
