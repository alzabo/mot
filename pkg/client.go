// Copyright 2025 Ryan White
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mot

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
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

var bufPool sync.Pool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

type Client struct {
	HttpClient http.Client
	BaseUrl    string
	Limiter    *rate.Limiter
	backoff    atomic.Int64
}

func NewClient(url string, username string, password string) (*Client, error) {
	var err error
	client := &Client{BaseUrl: url}

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
			ReadBufferSize:      32 << 10,
		},
	}
	err = client.Login(username, password)

	// TODO: Magic numbers here. These limits generally work OK, but sometimes
	// hang. Need to handle hanging requests better, actually back off and
	// recover. Rates as high as 3000 work when the server is lightly loaded,
	// but when the server is performing an operation like checking a torrent,
	// the server may become overloaded.
	client.Limiter = rate.NewLimiter(rate.Limit(2000), 1)

	return client, err
}

func (c *Client) JoinURL(p string) string {
	u, _ := url.JoinPath(c.BaseUrl, p)
	return u
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

	tmp := make([]byte, 4096)
	var buf bytes.Buffer
	io.CopyBuffer(&buf, info.Body, tmp)
	if err != nil {
		log.Fatalf("failed to read body: %s", err)
	}
	items := make([]torrent.Info, 4000) // TODO: magic number here
	if err := json.Unmarshal(buf.Bytes(), &items); err != nil {
		log.Fatalf("failed to unmarshal torrent info: %s", err)
	}
	return items
}

func (c *Client) Files(hash string) torrent.Files {
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
	var files torrent.Files
	if err := json.Unmarshal(body, &files); err != nil {
		log.Fatalf("failed to unmarshal file info: %s", err)
	}
	for i := range files {
		files[i].Hash = hash
	}
	return files
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

func (c *Client) Trackers(hashes []string) torrent.Trackers {
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
	tmp := make([]byte, 4096)
	for _, resp := range resps {
		defer resp.Body.Close()

		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bufPool.Put(buf)

		_, err = io.CopyBuffer(buf, resp.Body, tmp)

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

	return items
}

func (c *Client) sendBatchRequests(reqs []*http.Request) []*http.Response {
	respChan := make(chan *http.Response, len(reqs))
	wg := sync.WaitGroup{}
	wg.Add(len(reqs))
	for _, req := range reqs {
		//fmt.Println(req)
		// Launch a goroutine for each request
		go func(r *http.Request) {
			defer wg.Done()
			// TODO: Handle errors
			//fmt.Println(r.URL)
			resp, err := c.sendRequest(r)
			if err != nil {
				fmt.Println(err)
			}
			respChan <- resp
		}(req)
	}

	wg.Wait()

	responses := make([]*http.Response, len(reqs))
	for i := range len(reqs) {
		responses[i] = <-respChan
	}

	close(respChan)
	return responses
}

func (c *Client) sendRequest(req *http.Request) (*http.Response, error) {
	r := c.Limiter.Reserve()
	time.Sleep(r.Delay())

	var resp *http.Response
	var err error
	ch := make(chan error, 1)

	go func() {
		resp, err = c.HttpClient.Do(req)
		ch <- err
	}()

	select {
	case <-ch:
		//c.Limiter.SetLimit(c.Limiter.Limit() + 100)
		if c.backoff.Load() > 0 {
			c.backoff.Add(-1)
		}
		return resp, err
	case <-time.After(500 * time.Millisecond):
		//c.Limiter.SetLimit(c.Limiter.Limit() - 200)
		// TODO: Log here
		//fmt.Println("timed out", req)
		c.backoff.Add(1)
		time.Sleep(time.Duration(c.backoff.Load() * 1_000_000))
		return c.sendRequest(req)
	}
}

func checkResponseOK(resp *http.Response) error {
	defer resp.Body.Close()
	err := checkStatusOK(resp)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err)
	}
	if string(body) != "Ok." {
		return fmt.Errorf("unexpected response from server: %s", body)
	}
	return nil
}

func checkStatusOK(resp *http.Response) error {
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error from server: %s", resp.Status)
	}
	return nil
}
