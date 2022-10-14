package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
)

type AmqpHeadesCarrier map[string]interface{}

func (a AmqpHeadesCarrier) Get(key string) string {
	v, ok := a[key]
	if !ok {
		return ""
	}
	return v.(string)
}

func (a AmqpHeadesCarrier) Set(key string, value string) {
	a[key] = value
}

func (a AmqpHeadesCarrier) Keys() []string {
	i := 0
	r := make([]string, len(a))

	for k := range a {
		r[i] = k
		i++
	}

	return r
}

func InjectAMQPHeaders(ctx context.Context) map[string]interface{} {
	h := make(AmqpHeadesCarrier)
	otel.GetTextMapPropagator().Inject(ctx, h)
	return h
}

func ExtractAMPQHeaders(ctx context.Context, headers map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, AmqpHeadesCarrier(headers))
}
