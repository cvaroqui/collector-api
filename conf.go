package main

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

func initConf() error {
	// env
	viper.SetEnvPrefix("API")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// defaults
	viper.SetDefault("listen", "127.0.0.1:8080")
	viper.SetDefault("db.username", "opensvc")
	viper.SetDefault("db.host", "127.0.0.1")
	viper.SetDefault("db.port", "3306")
	viper.SetDefault("db.log.level", "warn")
	viper.SetDefault("db.log.slow_query_threshold", "1s")

	// config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	log.Printf("add config path: %s", "/etc/api")
	viper.AddConfigPath("/etc/api")
	log.Printf("add config path: %s", "$HOME/.api")
	viper.AddConfigPath("$HOME/.api")
	log.Printf("add config path: %s", ".")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		} else {
			log.Println(err)
		}
	}
	return nil
}
