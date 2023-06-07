package main

import (
	"bufio"
	"text/template"
)

// Writer is an interface for writing GPX, CSV, GeoJSON, etc.
type Writer interface {
	// WriteHeader writes the header
	WriteHeader() error

	// WriteLocation writes a location
	WriteLocation(l Location) error

	// WriteNewSegment writes a new segment
	WriteNewSegment() error

	// WriteNewTrack writes a new track
	WriteNewTrack() error

	// WriteFooter writes the footer
	WriteFooter() error

	// Flush flushes the writer
	Flush() error
}

// e7toDec converts a latitude or longitude from e7 format to decimal
func e7toDec(e7 IntString) string {
	return string(e7[:len(e7)-7] + "." + e7[len(e7)-7:])
}

// funcMap is the map of functions used in the templates
var funcMap template.FuncMap = map[string]interface{}{}

func init() {
	funcMap["e7todec"] = e7toDec
}

type TemplateWriter struct {
	w          *bufio.Writer
	header     string
	location   *template.Template
	newSegment string
	newTrack   string
	footer     string
}

func (t *TemplateWriter) WriteHeader() error {
	_, err := t.w.WriteString(t.header)
	return err
}

func (t *TemplateWriter) WriteLocation(l Location) error {
	return t.location.Execute(t.w, l)
}

func (t *TemplateWriter) WriteNewSegment() error {
	_, err := t.w.WriteString(t.newSegment)
	return err
}

func (t *TemplateWriter) WriteNewTrack() error {
	_, err := t.w.WriteString(t.newTrack)
	return err
}

func (t *TemplateWriter) WriteFooter() error {
	_, err := t.w.WriteString(t.footer)
	return err
}

func (t *TemplateWriter) Flush() error {
	return t.w.Flush()
}
