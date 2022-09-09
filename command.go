package shodo

import (
	"context"
	"io"
)

var (
	dispatch = map[string]*command{
		"lint": {run: doLint},
	}
)

type command struct {
	run func(context.Context, []string, io.Writer, io.Writer) error
}

func (c *command) Run(ctx context.Context, argv []string, out, errOut io.Writer) error {
	return c.run(ctx, argv, out, errOut)
}
