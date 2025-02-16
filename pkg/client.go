package mot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"golang.org/x/net/publicsuffix"
	"golang.org/x/time/rate"

	"github.com/alzabo/mot/torrent"
)

const (
	login             = "/api/v2/auth/login"
	torrentInfo       = "/api/v2/torrents/info"
	torrentHashHeader = "X-Torrent-Hash"
)

type Client struct {
	HttpClient http.Client
	BaseUrl    string
	Limiter    *rate.Limiter
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
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}
	err = client.Login(username, password)

	// TODO: Magic numbers here. These limits generally work OK, but sometimes hang.
	// Need to handle hanging requests better, actually back off and recover.
	client.Limiter = rate.NewLimiter(rate.Limit(1500), 300)

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

// TODO: return a struct that implements an interface
func WithHashes(hashes []string) QueryOption {
	return func(v *url.Values) {
		v.Set("hashes", strings.Join(hashes, "|"))
	}
}

func WithValue(key, val string) QueryOption {
	return func(v *url.Values) {
		v.Set(key, val)
	}
}

func (c *Client) Torrents(opts ...QueryOption) []torrent.Info {
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
	var buf bytes.Buffer
	io.Copy(&buf, info.Body)
	if err != nil {
		log.Fatalf("failed to read body: %s", err)
	}
	items := make([]torrent.Info, 4000) // TODO: magic number here
	if err := json.Unmarshal(buf.Bytes(), &items); err != nil {
		log.Fatalf("failed to unmarshal torrent info: %s", err)
	}
	return items
}

func (c *Client) Files(hash string) []torrent.Values {
	p, err := url.JoinPath(c.BaseUrl, "api/v2/torrents/files")
	if err != nil {
		log.Fatalf("failed to create url: %s", err)
	}
	u, _ := url.Parse(p)
	u.RawQuery = url.Values{"hash": {hash}}.Encode()
	g, err := c.HttpClient.Get(u.String())
	if err != nil {
		log.Fatalf("failed to get files for hash %s: %s", hash, err)
	}
	defer g.Body.Close()

	body, err := io.ReadAll(g.Body)
	if err != nil {
		log.Fatalf("failed to read body: %s", err)
	}
	var items torrent.Files
	if err := json.Unmarshal(body, &items); err != nil {
		log.Fatalf("failed to unmarshal file info: %s", err)
	}
	values := make([]torrent.Values, len(items))
	for i, item := range items {
		values[i] = item.Values(map[string]string{"hash": hash})
	}
	return values
}

// TODO: The API for these is inconsistent. Other methods panic, this returns error
func (c *Client) Recheck(opts ...QueryOption) error {
	p, _ := url.JoinPath(c.BaseUrl, "api/v2/torrents/recheck")
	u, _ := url.Parse(p)
	params := url.Values{}
	for _, opt := range opts {
		opt(&params)
	}
	req, err := c.HttpClient.PostForm(u.String(), params)
	if err != nil {
		return fmt.Errorf("failed to recheck with url %s: %s", u, err)
	}
	defer req.Body.Close()

	return nil
}

// TODO: This only works on WithHashes
func (c *Client) Resume(opts ...QueryOption) error {
	p, _ := url.JoinPath(c.BaseUrl, "api/v2/torrents/resume")
	u, _ := url.Parse(p)
	params := url.Values{}
	for _, opt := range opts {
		opt(&params)
	}
	g, err := c.HttpClient.PostForm(u.String(), params)
	if err != nil {
		return fmt.Errorf("failed to resume torrents with url %s: %s", u, err)
	}
	defer g.Body.Close()

	r, _ := io.ReadAll(g.Body)
	if len(r) > 0 {
		return fmt.Errorf("received unexpected response: %s", r)
	}

	return nil
}

// TODO: This only works on WithHashes
func (c *Client) Pause(opts ...QueryOption) error {
	p, _ := url.JoinPath(c.BaseUrl, "api/v2/torrents/pause")
	u, _ := url.Parse(p)
	params := url.Values{}
	for _, opt := range opts {
		opt(&params)
	}
	g, err := c.HttpClient.PostForm(u.String(), params)
	if err != nil {
		return fmt.Errorf("failed to resume torrents with url %s: %s", u, err)
	}
	defer g.Body.Close()

	r, _ := io.ReadAll(g.Body)
	if len(r) > 0 {
		return fmt.Errorf("received unexpected response: %s", r)
	}

	return nil
}

func (c *Client) DeleteTorrents(opts ...QueryOption) error {
	p, _ := url.JoinPath(c.BaseUrl, "api/v2/torrents/delete")
	u, _ := url.Parse(p)
	params := url.Values{}
	for _, opt := range opts {
		opt(&params)
	}
	g, err := c.HttpClient.PostForm(u.String(), params)
	if err != nil {
		return fmt.Errorf("failed to delete torrents with url %s: %s", u, err)
	}
	defer g.Body.Close()

	r, _ := io.ReadAll(g.Body)
	if len(r) > 0 {
		return fmt.Errorf("received unexpected response: %s", r)
	}

	return nil
}

func (c *Client) Trackers(hashes []string) []torrent.Values {
	var err error
	p, err := url.JoinPath(c.BaseUrl, "api/v2/torrents/trackers")
	if err != nil {
		log.Fatalf("failed to create url: %s", err)
	}
	reqs := make([]*http.Request, len(hashes))
	for i, h := range hashes {
		u, _ := url.Parse(p)
		u.RawQuery = url.Values{"hash": {h}}.Encode()
		reqs[i] = &http.Request{
			Method: "GET",
			URL:    u,
			Header: http.Header{
				torrentHashHeader: []string{h},
			},
		}
	}

	items := make(torrent.Trackers, 0, len(hashes))
	resps := c.sendBatchRequests(reqs)
	for _, resp := range resps {
		defer resp.Body.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, resp.Body)

		if err != nil {
			log.Fatalf("failed to read body: %s", err)
		}

		var trackers torrent.Trackers
		if err := json.Unmarshal(buf.Bytes(), &trackers); err != nil {
			log.Fatalf("failed to unmarshal response with error: %s", err)
		}

		for _, t := range trackers {
			t.Hash = resp.Request.Header.Get(torrentHashHeader)
			items = append(items, t)
		}
	}

	values := make([]torrent.Values, len(items))
	for i, item := range items {
		values[i] = item.Values(map[string]string{})
	}
	return values
}

func (c *Client) sendBatchRequests(reqs []*http.Request) []*http.Response {
	respChan := make(chan *http.Response, len(reqs))
	ctx := context.Background()

	for _, req := range reqs {
		//fmt.Println(req)
		c.Limiter.Wait(ctx)
		// Launch a goroutine for each request
		go func(r *http.Request) {
			// TODO: Handle errors
			resp, _ := c.sendRequest(r)
			respChan <- resp
		}(req)
	}

	responses := make([]*http.Response, len(reqs))
	for i := range len(reqs) {
		responses[i] = <-respChan
	}

	close(respChan)
	return responses
}

func (c *Client) sendRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
