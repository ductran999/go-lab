// Package tracing bootstraps OpenTelemetry once per binary: OTLP/gRPC
// exporter to the collector, W3C trace-context propagation, always-on
// sampling for the lab (prod samples tail-based instead).
package tracing

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// ServiceInfo identifies one binary in the trace backend: name groups
// the domain, version pins the build, namespace the biz scope.
// Instance and Environment come from the runtime (hostname, DEPLOY_ENV).
// Pass it to Setup, never strings loose.
type ServiceInfo struct {
	Name        string
	Version     string
	Namespace   string
	InstanceID  string
	Environment string
}

// NewServiceInfo builds ServiceInfo with runtime defaults: hostname for
// instance (K8s sets it to the pod name), DEPLOY_ENV or "development".
func NewServiceInfo(name, version, namespace string) ServiceInfo {
	host, _ := os.Hostname()

	env := os.Getenv("DEPLOY_ENV")
	if env == "" {
		env = "development"
	}

	return ServiceInfo{
		Name:        name,
		Version:     version,
		Namespace:   namespace,
		InstanceID:  host,
		Environment: env,
	}
}

// Setup wires the global tracer provider. Call once in main; the
// returned func shuts the exporter down (defer it).
func Setup(ctx context.Context, info ServiceInfo, endpoint string) (func(context.Context) error, error) {
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("tracing: exporter: %w", err)
	}

	attrs := resource.WithAttributes(
		semconv.ServiceName(info.Name),
		semconv.ServiceVersion(info.Version),
		semconv.ServiceNamespace(info.Namespace),
		semconv.ServiceInstanceID(info.InstanceID),
		semconv.DeploymentEnvironment(info.Environment),
	)

	res, err := resource.New(ctx, attrs)
	if err != nil {
		return nil, fmt.Errorf("tracing: resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider.Shutdown, nil
}

// Business stamps tenant/user onto the current span AND into baggage:
// attributes live one span, baggage rides the whole journey (svc-b
// reads it without re-auth). Call it in delivery (post-auth).
func Business(ctx context.Context, tenantID, userID string) context.Context {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("tenant.id", tenantID),
		attribute.String("user.id", userID),
	)

	member1, _ := baggage.NewMember("tenant.id", tenantID)
	member2, _ := baggage.NewMember("user.id", userID)

	bag, _ := baggage.FromContext(ctx).SetMember(member1)
	bag, _ = bag.SetMember(member2)

	return baggage.ContextWithBaggage(ctx, bag)
}

// TraceIDOf reads the incoming parent trace id from the raw
// traceparent header — never from ctx (inside otelhttp the ctx already
// carries the live span, which would echo our own id back).
// "-" means edge: no parent sent one, this service mints the trace.
func TraceIDOf(r *http.Request, _ context.Context) string {
	parts := strings.Split(r.Header.Get("Traceparent"), "-")
	if len(parts) != 4 || parts[1] == "" {
		return "-"
	}

	return parts[1]
}
