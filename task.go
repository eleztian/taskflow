package taskflow

type Task interface {
	Name() string
	Runner
}

type TaskFunc struct {
	name string
	Runner
}

func NewTask(name string, runner Runner) Task {
	return &TaskFunc{Runner: runner, name: name}
}

func (t TaskFunc) Name() string {
	return t.name
}

var (
	_ Task = (*TaskFunc)(nil)
)
