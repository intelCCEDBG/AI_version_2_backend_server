package config

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
)

var Viper *viper.Viper

// Config is a struct that holds the configuration for the application
func LoadConfig() *viper.Viper {
	vp := viper.New()
	dirname, _ := os.Getwd()
	dirname = path.Join(dirname, "../../")
	dirname = path.Join(dirname, "config")
	vp.AddConfigPath(dirname)
	vp.SetConfigName("config")
	vp.SetConfigType("env")
	vp.AutomaticEnv()
	if err := vp.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", vp.ConfigFileUsed())
		Viper = vp
		return vp
	} else {
		fmt.Println("Error loading config file:", err)
		return nil
	}
}

func Find_suffix(target string) []string {
	var ans []string
	for _, key := range Viper.AllKeys() {
		if strings.HasSuffix(strings.ToUpper(key), target) {
			value := Viper.GetString(key)
			ans = append(ans, value)
			fmt.Printf("%s=%s\n", key, value)
		}
	}
	return ans
}
