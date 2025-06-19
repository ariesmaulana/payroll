// lib/contextutil/context.go
package contextutil

import (
	"context"
)

type contextKey string

const (
	TraceKey    contextKey = "trace"
	TraceIdKey  contextKey = "traceId"
	authUserKey contextKey = "auth_user"
)

// GetTraceID retrieves the trace ID from the context
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceId, ok := ctx.Value(TraceIdKey).(string); ok {
		return traceId
	}
	return ""
}

// GetTrace retrieves the Trace object from the context
func GetTrace(ctx context.Context) (*Trace, bool) {
	if ctx == nil {
		return nil, false
	}
	trace, ok := ctx.Value(TraceKey).(*Trace)
	return trace, ok
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, TraceIdKey, traceId)
}

// WithTrace adds the Trace object to the context
func WithTrace(ctx context.Context, trace *Trace) context.Context {
	return context.WithValue(ctx, TraceKey, trace)
}
