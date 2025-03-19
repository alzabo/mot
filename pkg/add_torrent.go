package mot

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type AddTorrent struct {
	URL                    *[]string `form:"urls" usage:"torrent URL" short:"U"`
	Torrent                *[]string `form:"torrents" usage:"path to torrent file" short:"t"`
	SavePath               *string   `form:"savepath" usage:"torrent download folder"`
	Cookie                 *string   `form:"cookie" usage:"cookie sent to download the .torrent file when URLs are provided"`
	Category               *string   `form:"category" usage:"category for the torrent"`
	Tags                   *[]string `form:"tags" usage:"tags for the torrent"`
	SkipChecking           *bool     `form:"skip_checking" usage:"skip hash checking"`
	Paused                 *bool     `form:"paused" usage:"add torrents in the paused state"`
	CreateRootFolder       *bool     `form:"root_folder" usage:"whether to create the root folder"`
	Rename                 *string   `form:"rename" usage:"rename torrent to the supplied string"`
	UploadLimit            *int64    `form:"upLimit" usage:"set torrent upload speed limit (bytes/second)"`
	DownloadLimit          *int64    `form:"dlLimit" usage:"set torrent download speed limit (bytes/second)"`
	RatioLimit             *float64  `form:"ratioLimit" usage:"set torrent share ratio limit"`
	SeedingTimeLimit       *int64    `form:"seedingTimeLimit" usage:"set torrent seeding time limit (minutes)"`
	AutoTMM                *bool     `form:"autoTMM" usage:"enable Automatic Torrent Management"`
	SequentialDownload     *bool     `form:"sequentialDownload" usage:"enable sequential download"`
	FirstLastPiecePriority *bool     `form:"firstLastPiecePrio" usage:"prioritize downloading first and last pieces"`
}

func (c *Client) AddTorrent(request *AddTorrent) error {
	b := &bytes.Buffer{}
	writer := multipart.NewWriter(b)

	rv := reflect.ValueOf(*request)
	rt := reflect.TypeOf(*request)
	for i := range rt.NumField() {
		v := rv.Field(i)
		if v.IsNil() {
			continue
		}

		form := rt.Field(i).Tag.Get("form")
		switch v.Elem().Kind() {
		case reflect.Bool:
			writer.WriteField(form, fmt.Sprintf("%v", v.Elem().Bool()))
		case reflect.Int64:
			writer.WriteField(form, fmt.Sprintf("%d", v.Elem().Int()))
		case reflect.Float64:
			writer.WriteField(form, fmt.Sprintf("%.2f", v.Elem().Float()))
		case reflect.String:
			writer.WriteField(form, v.Elem().String())
		case reflect.Slice:
			val, ok := (v.Elem().Interface()).([]string)
			if !ok {
				return fmt.Errorf("failed to convert value from field %s", v)
			}
			switch rt.Field(i).Name {
			case "Torrent":
				for _, n := range val {
					f, err := os.Open(n)
					if err != nil {
						return fmt.Errorf("failed to open file %s for reading: %s", n, err)
					}
					defer f.Close()
					torrent, _ := writer.CreateFormFile(form, filepath.Base(n))
					io.Copy(torrent, f)
				}
			case "URL":
				writer.WriteField(form, strings.Join(val, "\n"))
			case "Tags":
				writer.WriteField(form, strings.Join(val, ","))
			}
		default:
			return fmt.Errorf("failed to encode field of type %s", v.Elem().Kind())
		}
	}

	writer.Close()

	//fmt.Println(b.String())

	p, _ := url.JoinPath(c.BaseUrl, "api/v2/torrents/add")
	u, _ := url.Parse(p)
	req := &http.Request{
		URL:    u,
		Host:   u.Host,
		Method: "POST",
		Header: http.Header{
			"Content-Type": {writer.FormDataContentType()},
		},
		Body:          io.NopCloser(b),
		ContentLength: int64(b.Len()),
	}

	//fmt.Println(req)

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error from server: %s", resp.Status)
	}
	fmt.Println(resp)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err)
	}
	if string(body) != "Ok." {
		return fmt.Errorf("unexpected response from server: %s", body)
	}
	return nil
}
