package config

import (
	"github.com/spf13/viper"
)

// Config is the configuration for the recorder
func Main() error {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
