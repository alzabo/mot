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

func Parse(r io.ReadSeeker) Torrent {
	torrent := Torrent{}
	dictStack := []tokenizer.Type{}
	info := indexes{}

	tokens, err := tokenizer.Tokenize(r)
	if err != nil {
		panic("error encountered while tokenizing file" + err.Error())
	}
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		switch t.Type {
		case tokenizer.ByteString:
			if i+1 >= len(tokens) {
				panic("final token cannot be ByteString")
			}
			next := tokens[i+1]
			if bytes.Equal(t.Bytes, infoBytes) {
				// A torrent's bencode representation will always be represented
				// as a top-level dict. the info dict will start with a
				// ByteString key "info", followed by a nested dict with a dict
				if len(dictStack) != 1 {
					continue
				}
				if next.Type != tokenizer.DictStart {
					panic("expected dict start after info key")
				}
				info.start = next.Pos
			} else if bytes.Equal(t.Bytes, announceBytes) {
				if next.Type != tokenizer.ByteString {
					panic("expected ByteString after announce key")
				}
				torrent.Announce = string(next.Bytes)
				i++
			} else if bytes.Equal(t.Bytes, commentBytes) {
				if next.Type != tokenizer.ByteString {
					panic("expected ByteString after comment key")
				}
				torrent.Comment = string(next.Bytes)
				i++
			} else if bytes.Equal(t.Bytes, nameBytes) {
				if next.Type != tokenizer.ByteString {
					panic("expected ByteString after name key")
				}
				torrent.Name = string(next.Bytes)
				i++
			} else if bytes.Equal(t.Bytes, privateBytes) {
				if next.Type != tokenizer.Integer {
					panic("expected Integer after private key")
				}
				b, _ := strconv.ParseBool(string(next.Bytes)) // TODO: Check error
				torrent.Private = b
				i++
			} else if bytes.Equal(t.Bytes, lengthBytes) {
				if next.Type != tokenizer.Integer {
					panic("expected Integer after size key")
				}
				s, _ := strconv.ParseInt(string(next.Bytes), 10, 64) // TODO: Check error
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
	return torrent
}

func infoHash(r io.ReadSeeker, i indexes) string {
	r.Seek(0, 0)
	b, _ := io.ReadAll(r)
	hash := sha1.Sum(b[i.start : i.end+1])
	return fmt.Sprintf("%x", hash)
}
