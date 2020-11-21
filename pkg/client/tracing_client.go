package client

import (
	"github.com/honeycombio/opentelemetry-exporter-go/honeycomb"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"os"
)

func InitHoneyCombTracer(serviceName string) func() {
	// Get this via https://ui.honeycomb.io/account after signing up for Honeycomb
	apikey, _ := os.LookupEnv("HONEYCOMB_API_KEY")
	dataset, _ := os.LookupEnv("HONEYCOMB_DATASET")
	hny, err := honeycomb.NewExporter(
		honeycomb.Config{
			APIKey: apikey,
		},
		honeycomb.TargetingDataset(dataset),
		honeycomb.WithServiceName(serviceName), // replace with your app's name
	)
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error while creating new exporter")
	}

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(hny),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error while creating new provider")
	}
	global.SetTraceProvider(tp)
	return hny.Close
}

func InitLocalTracer(serviceName string) {

	exp, err := jaeger.NewRawExporter(jaeger.WithCollectorEndpoint("http://localhost:14269/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: serviceName,
		}))
	if err != nil {
		log.Fatal().Err(err).Msg("can't create new jaeger exporter")
	}
	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exp),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error while creating new provider")
	}
	global.SetTraceProvider(tp)

}
