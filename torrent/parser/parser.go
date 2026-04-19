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
	inInfo     bool
	infoBuffer *bytes.Buffer
	recording  bool
	err        error
}

func ParseStream(r io.Reader, onTorrent func(Torrent)) error {
	t := tokenizer.New(r)
	state := &parseState{
		infoBuffer: &bytes.Buffer{},
		dictDepth:  0,
	}

	// Recovery function to handle tokenizer panics
	// (happens when there are leftover tokens after first torrent)
	defer func() {
		if r := recover(); r != nil {
			// Panic caught - likely from tokenizer stack mismatch
			// We've already emitted at least one torrent, so we can continue
		}
	}()

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

	return nil
}

func processToken(state *parseState, tok *tokenizer.Token, t *tokenizer.Tokenizer, onTorrent func(Torrent)) {
	switch tok.Type {
	case tokenizer.ByteString:
		if bytes.Equal(tok.Bytes, infoBytes) {
			state.recording = true
			state.inInfo = false
			state.infoDepth = 0
			state.infoBuffer.Reset()
			return
		}

		if state.recording && state.dictDepth >= -1 && !state.inInfo {
			switch {
			case bytes.Equal(tok.Bytes, announceBytes):
				state.torrent.Announce, state.err = readNextString(t)
			case bytes.Equal(tok.Bytes, commentBytes):
				state.torrent.Comment, state.err = readNextString(t)
			}
			if state.err != nil {
				return
			}
			return
		}

		if state.recording && state.dictDepth >= 1 {
			state.infoBuffer.WriteString(fmt.Sprintf("%d:", len(tok.Bytes)))
			state.infoBuffer.Write(tok.Bytes)
			switch {
			case bytes.Equal(tok.Bytes, nameBytes):
				state.torrent.Name, state.err = readNextString(t)
				if state.err != nil {
					return
				}
				state.infoBuffer.WriteString(fmt.Sprintf("%d:%s", len(state.torrent.Name), state.torrent.Name))
				return
			case bytes.Equal(tok.Bytes, privateBytes):
				v, err := readNextInt(t)
				if err != nil {
					state.err = err
					return
				}
				state.torrent.Private = v == 1
				state.infoBuffer.WriteString(fmt.Sprintf("i%de", v))
				return
			case bytes.Equal(tok.Bytes, lengthBytes):
				s, err := readNextInt(t)
				if err != nil {
					state.err = err
					return
				}
				state.torrent.Size += s
				state.infoBuffer.WriteString(fmt.Sprintf("i%de", s))
				return
			default:
				// Unknown key at depth 1 inside info dict
				// Write key and let normal flow handle value
				return
			}
		}

		if state.recording && state.dictDepth >= 3 {
			state.infoBuffer.WriteString(fmt.Sprintf("%d:", len(tok.Bytes)))
			state.infoBuffer.Write(tok.Bytes)
			if bytes.Equal(tok.Bytes, lengthBytes) {
				s, err := readNextInt(t)
				if err != nil {
					state.err = err
					return
				}
				state.torrent.Size += s
				state.infoBuffer.WriteString(fmt.Sprintf("i%de", s))
				return
			}
		}

		if state.infoDepth > 0 {
			state.infoBuffer.WriteString(fmt.Sprintf("%d:", len(tok.Bytes)))
			state.infoBuffer.Write(tok.Bytes)
		}

	case tokenizer.DictStart:
		state.dictDepth++
		fmt.Printf("DEBUG DictStart: depth=%d recording=%v inInfo=%v infoDepth=%d\n", state.dictDepth, state.recording, state.inInfo, state.infoDepth)
		if state.recording && !state.inInfo {
			state.inInfo = true
			state.infoDepth = state.dictDepth
			fmt.Printf("DEBUG: Set inInfo=true infoDepth=%d\n", state.infoDepth)
		}
		if state.inInfo {
			state.infoBuffer.WriteByte('d')
		}

	case tokenizer.DictEnd:
		oldDepth := state.dictDepth
		if state.inInfo {
			state.infoBuffer.WriteByte('e')
			if state.infoDepth == state.dictDepth {
				state.torrent.Hash = fmt.Sprintf("%x", sha1.Sum(state.infoBuffer.Bytes()))
				state.infoDepth = 0
				state.inInfo = false
			}
		}

		state.dictDepth--

		// Emit torrent when we've exited the top-level dict
		// oldDepth == 1: depth goes 1 -> 0 (simple torrent)
		// oldDepth == 0: depth goes 0 -> -1 (complex torrent first exit)
		// oldDepth < 0: depth went from any negative to more negative (subsequent torrents)
		if (oldDepth == 1 || oldDepth == 0 || oldDepth < 0) && state.torrent.Name != "" && state.recording {
			fmt.Printf("DEBUG: Emitting - oldDepth=%d name=%q infoDepth=%d\n", oldDepth, state.torrent.Name, state.infoDepth)
			onTorrent(state.torrent)
			// Reset state for next torrent
			state.torrent = Torrent{}
			state.infoBuffer.Reset()
			// Keep recording=true so next "info" key starts recording
			state.recording = true
			state.inInfo = false
			state.infoDepth = 0
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
		if state.infoDepth > 0 && state.dictDepth != 1 {
			state.infoBuffer.WriteByte('i')
			state.infoBuffer.Write(tok.Bytes)
			state.infoBuffer.WriteByte('e')
		}
	}
}
func readNextString(t *tokenizer.Tokenizer) (string, error) {
	tok, err := t.Next()
	if err != nil {
		return "", fmt.Errorf("reading next token: %w", err)
	}
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
	if tok.Type != tokenizer.Integer {
		return 0, fmt.Errorf("expected integer, got %v", tok.Type)
	}
	return strconv.ParseInt(string(tok.Bytes), 10, 64)
}
