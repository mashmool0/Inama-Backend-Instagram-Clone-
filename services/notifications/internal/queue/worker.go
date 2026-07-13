package queue

import "context"

type Worker interface {
	Run(ctx context.Context) error
}

type NoopWorker struct{}

func (NoopWorker) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
