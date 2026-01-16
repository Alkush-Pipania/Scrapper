package app

import "context"

type Container struct {
}

func NewContainer(ctx context.Context) *Container {
	return &Container{}
}
