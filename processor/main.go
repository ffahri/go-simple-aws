package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"simpleawsgo/pkg/client"
	"simpleawsgo/pkg/processor"
)

func main() {
	cleanup := client.InitTracer("sqspoller-webischia")
	defer cleanup()
	readConfig()
	qService := processor.Service{}
	if err := qService.Init(); err != nil {
		log.Fatal().Timestamp().Err(err).Msg("init err")
	} else {
		log.Info().Timestamp().Msg("clients are initialized")
	}
	qService.StartPoller()
}

func readConfig() { //TODO call from pkg to prevent duplication
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
