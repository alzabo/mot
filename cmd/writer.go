package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func writer(output io.Writer) tableWriter {
	bufWriter := bufio.NewWriter(output)
	w := new(tabwriter.Writer)
	w.Init(bufWriter, 1, 4, 3, ' ', 0)

	tw := tableWriter{
		out:        output,
		cache:      map[string]struct{}{},
		writer:     bufWriter,
		lineWriter: w,
	}
	return tw
}

type tableWriter struct {
	out        io.Writer
	cache      map[string]struct{}
	writer     *bufio.Writer
	lineWriter *tabwriter.Writer
}

func (tw *tableWriter) Flush() error {
	return errors.Join(
		tw.lineWriter.Flush(),
		tw.writer.Flush(),
	)
}

func (tw *tableWriter) FlushLine() error {
	return tw.lineWriter.Flush()
}

func (tw *tableWriter) Write(fields []string) {
	fmt.Fprintln(tw.lineWriter, tw.join(fields))
}

func (tw *tableWriter) WriteOnce(fields []string) {
	ln := tw.join(fields)
	if _, ok := tw.cache[ln]; ok {
		return
	}
	tw.cache[ln] = struct{}{}
	fmt.Fprintln(tw.lineWriter, tw.join(fields))
}

func (tw *tableWriter) WriteFunc(fields []string, f func(string) string) {
	if f == nil {
		tw.Write(fields)
	}
	items := make([]string, len(fields))
	for i, j := range fields {
		items[i] = f(j)
	}
	tw.Write(items)
}

func (tw *tableWriter) join(fields []string) string {
	return fmt.Sprintf("%s\t", strings.Join(fields, "\t"))
}
