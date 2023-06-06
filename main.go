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
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/goccy/go-json"
	"github.com/gosuri/uilive"
)

// The version that is set by goreleaser
var version = "dev"

// The usage string
var usage = "hist2gpx [version: " + version + "]" + ` extract history data from Google Location History.
	
Usage:
  hist2gpx [-h] -s <start> [options] <input>
  
Options:
  -h --help              Show this screen.
  -s <start>             Start date in YYYY-MM-DD format
  -e <end>               End date in YYYY-MM-DD format [default: <start>]
  -a <accuracy>          Keeps only locations with accuracy less than <accuracy> meters [default: 40]
	-tp <tp>               New track if coordinates have less than <tp> digits in common [default: 1]
	-sp <sp>               New segment if coordinates have less than <sp> digits in common [default: 2]
  -o <output>            Output file name [default: history_<start>_<end>.gpx]
  <input>                Input file name (zip or json)

Examples:
  hist2gpx -s 2012-01-01 -e 2012-01-31 -a 40 takeout.zip
`

type IntString string

type Location struct {
	LatitudeE7  IntString `json:"latitudeE7"`
	LongitudeE7 IntString `json:"longitudeE7"`
	Accuracy    IntString `json:"accuracy"`
	Timestamp   string    `json:"timestamp"`
}

// Json unmashalling for IntString
func (i *IntString) UnmarshalJSON(data []byte) error {
	*i = IntString(data)
	return nil
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
      <trkpt lat="%s" lon="%s">
        <time>%s</time>
        <accuracy>%s</accuracy>
      </trkpt>`
	newTrack = `
    </trkseg>
  </trk>
  <trk>
    <trkseg>`
	newSegment = `
    </trkseg>
    <trkseg>`
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

func e7toDec(e7 IntString) string {
	return string(e7[:len(e7)-7] + "." + e7[len(e7)-7:])
}

func acceptAccuracy(a IntString, max string) bool {
	if len(a) != len(max) {
		return len(a) <= len(max)
	}
	return string(a) <= max
}

func sameDigits(a, b IntString) int {
	// convert with atoi
	na, _ := strconv.Atoi(string(a))
	nb, _ := strconv.Atoi(string(b))
	if na == nb {
		return 7
	}
	diff := na - nb
	if diff < 0 {
		diff = -diff
	}
	strDiff := strconv.Itoa(diff)
	return 7 - len(strDiff)

	// if len(a) != len(b) {
	// 	return 0
	// }
	// for i := 0; i < len(a) && i < len(b); i++ {
	// 	if a[i] != b[i] {
	// 		return 7 + i - len(a) // 0 corresponds the first digit after the decimal point
	// 	}
	// }
	// return 7
}

func main() {
	// strtup time
	now := time.Now()
	// new terminal writer
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	print := func(r, w, s, t int, timestamp string, sec float64) {
		fmt.Fprintf(writer, "Read %d positions in %.2f seconds", r, sec)
		if timestamp != "" {
			fmt.Fprintf(writer, " until %s", timestamp)
		}
		fmt.Fprintln(writer)
		fmt.Fprintf(writer.Newline(), "Wrote %d positions in %d segments in %d tracks\n", w, s, t)
	}

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
	accuracy, err := arguments.String("-a")
	check(err)
	tp, err := arguments.Int("-tp")
	check(err)
	sp, err := arguments.Int("-sp")
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
		// try to read all the file in memory
		var zf *zip.Reader
		content, err := os.ReadFile(inputname)
		if err == nil {
			// associate a zip reader to the content
			zf, err = zip.NewReader(bytes.NewReader(content), int64(len(content)))
			check(err)
			fmt.Println("Using zip file in memory")
		} else {
			// Open the zip file
			zf, err := zip.OpenReader(inputname)
			check(err)
			defer zf.Close()
		}
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

	r, w, t, s := 0, 0, 0, 0
	var lastLat, lastLon IntString
	for l := range locations {
		r++
		if l.Timestamp >= start && l.Timestamp < endNext && acceptAccuracy(l.Accuracy, accuracy) {
			if w > 0 {
				if sameDigits(lastLat, l.LatitudeE7) < tp || sameDigits(lastLon, l.LongitudeE7) < tp {
					t++
					s++
					outfile.WriteString(newTrack)
				} else if sameDigits(lastLat, l.LatitudeE7) < sp || sameDigits(lastLon, l.LongitudeE7) < sp {
					s++
					outfile.WriteString(newSegment)
				}
			}
			w++
			lastLat, lastLon = l.LatitudeE7, l.LongitudeE7
			fmt.Fprintf(outfile, locFormat, e7toDec(l.LatitudeE7), e7toDec(l.LongitudeE7), l.Timestamp, l.Accuracy)
		}
		// display the progress every 0x8000=32768 records
		if r&0x7fff == 0 {
			print(r, w, s, t, l.Timestamp, time.Since(now).Seconds())
		}
	}
	// Write the footer
	outfile.WriteString(footer)

	// The end
	print(r, w, s, t, "", time.Since(now).Seconds())
}
