package observability

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type ShutdownFunc func(context.Context) error

func Setup(ctx context.Context, serviceName, endpoint string, insecure bool) (ShutdownFunc, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	var traceProvider *trace.TracerProvider
	var metricProvider *metric.MeterProvider

	if endpoint == "" {
		traceProvider = trace.NewTracerProvider(trace.WithResource(res))
		metricProvider = metric.NewMeterProvider(metric.WithResource(res))
	} else {
		exportCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		traceOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(endpoint)}
		metricOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(endpoint)}
		if insecure {
			traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
			metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
		}

		traceExporter, err := otlptracegrpc.New(exportCtx, traceOpts...)
		if err != nil {
			return nil, err
		}
		metricExporter, err := otlpmetricgrpc.New(exportCtx, metricOpts...)
		if err != nil {
			return nil, err
		}

		traceProvider = trace.NewTracerProvider(
			trace.WithResource(res),
			trace.WithBatcher(traceExporter),
		)
		metricProvider = metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		)
	}

	otel.SetTracerProvider(traceProvider)
	otel.SetMeterProvider(metricProvider)

	shutdown := func(shutdownCtx context.Context) error {
		var shutdownErr error
		if err := traceProvider.Shutdown(shutdownCtx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
		if err := metricProvider.Shutdown(shutdownCtx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
		return shutdownErr
	}

	return shutdown, nil
}
