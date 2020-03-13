package spider

const (
	baseURL   = "https://s.weibo.com"
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36"
)

type Config struct {
	Redis struct {
		Address string
		Prefix  string
	}
	Delay     int
	MaxTopics int
	MaxPages  int
}
