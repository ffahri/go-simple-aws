package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"simpleawsgo/pkg/api"
)

func main() {
	readConfig()
	service := api.Service{}
	if err := service.Init(); err != nil {
		log.Fatal().Timestamp().Err(err).Msg("init err")
	} else {
		log.Info().Timestamp().Msg("clients are initialized")
	}
	r := gin.New()
	r.POST("/send", service.SendHandler) //TODO add tracing
	if err := r.Run(); err != nil {
		log.Fatal().Err(err).Timestamp().Msg("server stopped with error")
	}
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
