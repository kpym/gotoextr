package main

import (
	"bufio"
	"io"
	"text/template"
)

const (
	gpxHeader = `<?xml version="1.0" encoding="UTF-8"?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" version="1.1" creator="Google Latitude JSON Converter" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd">
	<metadata>
		<name>Location History</name>
	</metadata>
	<trk>
		<trkseg>`
	gpxLocTemplate = `
			<trkpt lat="{{ .LatitudeE7 | e7todec }}" lon="{{ .LongitudeE7 | e7todec }}">
				<time>{{ .Timestamp }}</time>
				<accuracy>{{ .Accuracy }}</accuracy>
			</trkpt>`
	gpxNewTrack = `
		</trkseg>
	</trk>
	<trk>
		<trkseg>`
	gpxNewSegment = `
		</trkseg>
		<trkseg>`
	gpxFooter = `
		</trkseg>
	</trk>
</gpx>
`
)

func NewGPXWriter(w io.Writer) Writer {
	// compile the templates
	locTemplate := template.New("gpx").Funcs(funcMap)
	locTemplate = template.Must(locTemplate.Parse(gpxLocTemplate))

	return &TemplateWriter{
		w:          bufio.NewWriter(w),
		header:     gpxHeader,
		location:   locTemplate,
		newSegment: gpxNewSegment,
		newTrack:   gpxNewTrack,
		footer:     gpxFooter,
	}
}
