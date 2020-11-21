package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	middleware "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin"
	"simpleawsgo/pkg/api"
	"simpleawsgo/pkg/client"
)

func main() {
	/*cleanup := client.InitHoneyCombTracer("api-webischia")
	defer cleanup()
	*/
	client.InitLocalTracer("api-webischia")
	readConfig()
	service := api.Service{}
	if err := service.Init(); err != nil {
		log.Fatal().Timestamp().Err(err).Msg("init err")
	} else {
		log.Info().Timestamp().Msg("clients are initialized")
	}

	r := gin.New()
	r.Use(middleware.Middleware("aws-sqs-api"))
	r.POST("/send", service.SendHandler)
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
