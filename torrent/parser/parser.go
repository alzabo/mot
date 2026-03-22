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
	"crypto/sha1"
	"fmt"
	"io"
	"strconv"

	"github.com/alzabo/mot/torrent/tokenizer"
)

type Torrent struct {
	Announce string
	Name     string
	Comment  string
	Hash     string
	Size     int64
	Private  bool
}

type indexes struct {
	start int
	end   int
}

var (
	announceBytes = []byte("announce")
	commentBytes  = []byte("comment")
	infoBytes     = []byte("info")
	nameBytes     = []byte("name")
	privateBytes  = []byte("private")
	lengthBytes   = []byte("length")
)

func Parse(r io.ReadSeeker) (Torrent, error) {
	torrent := Torrent{}
	dictStack := []tokenizer.Type{}
	info := indexes{}

	tokens, err := tokenizer.Tokenize(r)
	if err != nil {
		return torrent, fmt.Errorf("tokenizing: %w", err)
	}
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		switch t.Type {
		case tokenizer.ByteString:
			if i+1 >= len(tokens) {
				return torrent, fmt.Errorf("unexpected end of tokens")
			}
			next := tokens[i+1]
			if bytes.Equal(t.Bytes, infoBytes) {
				if len(dictStack) != 1 {
					continue
				}
				if next.Type != tokenizer.DictStart {
					return torrent, fmt.Errorf("expected dict after info key, got %v", next.Type)
				}
				info.start = next.Pos
			} else if bytes.Equal(t.Bytes, announceBytes) {
				if next.Type != tokenizer.ByteString {
					return torrent, fmt.Errorf("expected string after announce key, got %v", next.Type)
				}
				torrent.Announce = string(next.Bytes)
				i++
			} else if bytes.Equal(t.Bytes, commentBytes) {
				if next.Type != tokenizer.ByteString {
					return torrent, fmt.Errorf("expected string after comment key, got %v", next.Type)
				}
				torrent.Comment = string(next.Bytes)
				i++
			} else if bytes.Equal(t.Bytes, nameBytes) {
				if next.Type != tokenizer.ByteString {
					return torrent, fmt.Errorf("expected string after name key, got %v", next.Type)
				}
				torrent.Name = string(next.Bytes)
				i++
			} else if bytes.Equal(t.Bytes, privateBytes) {
				if next.Type != tokenizer.Integer {
					return torrent, fmt.Errorf("expected integer after private key, got %v", next.Type)
				}
				b, err := strconv.ParseBool(string(next.Bytes))
				if err != nil {
					return torrent, fmt.Errorf("parsing private: %w", err)
				}
				torrent.Private = b
				i++
			} else if bytes.Equal(t.Bytes, lengthBytes) {
				if next.Type != tokenizer.Integer {
					return torrent, fmt.Errorf("expected integer after length key, got %v", next.Type)
				}
				s, err := strconv.ParseInt(string(next.Bytes), 10, 64)
				if err != nil {
					return torrent, fmt.Errorf("parsing length: %w", err)
				}
				torrent.Size += s
				i++
			}
		case tokenizer.DictStart:
			dictStack = append(dictStack, t.Type)
		case tokenizer.DictEnd:
			dictStack = dictStack[:len(dictStack)-1]
			if len(dictStack) == 1 && info.start != 0 && info.end == 0 {
				info.end = t.Pos
			}
		}
	}

	torrent.Hash = infoHash(r, info)
	return torrent, nil
}

func infoHash(r io.ReadSeeker, i indexes) string {
	r.Seek(0, 0)
	b, _ := io.ReadAll(r)
	hash := sha1.Sum(b[i.start : i.end+1])
	return fmt.Sprintf("%x", hash)
}

type parseState struct {
	torrent    Torrent
	dictDepth  int
	infoDepth  int
	infoBuffer *bytes.Buffer
	recording  bool
	err        error
}

func ParseStream(r io.Reader, onTorrent func(Torrent)) error {
	t := tokenizer.New(r)
	state := &parseState{
		infoBuffer: &bytes.Buffer{},
	}

	for {
		tok, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tokenization error: %w", err)
		}

		processToken(state, tok, t, onTorrent)
		if state.err != nil {
			return state.err
		}
	}

	if state.torrent.Name != "" || state.torrent.Hash != "" || state.torrent.Size > 0 {
		onTorrent(state.torrent)
	}

	return nil
}

func processToken(state *parseState, tok *tokenizer.Token, t *tokenizer.Tokenizer, onTorrent func(Torrent)) {
	switch tok.Type {
	case tokenizer.ByteString:
		if bytes.Equal(tok.Bytes, infoBytes) {
			state.recording = true
			state.infoBuffer.Reset()
			return
		}

		if state.recording && state.dictDepth == 1 {
			switch {
			case bytes.Equal(tok.Bytes, announceBytes):
				state.torrent.Announce, state.err = readNextString(t)
			case bytes.Equal(tok.Bytes, commentBytes):
				state.torrent.Comment, state.err = readNextString(t)
			}
			if state.err != nil {
				return
			}
		}

		if state.recording && state.dictDepth == 2 {
			switch {
			case bytes.Equal(tok.Bytes, nameBytes):
				state.torrent.Name, state.err = readNextString(t)
			case bytes.Equal(tok.Bytes, privateBytes):
				s, err := readNextString(t)
				if err != nil {
					state.err = err
					return
				}
				state.torrent.Private, state.err = strconv.ParseBool(s)
			case bytes.Equal(tok.Bytes, lengthBytes):
				state.torrent.Size, state.err = readNextInt(t)
			}
			if state.err != nil {
				return
			}
		}

		if state.infoDepth > 0 {
			state.infoBuffer.Write(tok.Bytes)
		}

	case tokenizer.DictStart:
		state.dictDepth++
		if state.recording {
			state.infoDepth = state.dictDepth
		}
		if state.infoDepth > 0 {
			state.infoBuffer.WriteByte('d')
		}

	case tokenizer.DictEnd:
		if state.infoDepth > 0 {
			state.infoBuffer.WriteByte('e')
			if state.infoDepth == state.dictDepth {
				state.torrent.Hash = fmt.Sprintf("%x", sha1.Sum(state.infoBuffer.Bytes()))
				state.infoDepth = 0
			}
		}
		state.dictDepth--

		if state.dictDepth == 0 {
			onTorrent(state.torrent)
			state.torrent = Torrent{}
			state.infoBuffer.Reset()
			state.recording = false
		}

	case tokenizer.ListStart:
		state.dictDepth++
		if state.infoDepth > 0 {
			state.infoBuffer.WriteByte('l')
		}

	case tokenizer.ListEnd:
		state.dictDepth--
		if state.infoDepth > 0 {
			state.infoBuffer.WriteByte('e')
		}

	case tokenizer.Integer:
		if state.infoDepth > 0 {
			state.infoBuffer.WriteByte('i')
			state.infoBuffer.Write(tok.Bytes)
			state.infoBuffer.WriteByte('e')
		}
		if state.recording && state.dictDepth == 2 {
			if bytes.Equal(getLastKey(), lengthBytes) {
				state.torrent.Size, state.err = strconv.ParseInt(string(tok.Bytes), 10, 64)
			}
		}
	}
}

var lastKey []byte

func getLastKey() []byte {
	return lastKey
}

func readNextString(t *tokenizer.Tokenizer) (string, error) {
	tok, err := t.Next()
	if err != nil {
		return "", fmt.Errorf("reading next token: %w", err)
	}
	lastKey = tok.Bytes
	if tok.Type != tokenizer.ByteString {
		return "", fmt.Errorf("expected string, got %v", tok.Type)
	}
	return string(tok.Bytes), nil
}

func readNextInt(t *tokenizer.Tokenizer) (int64, error) {
	tok, err := t.Next()
	if err != nil {
		return 0, fmt.Errorf("reading next token: %w", err)
	}
	lastKey = tok.Bytes
	if tok.Type != tokenizer.Integer {
		return 0, fmt.Errorf("expected integer, got %v", tok.Type)
	}
	return strconv.ParseInt(string(tok.Bytes), 10, 64)
}
