package http_utils

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net/http"
)

type (
	OpenTelemetry struct {
		traceProvider trace.TracerProvider
		bsp           sdktrace.SpanProcessor
		tracer        trace.Tracer
	}

	Span interface {
		End()
	}

	Spanner interface {
		Start(ctx context.Context, name string) (context.Context, Span)
		Shutdown(ctx context.Context)
	}

	dummySpanner struct {
	}
	otelSpan struct {
		span trace.Span
	}
)

func NewRequest(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))
	return r, nil
}

func DoRequest(spanner Spanner, name string, r *http.Request) (*http.Response, error) {
	ctx, span := spanner.Start(r.Context(), name)
	defer span.End()
	return http.DefaultClient.Do(r.WithContext(ctx))
}

func NewOpenTelemetry(ctx context.Context, serviceName, collectorUrl string) (*OpenTelemetry, error) {
	log.Printf("Starting otel tracing service for %s: %s", serviceName, collectorUrl)

	instance := &OpenTelemetry{}

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(serviceName),
	))
	if err != nil {
		return nil, err
	}
	grpcConn, err := grpc.NewClient(collectorUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	transporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(grpcConn))
	if err != nil {
		return nil, err
	}
	instance.bsp = sdktrace.NewBatchSpanProcessor(transporter)
	instance.traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(instance.bsp),
	)
	otel.SetTracerProvider(instance.traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	instance.tracer = otel.Tracer(serviceName + "-tracer")

	return instance, nil
}

func NewEmptySpanner() Spanner {
	return &dummySpanner{}
}

func (o *OpenTelemetry) Shutdown(ctx context.Context) {
	log.Println("Shuting down OpenTelemetry...")
	_ = o.bsp.Shutdown(ctx)
	log.Println("OpenTelemetry shutdown completed")
}

func (o *OpenTelemetry) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		carier := propagation.HeaderCarrier(r.Header)
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), carier)
		ctx, span := o.tracer.Start(ctx, fmt.Sprintf("[%s] %s", r.Method, r.URL.Path))
		defer span.End()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (o *OpenTelemetry) Start(ctx context.Context, name string) (context.Context, Span) {
	ctx, span := o.tracer.Start(ctx, name)
	return ctx, &otelSpan{span: span}
}

func (d *dummySpanner) Start(ctx context.Context, name string) (context.Context, Span) {
	return ctx, d
}

func (d *dummySpanner) End() {
}

func (d *dummySpanner) Shutdown(ctx context.Context) {
}

func (s *otelSpan) End() {
	s.span.End()
}
