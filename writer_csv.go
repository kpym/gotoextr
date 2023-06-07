package main

import (
	"bufio"
	"io"
	"text/template"
)

const (
	csvHeader      = "timestamp,lat,lon,accuracy\n"
	csvLocTemplate = "{{ .Timestamp }},{{ .LatitudeE7 | e7todec }},{{ .LongitudeE7 | e7todec }},{{ .Accuracy }}\n"
)

func NewCSVWriter(w io.Writer) Writer {
	// compile the templates
	locTemplate := template.New("csv").Funcs(funcMap)
	locTemplate = template.Must(locTemplate.Parse(csvLocTemplate))

	return &TemplateWriter{
		w:        bufio.NewWriter(w),
		header:   csvHeader,
		location: locTemplate,
	}
}
