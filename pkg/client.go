package mot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"

	"github.com/alzabo/mot/torrents"
)

const (
	login       = "/api/v2/auth/login"
	torrentInfo = "/api/v2/torrents/info"
)

type Client struct {
	HttpClient http.Client
	BaseUrl    string
}

func NewClient(url string, username string, password string) (Client, error) {
	var err error
	client := Client{BaseUrl: url}

	// All users of cookiejar should import "golang.org/x/net/publicsuffix"
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return client, fmt.Errorf("failed to initialize cookiejar: %s", err)
	}
	client.HttpClient = http.Client{
		Jar: jar,
	}
	err = client.Login(username, password)
	return client, err
}

func (c *Client) Login(username, password string) error {
	loginUrl, err := url.JoinPath(c.BaseUrl, login)
	if err != nil {
		return fmt.Errorf("failed to create login URL: %s", err)
	}
	params := url.Values{
		"username": {username},
		"password": {password},
	}
	resp, err := c.HttpClient.PostForm(loginUrl, params)
	if err != nil {
		return fmt.Errorf("login request failed: %s", err)
	}
	defer resp.Body.Close()
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("encountered error reading login response: %s", err)
	}
	if string(r) != "Ok." {
		return fmt.Errorf("failed to log in with message: %s", r)
	}
	return nil
}

type QueryOption func(*url.Values)

func WithHashes(hashes []string) QueryOption {
	return func(v *url.Values) {
		v.Set("hashes", strings.Join(hashes, "|"))
	}
}

// TODO: accept variadc options
func (c *Client) TorrentList(opts ...QueryOption) []torrents.Info {
	infoApi, err := url.JoinPath(c.BaseUrl, torrentInfo)
	if err != nil {
		log.Fatalf("failed to create url: %s", err)
	}
	infoUrl, _ := url.Parse(infoApi)
	params := url.Values{}
	for _, opt := range opts {
		opt(&params)
	}
	infoUrl.RawQuery = params.Encode()
	info, err := c.HttpClient.Get(infoUrl.String())
	if err != nil {
		log.Fatalf("failed to get torrent info from %s: %s", infoUrl, err)
	}
	defer info.Body.Close()
	_ = info
	body, err := io.ReadAll(info.Body)
	if err != nil {
		log.Fatalf("failed to read body: %s", err)
	}
	items := make([]torrents.Info, 30)
	if err := json.Unmarshal(body, &items); err != nil {
		log.Fatalf("failed to unmarshal torrent info: %s", err)
	}
	return items
}
