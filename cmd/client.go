package cmd

import (
	"log"

	mot "github.com/alzabo/mot/pkg"
	"github.com/spf13/viper"
)

func newClient() mot.Client {
	c, err := mot.NewClient(viper.GetString("url"), viper.GetString("username"), viper.GetString("password"))
	if err != nil {
		log.Fatal(err)
	}
	return c
}
