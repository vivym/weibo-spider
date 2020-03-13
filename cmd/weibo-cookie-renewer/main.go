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

	"github.com/go-redis/redis/v7"

	"github.com/getsentry/sentry-go"

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

	// rawCookies := `_s_tentry=-; Apache=2738506627573.054.1583602904976; SINAGLOBAL=2738506627573.054.1583602904976; TC-V5-G0=4de7df00d4dc12eb0897c97413797808; ULV=1583602905114:1:1:1:2738506627573.054.1583602904976:; UOR=,,www.cnblogs.com; Ugrow-G0=e1a5a1aae05361d646241e28c550f987; login_sid_t=2502b18acef56db2b0f075265b31e724; cross_origin_proto=SSL; wb_view_log=1440*9002; SUBP=0033WrSXqPxfM725Ws9jqgMF55529P9D9WF6Wjx-ufIVS5xnBo2216SM5JpX5K2hUgL.FoM01KM0S02pSKM2dJLoIEBLxK-LB--L1h.LxK-LBo5L12qLxKBLBonLBoqLxK.L1h2L1KMt; ALF=1615658708; SSOLoginState=1584122709; SCF=AjxQB8pXJEmLM4RU1JQNK3aD7jaTlGuPVVnBubqpuBjaxIuPQio0GloKvkPKRFHAnAd2_srL0sI3SaJg7AvXy0Y.; SUB=_2A25zb7sFDeRhGeFN4lUS9y_NzjuIHXVQHKvNrDV8PUNbmtAKLVXykW9NQ7ow0zcE5GQF-EBheHQYjkpGWJFQm8ZZ; SUHB=0auABYNZAq4GHO; un=13036760718; wvr=6; wb_view_log_7397371157=1440*9002; WBtopGlobal_register_version=3d5b6de7399dfbdb; TC-Page-G0=45685168db6903150ce64a1b7437dbbb|1584123446|1584123446; webim_unReadCount=%7B%22time%22%3A1584124352583%2C%22dm_pub_total%22%3A1%2C%22chat_group_client%22%3A0%2C%22allcountNum%22%3A37%2C%22msgbox%22%3A0%7D`
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
