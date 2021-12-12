package taskflow

import "context"

type Runner interface {
	Run(ctx context.Context) (interface{}, error)
}

type FuncRunner func(ctx context.Context) (interface{}, error)

func (f FuncRunner) Run(ctx context.Context) (interface{}, error) {
	return f(ctx)
}
