package port

import (
	"context"
)

// DiagramHandler handler to generate a diagram given the input.
type DiagramHandler func(ctx context.Context, input Input) (Output, error)
