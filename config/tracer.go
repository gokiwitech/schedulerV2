package config

import (
	"context"
	"schedulerV2/models"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer() func(context.Context) error {

	var secureOption otlptracegrpc.Option

	if Env == "local" || Env == "staging" {
		return nil
	}

	if models.AppConfig.InsecureMode == false {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(models.AppConfig.CollectorURL),
		),
	)

	if err != nil {
		lg.Fatal().Err(err).Msg("Failed to Initalise the tracer")
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", models.AppConfig.ServiceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		lg.Error().Msgf("ZooKeeper session expired. Re-establishing connection... %v", err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)
	return exporter.Shutdown
}
