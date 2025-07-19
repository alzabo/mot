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

package tokenizer

import (
	"bufio"
	"io"
	"strconv"
)

type Type int

const (
	Unknown = iota
	DictStart
	DictEnd
	ListStart
	ListEnd
	ByteString
	Integer
)

type Token struct {
	Pos   int
	Type  Type
	Bytes []byte
}

type tokenizer struct {
	reader     *bufio.Reader
	pos        int
	tokens     []*Token
	tokenStack []*Token
}

func (t *tokenizer) emit(tok *Token) {
	t.tokens = append(t.tokens, tok)
	if tok.Type == DictStart || tok.Type == ListStart {
		t.stackPush(tok)
	}
}

func (t *tokenizer) stackPush(tok *Token) {
	if t.tokenStack == nil {
		t.tokenStack = make([]*Token, 0, 1)
	}
	t.tokenStack = append(t.tokenStack, tok)
}

func (t *tokenizer) stackPop() *Token {
	if len(t.tokenStack) == 0 {
		panic("no token to pop")
	}
	tok := t.tokenStack[len(t.tokenStack)-1]
	t.tokenStack = t.tokenStack[:len(t.tokenStack)-1]
	return tok
}

func (t *tokenizer) emitString() {
	header, err := t.reader.ReadBytes(':')
	if err != nil {
		panic("malformed string!")
	}
	t.pos += len(header)
	if len(header) < 2 {
		panic("malformed byte string")
	}
	length := header[:len(header)-1]

	l, err := strconv.Atoi(string(length))
	if err != nil {
		panic("Failed to convert header to int")
	}
	bytes := make([]byte, 0, l)
	for range l {
		b, err := t.readByte()
		if err != nil {
			panic("Failed to read string body")
		}
		bytes = append(bytes, b)
	}
	t.emit(&Token{
		Pos:   t.pos,
		Type:  ByteString,
		Bytes: bytes,
	})
}

func (t *tokenizer) readByte() (byte, error) {
	b, err := t.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	t.pos++
	return b, nil
}

// unreadByte unreads the last byte and decrements the position counter.
func (t *tokenizer) unreadByte() {
	t.reader.UnreadByte()
	t.pos--
}

func (t *tokenizer) emitInt() {
	bytes := []byte{}
outer:
	for {
		b, err := t.readByte()
		if err != nil {
			panic("Failed to read int")
		}
		switch b {
		case 'i':
		case 'e':
			break outer
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			bytes = append(bytes, b)
		case '-':
			if len(bytes) > 0 {
				panic("- sign may only appear at the start of an integer")
			}
			bytes = append(bytes, b)
		default:
			panic("malformed integer")
		}
	}
	t.emit(&Token{
		Pos:   t.pos,
		Type:  Integer,
		Bytes: bytes,
	})
}

func Tokenize(r io.Reader) ([]*Token, error) {
	t := tokenizer{
		pos:        -1,
		reader:     bufio.NewReader(r),
		tokens:     []*Token{},
		tokenStack: []*Token{},
	}

	for {
		b, err := t.readByte()
		if err != nil {
			break
		}
		switch b {
		case 'd':
			t.emit(&Token{
				Pos:  t.pos,
				Type: DictStart,
			})
		case 'l':
			t.emit(&Token{
				Pos:  t.pos,
				Type: ListStart,
			})
		case 'i':
			t.emitInt()
		case 'e':
			tok := &Token{
				Pos: t.pos,
			}
			prev := t.stackPop()
			switch prev.Type {
			case DictStart:
				tok.Type = DictEnd
			case ListStart:
				tok.Type = ListEnd
			default:
				panic("invalid stack")
			}
			t.emit(tok)
		default:
			if b >= '0' && b <= '9' {
				// emitString has to consume the full string identifier, so the
				// last read byte must be unread to allow the string to be
				// parsed.
				t.unreadByte()
				t.emitString()
			}
		}
	}
	return t.tokens, nil
}
