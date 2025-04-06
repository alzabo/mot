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

package output

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// TODO: The cache map can grow unbounded. Chances are it doesn't matter much in
// practice, but interning the strings would allow the garbage collector to
// reclaim memory. https://go.dev/blog/unique
func NewTableWriter(output io.Writer) *TableWriter {
	bufWriter := bufio.NewWriter(output)
	w := new(tabwriter.Writer)
	w.Init(bufWriter, 1, 4, 3, ' ', 0)

	tw := TableWriter{
		out:        output,
		cache:      map[string]struct{}{},
		writer:     bufWriter,
		lineWriter: w,
	}
	return &tw
}

type TableWriter struct {
	out        io.Writer
	cache      map[string]struct{}
	writer     *bufio.Writer
	lineWriter *tabwriter.Writer
}

func (tw *TableWriter) Flush() error {
	return errors.Join(
		tw.lineWriter.Flush(),
		tw.writer.Flush(),
	)
}

func (tw *TableWriter) FlushLine() error {
	return tw.lineWriter.Flush()
}

func (tw *TableWriter) Write(fields []string) {
	fmt.Fprintln(tw.lineWriter, tw.join(fields))
}

func (tw *TableWriter) WriteOnce(fields []string) {
	ln := tw.join(fields)
	if _, ok := tw.cache[ln]; ok {
		return
	}
	tw.cache[ln] = struct{}{}
	fmt.Fprintln(tw.lineWriter, ln)
}

func (tw *TableWriter) WriteFunc(fields []string, f func(string) string) {
	if f == nil {
		tw.Write(fields)
	}
	items := make([]string, len(fields))
	for i, j := range fields {
		items[i] = f(j)
	}
	tw.Write(items)
}

func (tw *TableWriter) join(fields []string) string {
	return strings.Join(fields, "\t") + "\t"
}
