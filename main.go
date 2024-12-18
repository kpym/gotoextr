// Before 2024
// Read Records.json and extract history data.
// The input date is json with the following format
// {
// "locations": [
//
//	{
//	  "latitudeE7": 506553765,
//	  "longitudeE7": 30632229,
//	  "accuracy": 24,
//	  "timestamp": "2012-01-27T21:14:42.352Z"
//	  ...
//	},
//	  ...
//
// ]
//
// The file is very large, so we read it using json.Decoder.
// If the file is .zip (smaller) we try to read it in memory.
//
// Since 2024.
// We can export the location history from an Android device to a JSON file.
// The JSON file has the following format:
//
//	{
//	  ...
//	  "rawSignals": [
//	    {
//	      ...
//	      "position": {
//	        "LatLng": "50.6443831°, 3.0536723°",
//	        "accuracyMeters": 13,
//	        "altitudeMeters": 65.30000305175781,
//	        "source": "UNKNOWN",
//	        "timestamp": "2024-12-07T17:46:25.000+01:00",
//	        "speedMetersPerSecond": 0.0
//	      }
//	      ,
//	      ...
//	    }
//	  ]
//	}
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/goccy/go-json"
	"github.com/gosuri/uilive"
)

// The version that is set by goreleaser
var version = "dev"

// The usage string
var usage = "gotoextr [version: " + version + "]" + ` extract history data from Google Location History.

Usage:
  gotoextr [-h] -s <start> [options] <input>
  
Options:
  -h --help        Show this screen.
  -s <start>       Start date in YYYY-MM-DD format
  -e <end>         End date in YYYY-MM-DD format [default: <start>]
  -a <accuracy>    Keeps only locations with accuracy less than <accuracy> meters [default: 40]
  -t <tp>          New track if coordinates have less than <tp> digits in common [default: 1]
  -g <sp>          New segment if coordinates have less than <sp> digits in common [default: 2]
  -f <format>      Output format (gpx|kml|tcx|csv|nmea) [default: gpx]
  -o <output>      Output file name [default: history_<start>_<end>.<format>]
  <input>          Input file name (zip or json)

Examples:
  gotoextr -s 2012-01-01 -e 2012-01-31 -a 40 takeout.zip
`

// IntString is a string that can be unmarshalled from an int
// This avoid number parsing
type IntString string

// Location is the struct for the location data
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

type latlng struct {
	Latitude  IntString
	Longitude IntString
}

type posObject struct {
	Position position `json:"position"`
}

type position struct {
	LatLng    latlng    `json:"LatLng"`
	Accuracy  IntString `json:"accuracyMeters"`
	Timestamp string    `json:"timestamp"`
}

// coordToIntString converts a string "XX.XXXXXXX°" to an IntString of E7 format
func coordToIntString(s string) (IntString, error) {
	// remove the °
	s = strings.TrimSuffix(s, "°")
	// split the string in two parts
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid coord %s", s)
	}
	// normalize the integer part
	if len(parts[0]) == 0 {
		parts[0] = "0"
	}
	// cut/pad the decimal part to 7 digits
	if len(parts[1]) == 7 {
		// nothing to do
	} else if len(parts[1]) > 7 {
		parts[1] = parts[1][:7]
	} else {
		parts[1] += strings.Repeat("0", 7-len(parts[1]))
	}

	// return the IntString
	return IntString(parts[0] + parts[1]), nil
}

// Json unmashalling for latlng from "XX.XXXXXXX°,YY.YYYYYYY°"
func (ll *latlng) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid LatLng %s", s)
	}
	ll.Latitude, err = coordToIntString(parts[0])
	if err != nil {
		return err
	}
	ll.Longitude, err = coordToIntString(parts[1])
	if err != nil {
		return err
	}
	return nil
}

// toUTC convert RFC3339 to a string without the timezone ending with "Z"
// example "2024-12-07T17:46:25.000+01:00" -> "2024-12-07T16:46:25.000Z"
func toUTC(s string) string {
	// if string do not contain a timezone (+ or -), return it
	if !strings.ContainsAny(s, "+-") {
		return s
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.UTC().Format(time.RFC3339)
}

// toLocation converts a position to a Location
func (p *position) toLocation() Location {
	return Location{
		LatitudeE7:  p.LatLng.Latitude,
		LongitudeE7: p.LatLng.Longitude,
		Accuracy:    p.Accuracy,
		Timestamp:   toUTC(p.Timestamp),
	}
}

// check is a helper function to check for errors
func check(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func getOldLocation(decoder *json.Decoder) (loc Location, err error) {
	err = decoder.Decode(&loc)
	return loc, err
}

func getNewLocation(decoder *json.Decoder) (Location, error) {
	var pos posObject
	err := decoder.Decode(&pos)
	if err != nil {
		return Location{}, err
	}
	if pos.Position.Timestamp == "" {
		return Location{}, fmt.Errorf("missing timestamp")
	}
	return pos.Position.toLocation(), nil
}

// locBufSize is the size of the channel buffer for locations
const locBufSize = 100

// Read the input file and return a channel of locations
func Read(reader io.Reader) chan Location {
	// Create a decoder
	decoder := json.NewDecoder(reader)

	var version int
	// Read up to the "locations" key
	for {
		t, err := decoder.Token()
		check(err)
		if t == "locations" {
			// found the locations key, so old format
			version = 1
			break
		}
		if t == "rawSignals" {
			// found the rawSignals key, so new format
			version = 2
			break
		}
	}

	// old or new format
	var getLocation func(*json.Decoder) (Location, error)
	switch version {
	case 1:
		getLocation = getOldLocation
	case 2:
		getLocation = getNewLocation
	default:
		check(fmt.Errorf("unknown json version"))
	}

	// Read the array start
	t, err := decoder.Token()
	check(err)
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		check(fmt.Errorf("expected array start, got %T: %v", t, t))
	}

	// Create a channel to send the locations
	locations := make(chan Location, locBufSize)

	// Start a goroutine to read the array
	go func() {
		// start reading the array
		for decoder.More() {
			loc, err := getLocation(decoder)
			// skip invalid locations
			if err == nil {
				locations <- loc
			}
		}
		// Close the channel when done
		close(locations)
	}()

	return locations
}

// header, locFormat, newTrack, newSegment, footer are the parts of the gpx file

// nextDay returns the next day in YYYY-MM-DD format
func nextDay(date string) string {
	t, err := time.Parse("2006-01-02", date)
	check(err)
	return t.AddDate(0, 0, 1).Format("2006-01-02")
}

// acceptAccuracy returns true if the accuracy is less than max
// it works with strings to avoid number parsing
func acceptAccuracy(a IntString, max string) bool {
	if len(a) != len(max) {
		return len(a) <= len(max)
	}
	return string(a) <= max
}

// sameDigits returns the number of digits in common between two coordinates
// for example 987654321 which represent 98.7654321 and 987650321 which represent 98.7650321
// have 3 decimal digits (after the point) in common
func sameDigits(a, b IntString) int {
	na := len(a)
	nb := len(b)
	if na != nb || na < 7 {
		// the coordinates are not the same length or invalid IntString
		return 0
	}
	if a[:na-7] != b[:nb-7] {
		// the integer part is different
		return 0
	}
	n := 0
	for i := na - 7; i < na; i++ {
		if a[i] != b[i] {
			break
		}
		n++
	}
	return n
}

func main() {
	// strtup time
	now := time.Now()
	// new terminal writer
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	// the info print function
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
	tp, err := arguments.Int("-t")
	check(err)
	sp, err := arguments.Int("-g")
	check(err)
	inputname, err := arguments.String("<input>")
	check(err)
	outputname, err := arguments.String("-o")
	check(err)
	format, err := arguments.String("-f")
	check(err)
	format = strings.ToLower(format)
	// if format is not one of the allowed, exit
	switch format {
	case "gpx", "kml", "tcx", "csv", "nmea": // ok
	default:
		check(fmt.Errorf("unknown format %s", format))
	}
	if outputname == "history_<start>_<end>.<format>" {
		if start == end {
			outputname = fmt.Sprintf("history_%s.%s", start, format)
		} else {
			outputname = fmt.Sprintf("history_%s_%s.%s", start, end, format)
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
		} else {
			// Cant read entire file in memory, so open the zip file from disk
			zfc, err := zip.OpenReader(inputname)
			check(err)
			defer zfc.Close()
			zf = &zfc.Reader
		}
		// Find the file inside the zip
		for _, f := range zf.File {
			if strings.HasSuffix(f.Name, "/Records.json") {
				input, err = f.Open()
				check(err)
				break
			}
		}
		if input == nil {
			check(fmt.Errorf("file 'Records.json' not found in '%s'", inputname))
		}
	} else {
		// if the file is not a zip, it should be the Records.json file
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

	// Create a new writer
	var output Writer
	switch format {
	case "gpx":
		output = NewGPXWriter(outfile)
	case "kml":
		output = NewKMLWriter(outfile)
	case "tcx":
		output = NewTCXWriter(outfile)
	case "csv":
		output = NewCSVWriter(outfile)
	case "nmea":
		output = NewNMEAWriter(outfile)
	default:
		// this should never happen
		panic(fmt.Errorf("unknown format %s, this should be verified before", format))
	}
	defer output.Flush()

	// Write the header
	output.WriteHeader()

	// Some counters :
	// r : number of positions read
	// w : number of positions written
	// t : number of tracks written
	// s : number of segments written
	r, w, t, s := 0, 0, 1, 1
	// the last position used to detect new segments and tracks
	var lastLat, lastLon IntString
	// loop over the locations
	for l := range locations {
		r++
		// check if the location is in the time range and has the required accuracy
		if l.Timestamp >= start && l.Timestamp < endNext && acceptAccuracy(l.Accuracy, accuracy) {
			if w > 0 {
				// if it is not the first location, check if the distance from the previous one
				// requires a new segment or a new track
				if sameDigits(lastLat, l.LatitudeE7) < tp || sameDigits(lastLon, l.LongitudeE7) < tp {
					t++
					s++
					output.WriteNewTrack()
				} else if sameDigits(lastLat, l.LatitudeE7) < sp || sameDigits(lastLon, l.LongitudeE7) < sp {
					s++
					output.WriteNewSegment()
				}
			}
			w++
			lastLat, lastLon = l.LatitudeE7, l.LongitudeE7
			// write the location in the output file
			output.WriteLocation(l)
		}
		// display the progress every 0x8000=32768 records
		if r&0x7fff == 0 {
			print(r, w, s, t, l.Timestamp, time.Since(now).Seconds())
		}
	}
	// Write the footer
	output.WriteFooter()

	// The end
	print(r, w, s, t, "", time.Since(now).Seconds())
}
