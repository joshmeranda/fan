package cmd

import (
	"fmt"
	"strings"
)

const (
	TABLE_COLUMN_WIDTH = 20
)

type header struct {
	name  string
	width int
}

type table struct {
	format  string
	headers []header
	rows    [][]any
}

func NewTable(headers []header) *table {
	formats := make([]string, len(headers))
	for i, h := range headers {
		formats[i] = fmt.Sprintf("%%%ds", h.width)
	}

	return &table{
		headers: headers,
		rows:    make([][]any, 0),
		format:  strings.Join(formats, " ") + "\n",
	}
}

func (t *table) addRow(row []string) error {
	if len(row) != len(t.headers) {
		return fmt.Errorf("length of row must be == length of headers")
	}

	data := make([]any, len(row))
	for i, field := range row {
		colWidth := t.headers[i].width
		if len(field) > colWidth {
			field = "..." + field[len(field)-colWidth+3:]
		}

		data[i] = field
	}

	t.rows = append(t.rows, data)

	return nil
}

func (t *table) display() {
	s := make([]any, 0, len(t.headers))
	for _, h := range t.headers {
		s = append(s, h.name)
	}

	fmt.Printf(t.format, s...)

	for _, r := range t.rows {
		fmt.Printf(t.format, r...)
	}
}
