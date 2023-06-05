// Read Records.json and extract history data
// The input date is json with the following format
// {
// "locations": [
//   {
//     "latitudeE7": 506553765,
//     "longitudeE7": 30632229,
//     "accuracy": 24,
//     "timestamp": "2012-01-27T21:14:42.352Z"
//     ...
//   },
//     ...
// ]
// The file is very large, so we read it using json.Decoder

package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
)

var (
	// The version that is set by goreleaser
	version = "dev"
	// The usage string
	usage = "hist2gpx [version: " + version + "]" + ` extract history data from Google Location History.
	
  Usage:
  extract_history.go [-h] -s <start> [-e <end>] [-a <accuracy>] [-o <output>] <input>
  
  Options:
  -h --help       Show this screen.
  -s <start>      Start date in YYYY-MM-DD format
  -e <end>        End date in YYYY-MM-DD format [default: <start>]
  -a <accuracy>   Maximum accuracy in meters [default: 40]
  -i <input>      Input file name [default: Records.json]
  -o <output>     Output file name [default: history_<start>_<end>.gpx]
  `
)

type Location struct {
	LatitudeE7  int    `json:"latitudeE7"`
	LongitudeE7 int    `json:"longitudeE7"`
	Accuracy    int    `json:"accuracy"`
	Timestamp   string `json:"timestamp"`
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Read(reader io.Reader) chan Location {
	// Create a decoder
	decoder := json.NewDecoder(reader)

	// Read up to the "locations" key
	for {
		t, err := decoder.Token()
		check(err)
		if t == "locations" {
			break
		}
	}

	// Read the array start
	t, err := decoder.Token()
	check(err)
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		check(fmt.Errorf("expected array start, got %T: %v", t, t))
	}

	// Create a channel to send the locations
	locations := make(chan Location, 1000)

	go func() {
		// start reading the array
		for decoder.More() {
			var loc Location
			err := decoder.Decode(&loc)
			check(err)
			// fmt.Printf("%+v\n", loc)
			locations <- loc
		}
		// Close the channel when done
		close(locations)
	}()

	return locations
}

const (
	header = `<?xml version="1.0" encoding="UTF-8"?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" version="1.1" creator="Google Latitude JSON Converter" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd">
  <metadata>
    <name>Location History</name>
  </metadata>
  <trk>
    <trkseg>`
	locFormat = `
      <trkpt lat="%.7f" lon="%.7f">
        <time>%s</time>
        <accuracy>%d</accuracy>
      </trkpt>`
	footer = `
    </trkseg>
  </trk>
</gpx>
`
)

func nextDay(date string) string {
	t, err := time.Parse("2006-01-02", date)
	check(err)
	return t.AddDate(0, 0, 1).Format("2006-01-02")
}

func main() {
	// strtup time
	now := time.Now()

	// Parse the command line
	arguments, err := docopt.ParseDoc(usage)
	check(err)
	// get the arguments
	start, err := arguments.String("-s")
	check(err)
	if arguments["-e"] == "<start>" {
		arguments["-e"] = arguments["-s"]
	}
	end, err := arguments.String("-e")
	check(err)
	endNext := nextDay(end)
	accuracy, err := arguments.Int("-a")
	check(err)
	inputname, err := arguments.String("<input>")
	check(err)
	outputname, err := arguments.String("-o")
	check(err)
	if outputname == "history_<start>_<end>.gpx" {
		if start == end {
			outputname = fmt.Sprintf("history_%s.gpx", start)
		} else {
			outputname = fmt.Sprintf("history_%s_%s.gpx", start, end)
		}
	}

	// The input reader
	var input io.Reader
	// If the file is .zip, access the file inside the zip
	if strings.HasSuffix(inputname, ".zip") {
		// Open the zip file
		zf, err := zip.OpenReader(inputname)
		check(err)
		defer zf.Close()
		// Find the file inside the zip
		for _, f := range zf.File {
			if strings.HasSuffix(f.Name, "/Records.json") {
				input, err = f.Open()
				check(err)
				break
			}
		}
	} else {
		file, err := os.Open(inputname)
		check(err)
		defer file.Close()
		input = file
	}

	// Create a new reader.
	reader := bufio.NewReader(input)

	// Read the locations
	locations := Read(reader)

	// Open the output file
	outfile, err := os.Create(outputname)
	check(err)
	defer outfile.Close()

	// Write the header
	outfile.WriteString(header)

	r, w := 0, 0
	for l := range locations {
		r++
		if l.Timestamp >= start && l.Timestamp < endNext && l.Accuracy <= accuracy {
			w++
			fmt.Fprintf(outfile, locFormat, float64(l.LatitudeE7)/1e7, float64(l.LongitudeE7)/1e7, l.Timestamp, l.Accuracy)
		}
	}

	// Write the footer
	outfile.WriteString(footer)
	duration := time.Since(now)
	fmt.Printf("Read %d records, wrote %d records in %s seconds\n", r, w, duration)
}
