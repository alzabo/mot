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
	"net/http"

	"github.com/alzabo/mot/torrent"
	"github.com/goccy/go-json"
)

type MainLog struct {
	Normal   *bool  `param:"normal" usage:"include normal messages (default: true)"`
	Info     *bool  `param:"info" usage:"include info messages (default: true)"`
	Warning  *bool  `param:"warning" usage:"include warning messages (default: true)"`
	Critical *bool  `param:"critical" usage:"include warning messages (default: true)"`
	AfterID  *int64 `param:"last_known_id" usage:"return messages with IDs greater than the provided numeric ID, all messages are retrieved by default"`
}

func (c *Client) Logs(r MainLog) (torrent.Logs, error) {
	path := c.JoinURL("api/v2/log/main")
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	query, err := encode(r)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = query

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatusOK(resp); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read body: %s", err)
	}

	var items torrent.Logs
	if err := json.Unmarshal(buf.Bytes(), &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response with error: %s", err)
	}

	return items, nil
}
