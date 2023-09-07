package scraper

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type Scraper struct {
	bearerToken    string
	client         *http.Client
	delay          int64
	guestToken     string
	guestCreatedAt time.Time
	includeReplies bool
	isLogged       bool
	isOpenAccount  bool
	oAuthToken     string
	oAuthSecret    string
	proxy          string
	searchMode     SearchMode
	wg             sync.WaitGroup
}

type SearchMode int

const (
	SearchTop SearchMode = iota
	SearchLatest
	SearchPhotos
	SearchVideos
	SearchUsers
)

const DefaultClientTimeout = 10 * time.Second

func New() *Scraper {
	jar, _ := cookiejar.New(nil)
	return &Scraper{
		bearerToken: randomGuestBearerToken,
		client: &http.Client{
			Jar:     jar,
			Timeout: DefaultClientTimeout,
		},
	}
}

func (s *Scraper) setBearerToken(token string) {
	s.bearerToken = token
	s.guestToken = ""
}

func (s *Scraper) IsGuestToken() bool {
	return s.guestToken != ""
}

func (s *Scraper) SetSearchMode(mode SearchMode) *Scraper {
	s.searchMode = mode
	return s
}

func (s *Scraper) WithDelay(seconds int64) *Scraper {
	s.delay = seconds
	return s
}

func (s *Scraper) WithReplies(b bool) *Scraper {
	s.includeReplies = b
	return s
}

func (s *Scraper) WithClientTimeout(timeout time.Duration) *Scraper {
	s.client.Timeout = timeout
	return s
}

func (s *Scraper) SetProxy(proxyAddr string) error {
	if proxyAddr == "" {
		s.client.Transport = &http.Transport{
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			DialContext: (&net.Dialer{
				Timeout: s.client.Timeout,
			}).DialContext,
		}
		s.proxy = ""
		return nil
	}
	if strings.HasPrefix(proxyAddr, "http") {
		urlproxy, err := url.Parse(proxyAddr)
		if err != nil {
			return err
		}
		s.client.Transport = &http.Transport{
			Proxy:        http.ProxyURL(urlproxy),
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			DialContext: (&net.Dialer{
				Timeout: s.client.Timeout,
			}).DialContext,
		}
		s.proxy = proxyAddr
		return nil
	}
	if strings.HasPrefix(proxyAddr, "socks5") {
		baseDialer := &net.Dialer{
			Timeout:   s.client.Timeout,
			KeepAlive: s.client.Timeout,
		}
		socksHostPort := strings.ReplaceAll(proxyAddr, "socks5://", "")
		dialSocksProxy, err := proxy.SOCKS5("tcp", socksHostPort, nil, baseDialer)
		if err != nil {
			return errors.New("error creating socks5 proxy :" + err.Error())
		}
		if contextDialer, ok := dialSocksProxy.(proxy.ContextDialer); ok {
			dialContext := contextDialer.DialContext
			s.client.Transport = &http.Transport{
				DialContext: dialContext,
			}
		} else {
			return errors.New("failed type assertion to DialContext")
		}
		s.proxy = proxyAddr
		return nil
	}
	return errors.New("only support http(s) or socks5 protocol")
}
