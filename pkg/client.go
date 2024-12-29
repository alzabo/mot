package mot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

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

func NewClient(url string, username string, password string) Client {
	client := Client{}
	client.BaseUrl = url
	// All users of cookiejar should import "golang.org/x/net/publicsuffix"
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	client.HttpClient = http.Client{
		Jar: jar,
	}
	client.Login(username, password)
	return client
}

func (c *Client) Login(username, password string) {
	loginUrl, err := url.JoinPath(c.BaseUrl, login)
	if err != nil {
		log.Fatalf("failed to create login URL: %s", err)
	}
	params := url.Values{
		"username": {username},
		"password": {password},
	}
	resp, err := c.HttpClient.PostForm(loginUrl, params)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	if string(r) != "Ok." {
		log.Fatalf("failed to log in with message: %s", r)
	}
}

func (c *Client) TorrentList() []torrents.Info {
	infoApi, err := url.JoinPath(c.BaseUrl, torrentInfo)
	if err != nil {
		log.Fatalf("failed to create url: %s", err)
	}
	infoUrl, _ := url.Parse(infoApi)
	params := url.Values{
		"foo": {"bar"},
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
	items := []torrents.Info{}
	if err := json.Unmarshal(body, &items); err != nil {
		log.Fatalf("failed to unmarshal torrent info: %s", err)
	}
	return items
}
