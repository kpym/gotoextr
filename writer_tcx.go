package main

import (
	"bufio"
	"io"
	"text/template"
)

const (
	tcxHeader = `<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<TrainingCenterDatabase xmlns="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2 http://www.garmin.com/xmlschemas/TrainingCenterDatabasev2.xsd">
	<Courses>
		<Course>
			<Name>New Course</Name>
			<Track>`
	tcxLocTemplate = `
				<Trackpoint>
					<Time>{{ .Timestamp }}</Time>
					<Position>
						<LatitudeDegrees>{{ .LatitudeE7 | e7todec }}</LatitudeDegrees>
						<LongitudeDegrees>{{ .LongitudeE7 | e7todec }}</LongitudeDegrees>
					</Position>
				</Trackpoint>`
	tcxNewTrack = `
			</Track>
		</Course>
		<Course>
			<Track>`
	tcxNewSegment = `
			</Track>
			<Track>`
	tcxFooter = `
			</Track>
		</Course>
	</Courses>
</TrainingCenterDatabase>
`
)

func NewTCXWriter(w io.Writer) Writer {
	// compile the templates
	locTemplate := template.New("tcx").Funcs(funcMap)
	locTemplate = template.Must(locTemplate.Parse(tcxLocTemplate))

	return &TemplateWriter{
		w:          bufio.NewWriter(w),
		header:     tcxHeader,
		location:   locTemplate,
		newSegment: tcxNewSegment,
		newTrack:   tcxNewTrack,
		footer:     tcxFooter,
	}
}
