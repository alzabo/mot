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
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// AddTorrent represents the qBittorrent Torrent add API payload.
type AddTorrent struct {
	URL                          *[]string `param:"urls" join:"\n" usage:"torrent URL" short:"U"`
	Torrent                      *[]string `param:"torrents" usage:"path to torrent file" short:"f"`
	SavePath                     *string   `param:"savepath" usage:"torrent download folder"`
	Cookie                       *string   `param:"cookie" usage:"cookie sent to download the .torrent file when URLs are provided"`
	Category                     *string   `param:"category" usage:"category for the torrent"`
	Tags                         *[]string `param:"tags" join:"," usage:"tags for the torrent"`
	SkipChecking                 *bool     `param:"skip_checking" usage:"skip hash checking"`
	Paused                       *bool     `param:"paused" usage:"add torrents in the paused state"`
	CreateRootFolder             *bool     `param:"root_folder" usage:"whether to create the root folder"`
	NameOverride                 *string   `param:"rename" usage:"set torrent name to the supplied string instead of the value in the torrent"`
	UploadLimit                  *int64    `param:"upLimit" usage:"set torrent upload speed limit (bytes/second)"`
	DownloadLimit                *int64    `param:"dlLimit" usage:"set torrent download speed limit (bytes/second)"`
	RatioLimit                   *float64  `param:"ratioLimit" usage:"set torrent share ratio limit"`
	SeedingTimeLimit             *int64    `param:"seedingTimeLimit" usage:"set torrent seeding time limit (minutes)"`
	EnableAutoTMM                *bool     `param:"autoTMM" usage:"enable Automatic Torrent Management"`
	EnableSequentialDownload     *bool     `param:"sequentialDownload" usage:"enable sequential download"`
	EnableFirstLastPiecePriority *bool     `param:"firstLastPiecePrio" usage:"prioritize downloading first and last pieces"`
}

func (r *AddTorrent) ParseMultiPartForm() (io.Reader, string, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	rv := reflect.ValueOf(*r)
	rt := reflect.TypeOf(*r)
	for i := range rt.NumField() {
		v := rv.Field(i)
		if v.IsNil() {
			continue
		}

		form := rt.Field(i).Tag.Get("param")
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
				return nil, "", fmt.Errorf("failed to convert value from field %s", v)
			}

			join := rt.Field(i).Tag.Get("join")
			switch rt.Field(i).Name {
			// TODO: change Torrent to [][]byte, read in the CLI.
			case "Torrent":
				for _, n := range val {
					f, err := os.Open(n)
					if err != nil {
						return nil, "", fmt.Errorf("failed to open file %s for reading: %s", n, err)
					}
					defer f.Close()
					torrent, _ := writer.CreateFormFile(form, filepath.Base(n))
					io.Copy(torrent, f)
				}
			default:
				writer.WriteField(form, strings.Join(val, join))
			}
		default:
			return nil, "", fmt.Errorf("failed to encode field of type %s", v.Elem().Kind())
		}
	}
	writer.Close()

	return buf, writer.FormDataContentType(), nil
}

func (c *Client) AddTorrent(r *AddTorrent) error {
	buf, contentType, err := r.ParseMultiPartForm()
	if err != nil {
		return err
	}
	path := c.JoinURL("api/v2/torrents/add")
	req, err := http.NewRequest("POST", path, buf)
	if err != nil {
		return err
	}
	req.Header["Content-Type"] = []string{contentType}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	return checkResponseOK(resp)
}

func checkResponseOK(resp *http.Response) error {
	defer resp.Body.Close()
	fmt.Println(resp.Header)
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error from server: %s", resp.Status)
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
