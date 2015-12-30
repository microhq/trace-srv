package trace

import (
	"github.com/micro/trace-srv/db"

	proto "github.com/micro/go-platform/trace/proto"
	"golang.org/x/net/context"
)

var (
	TraceTopic = "micro.trace.span"

	DefaultTrace = newTrace()
)

type trace struct{}

func newTrace() *trace {
	return &trace{}
}

func (t *trace) ProcessSpan(ctx context.Context, span *proto.Span) error {
	// Only sample if we have a Debug flag
	if !span.Debug {
		return nil
	}
	return db.Create(span)
}

func ProcessSpan(ctx context.Context, span *proto.Span) error {
	return DefaultTrace.ProcessSpan(ctx, span)
}
