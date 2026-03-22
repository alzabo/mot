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

package parser

import (
	"bytes"
	"testing"
)

func TestParseStreamSingle(t *testing.T) {
	data := []byte("d8:announce11:tracker.url4:infod4:name12:test.torrent6:lengthi12345eee")

	var torrents []Torrent
	err := ParseStream(bytes.NewReader(data), func(t Torrent) {
		torrents = append(torrents, t)
	})

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	if len(torrents) != 1 {
		t.Fatalf("expected 1 torrent, got %d", len(torrents))
	}

	if torrents[0].Name != "test.torrent" {
		t.Errorf("expected name 'test.torrent', got %q", torrents[0].Name)
	}

	if torrents[0].Size != 12345 {
		t.Errorf("expected size 12345, got %d", torrents[0].Size)
	}
}

func TestParseStreamMultiple(t *testing.T) {
	data := []byte("d8:announce11:tracker.url4:infod4:name13:first.torrent6:lengthi100eeed8:announce11:tracker.url4:infod4:name14:second.torrent6:lengthi200eee")

	var torrents []Torrent
	err := ParseStream(bytes.NewReader(data), func(t Torrent) {
		torrents = append(torrents, t)
	})

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	if len(torrents) != 2 {
		t.Fatalf("expected 2 torrents, got %d", len(torrents))
	}

	if torrents[0].Name != "first.torrent" {
		t.Errorf("expected first name 'first.torrent', got %q", torrents[0].Name)
	}

	if torrents[0].Size != 100 {
		t.Errorf("expected first size 100, got %d", torrents[0].Size)
	}

	if torrents[1].Name != "second.torrent" {
		t.Errorf("expected second name 'second.torrent', got %q", torrents[1].Name)
	}

	if torrents[1].Size != 200 {
		t.Errorf("expected second size 200, got %d", torrents[1].Size)
	}
}

func TestParseStreamHash(t *testing.T) {
	data := []byte("d4:infod4:name12:test.torrent6:lengthi100eee")

	var torrents []Torrent
	err := ParseStream(bytes.NewReader(data), func(t Torrent) {
		torrents = append(torrents, t)
	})

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	if len(torrents) != 1 {
		t.Fatalf("expected 1 torrent, got %d", len(torrents))
	}

	if torrents[0].Hash == "" {
		t.Error("expected non-empty hash")
	}

	if len(torrents[0].Hash) != 40 {
		t.Errorf("expected 40 char hash, got %d chars", len(torrents[0].Hash))
	}
}
