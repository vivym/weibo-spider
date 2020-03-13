package main

import (
	"fmt"
	"os"

	"github.com/vivym/weibo-spider/internal/nlp"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"logur.dev/logur"

	"github.com/vivym/weibo-spider/internal/db"
	"github.com/vivym/weibo-spider/internal/log"
	"github.com/vivym/weibo-spider/internal/spider"
)

// Provisioned by ldflags
// nolint: gochecknoglobals
var (
	version    string
	commitHash string
	buildDate  string
)

func main() {
	v, p := viper.New(), pflag.NewFlagSet(friendlyAppName, pflag.ExitOnError)

	configure(v, p)

	p.String("config", "", "Configuration file")
	p.Bool("version", false, "Show version information")

	_ = p.Parse(os.Args[1:])

	if v, _ := p.GetBool("version"); v {
		fmt.Printf("%s version %s (%s) built on %s\n", friendlyAppName, version, commitHash, buildDate)

		os.Exit(0)
	}

	if c, _ := p.GetString("config"); c != "" {
		v.SetConfigFile(c)
	}

	err := v.ReadInConfig()
	_, configFileNotFound := err.(viper.ConfigFileNotFoundError)
	if !configFileNotFound {
		emperror.Panic(errors.Wrap(err, "failed to read configuration"))
	}

	var config configuration
	err = v.Unmarshal(&config)
	emperror.Panic(errors.Wrap(err, "failed to unmarshal configuration"))

	err = config.PostProcess()
	emperror.Panic(errors.WithMessage(err, "failed to post-process configuration"))

	// Create logger (first thing after configuration loading)
	logger := log.NewLogger(config.Log)

	// Provide some basic context to all log lines
	logger = logur.WithFields(logger, map[string]interface{}{"application": appName})

	log.SetStandardLogger(logger)

	if configFileNotFound {
		logger.Warn("configuration file not found")
	}

	err = config.Validate()
	if err != nil {
		logger.Error(err.Error())

		os.Exit(3)
	}

	fmt.Printf("%+v\n", config)

	err = db.SetupDB(config.DB)
	if err != nil {
		logger.Error(err.Error())

		os.Exit(3)
	}

	var nlpToolkit *nlp.NLPToolkit
	nlpToolkit, err = nlp.New(config.NLP)
	if err != nil {
		logger.Error("nlp error: " + err.Error())
		os.Exit(-1)
	}

	app := spider.New(config.Spider, logger, nlpToolkit)
	errCode := app.Run()
	if errCode != 0 {
		os.Exit(errCode)
	}
}
