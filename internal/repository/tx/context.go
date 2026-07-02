package tx

import "context"

type contextKey struct{}

func injectTx(ctx context.Context, q Querier) context.Context {
	return context.WithValue(ctx, contextKey{}, q)
}

func FromContext(ctx context.Context) (Querier, bool) {
	q, ok := ctx.Value(contextKey{}).(Querier)
	return q, ok
}
