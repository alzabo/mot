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

// DO NOT EDIT
package torrent

type Properties struct {
	AdditionDate           int64   `json:"addition_date"`
	Comment                string  `json:"comment"`
	CompletionDate         int64   `json:"completion_date"`
	CreatedBy              string  `json:"created_by"`
	CreationDate           int64   `json:"creation_date"`
	DLLimit                int64   `json:"dl_limit"`
	DLSpeed                int64   `json:"dl_speed"`
	DLSpeedAvg             int64   `json:"dl_speed_avg"`
	DownloadPath           string  `json:"download_path"`
	ETA                    int64   `json:"eta"`
	Hash                   string  `json:"hash"`
	InfohashV1             string  `json:"infohash_v1"`
	InfohashV2             string  `json:"infohash_v2"`
	IsPrivate              bool    `json:"is_private"`
	LastSeen               int64   `json:"last_seen"`
	Name                   string  `json:"name"`
	NBConnections          int64   `json:"nb_connections"`
	NBConnectionsLimit     int64   `json:"nb_connections_limit"`
	Peers                  int64   `json:"peers"`
	PeersTotal             int64   `json:"peers_total"`
	PieceSize              int64   `json:"piece_size"`
	PiecesHave             int64   `json:"pieces_have"`
	PiecesNum              int64   `json:"pieces_num"`
	Reannounce             int64   `json:"reannounce"`
	SavePath               string  `json:"save_path"`
	SeedingTime            int64   `json:"seeding_time"`
	Seeds                  int64   `json:"seeds"`
	SeedsTotal             int64   `json:"seeds_total"`
	ShareRatio             float64 `json:"share_ratio"`
	TimeElapsed            int64   `json:"time_elapsed"`
	TotalDownloaded        int64   `json:"total_downloaded"`
	TotalDownloadedSession int64   `json:"total_downloaded_session"`
	TotalSize              int64   `json:"total_size"`
	TotalUploaded          int64   `json:"total_uploaded"`
	TotalUploadedSession   int64   `json:"total_uploaded_session"`
	TotalWasted            int64   `json:"total_wasted"`
	UpLimit                int64   `json:"up_limit"`
	UpSpeed                int64   `json:"up_speed"`
	UpSpeedAvg             int64   `json:"up_speed_avg"`
}
