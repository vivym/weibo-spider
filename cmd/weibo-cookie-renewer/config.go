package main

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type configuration struct {
	SentryDsn string

	Redis struct {
		Address string
		Prefix  string
	}
}

// configure configures some defaults in the Viper instance.
func configure(v *viper.Viper, p *pflag.FlagSet) {
	// Viper settings
	v.AddConfigPath(".")
	v.AddConfigPath(fmt.Sprintf("$%s_CONFIG_DIR/", strings.ToUpper(envPrefix)))

	// Environment variable settings
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	// Application constants
	v.Set("appName", appName)

	v.SetDefault("redis.address", "localhost:6379")
	v.SetDefault("redis.prefix", "VKMDFFYUOTPPNVBR")
}
