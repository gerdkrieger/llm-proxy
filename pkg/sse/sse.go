// Package sse provides utilities for Server-Sent Events (SSE).
package sse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Event represents a Server-Sent Event
type Event struct {
	ID    string
	Event string
	Data  string
	Retry int
}

// Format formats the event as an SSE message
func (e *Event) Format() string {
	var buf bytes.Buffer

	if e.ID != "" {
		buf.WriteString(fmt.Sprintf("id: %s\n", e.ID))
	}

	if e.Event != "" {
		buf.WriteString(fmt.Sprintf("event: %s\n", e.Event))
	}

	if e.Retry > 0 {
		buf.WriteString(fmt.Sprintf("retry: %d\n", e.Retry))
	}

	if e.Data != "" {
		// Data can be multiline, each line must be prefixed with "data: "
		lines := strings.Split(e.Data, "\n")
		for _, line := range lines {
			buf.WriteString(fmt.Sprintf("data: %s\n", line))
		}
	}

	// Empty line to indicate end of event
	buf.WriteString("\n")

	return buf.String()
}

// Writer wraps an io.Writer to write SSE events
type Writer struct {
	writer io.Writer
}

// NewWriter creates a new SSE writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

// WriteEvent writes an SSE event
func (w *Writer) WriteEvent(event *Event) error {
	_, err := w.writer.Write([]byte(event.Format()))
	return err
}

// WriteData writes a simple data-only event
func (w *Writer) WriteData(data string) error {
	event := &Event{Data: data}
	return w.WriteEvent(event)
}

// WriteComment writes an SSE comment (for keep-alive)
func (w *Writer) WriteComment(comment string) error {
	_, err := w.writer.Write([]byte(fmt.Sprintf(": %s\n\n", comment)))
	return err
}

// Flush flushes the writer if it implements http.Flusher
func (w *Writer) Flush() {
	if flusher, ok := w.writer.(interface{ Flush() }); ok {
		flusher.Flush()
	}
}

// Reader reads SSE events from an io.Reader
type Reader struct {
	scanner *bufio.Scanner
}

// NewReader creates a new SSE reader
func NewReader(r io.Reader) *Reader {
	return &Reader{
		scanner: bufio.NewScanner(r),
	}
}

// ReadEvent reads the next SSE event
func (r *Reader) ReadEvent() (*Event, error) {
	event := &Event{}

	for r.scanner.Scan() {
		line := r.scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if event.Data != "" || event.Event != "" {
				return event, nil
			}
			continue
		}

		// Comment line (ignore)
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Parse field: value
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}

		field := line[:colonIndex]
		value := strings.TrimSpace(line[colonIndex+1:])

		switch field {
		case "id":
			event.ID = value
		case "event":
			event.Event = value
		case "data":
			if event.Data != "" {
				event.Data += "\n"
			}
			event.Data += value
		case "retry":
			// Parse retry as integer
			var retry int
			fmt.Sscanf(value, "%d", &retry)
			event.Retry = retry
		}
	}

	if err := r.scanner.Err(); err != nil {
		return nil, err
	}

	// EOF
	if event.Data != "" || event.Event != "" {
		return event, nil
	}

	return nil, io.EOF
}
