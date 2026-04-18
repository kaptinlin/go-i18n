package i18n

import "context"

type localizerContextKey struct{}

// ContextWithLocalizer returns a copy of ctx that carries l.
func ContextWithLocalizer(ctx context.Context, l *Localizer) context.Context {
	return context.WithValue(ctx, localizerContextKey{}, l)
}

// LocalizerFromContext returns the localizer stored in ctx.
func LocalizerFromContext(ctx context.Context) (*Localizer, bool) {
	if ctx == nil {
		return nil, false
	}
	l, ok := ctx.Value(localizerContextKey{}).(*Localizer)
	return l, ok && l != nil
}
