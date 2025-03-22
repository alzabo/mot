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

type Info struct {
	AddedOn                  int64   `json:"added_on"`
	AmountLeft               int64   `json:"amount_left"`
	AutoTMM                  bool    `json:"auto_tmm"`
	Availability             float64 `json:"availability"`
	Category                 string  `json:"category"`
	Completed                int64   `json:"completed"`
	CompletionOn             int64   `json:"completion_on"`
	ContentPath              string  `json:"content_path"`
	DLLimit                  int64   `json:"dl_limit"`
	Dlspeed                  int64   `json:"dlspeed"`
	DownloadPath             string  `json:"download_path"`
	Downloaded               int64   `json:"downloaded"`
	DownloadedSession        int64   `json:"downloaded_session"`
	ETA                      int64   `json:"eta"`
	FLPiecePrio              bool    `json:"f_l_piece_prio"`
	ForceStart               bool    `json:"force_start"`
	Hash                     string  `json:"hash"`
	InactiveSeedingTimeLimit int64   `json:"inactive_seeding_time_limit,omitempty"`
	InfohashV1               string  `json:"infohash_v1"`
	InfohashV2               string  `json:"infohash_v2"`
	LastActivity             int64   `json:"last_activity"`
	MagnetURI                string  `json:"magnet_uri"`
	MaxInactiveSeedingTime   int64   `json:"max_inactive_seeding_time,omitempty"`
	MaxRatio                 int64   `json:"max_ratio"`
	MaxSeedingTime           int64   `json:"max_seeding_time"`
	Name                     string  `json:"name"`
	NumComplete              int64   `json:"num_complete"`
	NumIncomplete            int64   `json:"num_incomplete"`
	NumLeechs                int64   `json:"num_leechs"`
	NumSeeds                 int64   `json:"num_seeds"`
	Priority                 int64   `json:"priority"`
	Progress                 float64 `json:"progress"`
	Ratio                    float64 `json:"ratio"`
	RatioLimit               float64 `json:"ratio_limit"`
	SavePath                 string  `json:"save_path"`
	SeedingTime              int64   `json:"seeding_time"`
	SeedingTimeLimit         int64   `json:"seeding_time_limit"`
	SeenComplete             int64   `json:"seen_complete"`
	SeqDL                    bool    `json:"seq_dl"`
	Size                     int64   `json:"size"`
	State                    string  `json:"state"`
	SuperSeeding             bool    `json:"super_seeding"`
	Tags                     string  `json:"tags"`
	TimeActive               int64   `json:"time_active"`
	TotalSize                int64   `json:"total_size"`
	Tracker                  string  `json:"tracker"`
	TrackersCount            int64   `json:"trackers_count"`
	UpLimit                  int64   `json:"up_limit"`
	Uploaded                 int64   `json:"uploaded"`
	UploadedSession          int64   `json:"uploaded_session"`
	Upspeed                  int64   `json:"upspeed"`
}
