package contextx

import "context"

type SkipAuth bool

type contextKeySkipAuth struct{}

func WithSkipAuth(ctx context.Context, v SkipAuth) context.Context {
	return context.WithValue(ctx, contextKeySkipAuth{}, v)
}

func GetSkipAuth(ctx context.Context) SkipAuth {
	v, _ := ctx.Value(contextKeySkipAuth{}).(SkipAuth)
	return v
}
