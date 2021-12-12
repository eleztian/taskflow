package taskflow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
)

func TestFlow(t *testing.T) {
	res, err := NewFlow("root", WithLogger(log.New(os.Stdout, "", 0))).Parallel(10,
		NewTask("func1", FuncRunner(func(ctx context.Context) (res interface{}, err error) {
			fmt.Println("func1")
			return nil, nil
		})),
		NewTask("func2", FuncRunner(func(ctx context.Context) (res interface{}, err error) {
			fmt.Println("func2")
			return nil, errors.New("failed")
		})),
	).Success(
		NewFlow("success").Sequential(
			NewTask("func3", FuncRunner(func(ctx context.Context) (res interface{}, err error) {
				fmt.Println("func3")
				return nil, nil
			})),
		),
	).Failed(
		NewFlow("failed").Sequential(
			NewTask("failed", FuncRunner(func(ctx context.Context) (res interface{}, err error) {
				fmt.Println("failed")
				return nil, nil
			})),
		).Success(
			NewFlow("ff").Sequential(
				NewTask("failed1", FuncRunner(func(ctx context.Context) (res interface{}, err error) {
					fmt.Println("failed1")
					return nil, nil
				})),
			),
		),
	).Run(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(res)
}
