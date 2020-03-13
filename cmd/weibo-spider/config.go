package main

import (
	"strings"

	"github.com/vivym/weibo-spider/internal/nlp"

	"github.com/vivym/weibo-spider/internal/db"
	"github.com/vivym/weibo-spider/internal/log"
	"github.com/vivym/weibo-spider/internal/spider"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36"
)

// configuration holds any kind of configuration that comes from the outside world and
// is necessary for running the application.
type configuration struct {
	// Log configuration
	Log log.Config

	// DB configuration
	DB db.Config

	NLP nlp.Config

	// Spider configuration
	Spider spider.Config
}

// Process post-processes configuration after loading it.
// nolint: unparam
func (c configuration) PostProcess() error {
	return nil
}

// Validate validates the configuration.
func (c *configuration) Validate() error {
	return nil
}

func configure(v *viper.Viper, p *pflag.FlagSet) {
	v.AddConfigPath("./configs")

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	// Application constants
	v.Set("appName", appName)

	v.SetDefault("log.format", "logfmt")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.nocolor", true)

	v.SetDefault("db.dbname", "ncovis")
	v.SetDefault("db.uri", "mongodb://localhost:27017/")

	v.SetDefault("nlp.address", "localhost:12377")

	v.SetDefault("spider.redis.address", "localhost:6379")
	v.SetDefault("spider.redis.prefix", "VKMDFFYUOTPPNVBR")
	v.SetDefault("spider.delay", 800)
	v.SetDefault("spider.maxTopics", 10)
	v.SetDefault("spider.maxPages", 20)
}
