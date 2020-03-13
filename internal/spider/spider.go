package spider

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Kamva/mgm/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-redis/redis/v7"
	"github.com/go-resty/resty/v2"
	"logur.dev/logur"

	"github.com/vivym/weibo-spider/internal/model"
	"github.com/vivym/weibo-spider/internal/nlp"
)

type Spider struct {
	http   *resty.Client
	config *Config
	logger logur.LoggerFacade
	nlp    *nlp.NLPToolkit
}

func New(config Config, logger logur.LoggerFacade, nlpToolkit *nlp.NLPToolkit) *Spider {
	http := resty.New().
		SetRetryCount(3).
		SetRetryWaitTime(5*time.Second).
		SetRetryMaxWaitTime(20*time.Second).
		SetHostURL(baseURL).
		SetHeader("User-Agent", userAgent)

	return &Spider{
		http:   http,
		config: &config,
		logger: logger,
		nlp:    nlpToolkit,
	}
}

func (s *Spider) Run() int {
	client := redis.NewClient(&redis.Options{
		Addr: s.config.Redis.Address,
	})
	cookie, err := client.Get(s.config.Redis.Prefix + "_WEIBO_COOKIE").Result()
	if err != nil {
		s.logger.Error("fetch cookies error: " + err.Error())
		return -1
	}
	s.http.SetHeader("Cookie", cookie)

	hotTopicList, err := s.fetchHotTopics()
	if err != nil {
		s.logger.Error(err.Error())
		return -1
	}
	if len(hotTopicList.Data) == 0 {
		return 0
	}

	var weiboHotTopics model.WeiboHotTopics
	var sentences []string
	for _, hotTopic := range hotTopicList.Data {
		if hotTopic.Heat == 0 {
			continue
		}
		if len(sentences) > s.config.MaxTopics+1 {
			break
		}
		delay := time.Duration(s.config.Delay + rand.Intn(200))
		time.Sleep(delay * time.Millisecond)

		s.logger.Info(hotTopic.Title)
		sentence := strings.Join(s.fetchAllTweets(hotTopic.URL), "\n")
		sentences = append(sentences, sentence)

		weiboTopic := model.WeiboTopic{
			Title: hotTopic.Title,
			URL:   hotTopic.URL,
			Heat:  int32(hotTopic.Heat),
		}
		keywords, err := s.nlp.ExtractKeywords(sentence, 100)
		if err != nil {
			s.logger.Error("ExtractKeywords error: " + err.Error())
		}
		weiboTopic.Keywords = keywords

		weiboHotTopics.Topics = append(weiboHotTopics.Topics, weiboTopic)
	}

	sentence := strings.Join(sentences, "\n")
	keywords, err := s.nlp.ExtractKeywords(sentence, 100)
	if err != nil {
		s.logger.Error("ExtractKeywords error: " + err.Error())
	}
	weiboHotTopics.Keywords = keywords

	if err := mgm.Coll(&weiboHotTopics).Create(&weiboHotTopics); err != nil {
		s.logger.Error("insert hot topics error: " + err.Error())
		return -1
	}

	if err := mgm.Coll(hotTopicList).Create(hotTopicList); err != nil {
		s.logger.Error("insert hot list error: " + err.Error())
		return -1
	}

	return 0
}

func (s *Spider) fetchAllTweets(url string) []string {
	sentences, numPages, err := s.fetchTweetsOnePage(url, 1)
	if err != nil {
		s.logger.Error("fetchTweetsOnePage error: " + err.Error())
	}
	for i := 2; i <= s.config.MaxPages && i <= numPages; i++ {
		texts, _, err := s.fetchTweetsOnePage(url, i)
		if err != nil {
			s.logger.Error("fetchTweetsOnePage error: " + err.Error())
		}
		sentences = append(sentences, texts...)

		delay := time.Duration(s.config.Delay + rand.Intn(200))
		time.Sleep(delay * time.Millisecond)
	}

	return sentences
}

func (s *Spider) fetchTweetsOnePage(url string, page int) ([]string, int, error) {
	s.logger.Info("page " + strconv.Itoa(page))

	rsp, err := s.http.R().
		SetDoNotParseResponse(true).
		Get(url + fmt.Sprintf("&page=%d", page))
	if err != nil {
		return nil, 0, err
	}

	body := rsp.RawBody()
	defer body.Close()
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, 0, err
	}

	var sentences []string
	doc.Find(".card-wrap[action-type=\"feed_list_item\"]>.card").Each(func(_ int, el *goquery.Selection) {
		contentEl := el.Find(".card-feed>.content>.txt[node-type=\"feed_list_content_full\"]")
		if contentEl.Length() == 0 {
			contentEl = el.Find(".card-feed>.content>.txt[node-type=\"feed_list_content\"]")
		}
		sentences = append(sentences, contentEl.Text())
	})

	numPages := doc.Find(".m-page>div>.list>ul>li").Length()

	return sentences, numPages, nil
}

func (s *Spider) fetchHotTopics() (*model.HotTopicList, error) {
	rsp, err := s.http.R().
		SetDoNotParseResponse(true).
		Get("/top/summary?cate=realtimehot")
	if err != nil {
		return nil, err
	}
	body := rsp.RawBody()
	defer body.Close()
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	hotTopicList := model.HotTopicList{}
	hotTopicList.Data = make([]model.HotTopic, 0, 51)
	doc.Find("#pl_top_realtimehot>table>tbody>tr>.td-02>a").Each(func(_ int, el *goquery.Selection) {
		hotTopic := model.HotTopic{Title: el.Text()}
		hotTopic.URL, _ = el.Attr("href")
		if nextEl := el.Next(); nextEl != nil && nextEl.Is("span") {
			hotTopic.Heat, _ = strconv.Atoi(nextEl.Text())
		}
		if strings.HasPrefix(hotTopic.URL, "/weibo") {
			hotTopicList.Data = append(hotTopicList.Data, hotTopic)
		}
	})

	return &hotTopicList, nil
}
