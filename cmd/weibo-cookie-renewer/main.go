package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis/v7"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
		log.Panic("failed to read configuration", err)
	}

	var config configuration
	err = v.Unmarshal(&config)
	if err != nil {
		log.Panic("failed to unmarshal configuration", err)
	}

	if configFileNotFound {
		log.Println("configuration file not found")
	}

	fmt.Printf("%+v\n", config)

	err = sentry.Init(sentry.ClientOptions{
		Dsn: config.SentryDsn,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: config.Redis.Address,
	})
	key := config.Redis.Prefix + "_WEIBO_COOKIE"

	var rawCookies string
	rawCookies, err = client.Get(key).Result()
	if err != nil {
		log.Fatalf("get cookie from redis: %s=%s", key, err)
	}

	cookies := parsRawCookies(rawCookies)
	if len(cookies) == 0 {
		log.Fatalf("invalid rawCookies: %s", rawCookies)
	}

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, cookie := range cookies {
				success, err := network.SetCookie(cookie.Name, cookie.Value).
					WithDomain(".weibo.com").
					WithPath("/").
					WithHTTPOnly(true).
					Do(ctx)

				if err != nil {
					return err
				}
				if !success {
					return fmt.Errorf("could not set cookie %q to %q", cookie.Name, cookie.Value)
				}
			}

			return nil
		}),
		chromedp.Navigate("https://weibo.com/"),
		chromedp.WaitVisible("a[node-type=\"account\"]"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetAllCookies().Do(ctx)
			if err != nil {
				return err
			}

			var cookieStrings []string
			for _, cookie := range cookies {
				cookieStrings = append(cookieStrings, cookie.Name+"="+cookie.Value)
			}
			cookieString := strings.Join(cookieStrings, "; ")
			log.Println("New cookies: " + cookieString)
			if err := client.Set(key, cookieString, 0).Err(); err != nil {
				return err
			}
			return nil
		}),
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("done.")
}

func parsRawCookies(rawCookies string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	request := http.Request{Header: header}
	return request.Cookies()
}
