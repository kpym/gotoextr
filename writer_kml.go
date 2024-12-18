package main

import (
	"bufio"
	"io"
	"text/template"
)

const (
	kmlHeader = `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
	<Document>
		<name>Location History</name>`
	kmlLocTemplate = `
		<Placemark>
			<TimeStamp><when>{{ .Timestamp }}</when></TimeStamp>
			<ExtendedData>
				<Data name="accuracy"><value>{{ .Accuracy }}</value></Data>
			</ExtendedData>
			<Point><coordinates>{{ .LongitudeE7 | e7todec }},{{ .LatitudeE7 | e7todec }}</coordinates></Point>
		</Placemark>`
	kmlFooter = `
	</Document>
</kml>
`
)

func NewKMLWriter(w io.Writer) Writer {
	// compile the templates
	locTemplate := template.New("kml").Funcs(funcMap)
	locTemplate = template.Must(locTemplate.Parse(kmlLocTemplate))

	return &TemplateWriter{
		w:        bufio.NewWriter(w),
		header:   kmlHeader,
		location: locTemplate,
		footer:   kmlFooter,
	}
}
