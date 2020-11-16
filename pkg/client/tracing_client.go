package client

import (
	"github.com/honeycombio/opentelemetry-exporter-go/honeycomb"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/api/global"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"os"
)

func InitTracer(serviceName string) func() {
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
