package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"
)

const (
	nmeaLocTemplate = "{{ . | NMEA }}\n"
)

func init() {
	funcMap["NMEA"] = NMEA
}

func e7toDegMin(e7 IntString) string {
	deg := string(e7)[:len(e7)-7]
	min, _ := strconv.Atoi(string(e7)[len(e7)-7:])
	mins := fmt.Sprintf("%06.3f", float32(min*60)/1e7)
	return deg + mins
}

func latE7nmea(e7 IntString) string {
	l := e7toDegMin(e7)
	if l[0] == '-' {
		// padding with 0s to 8 chars
		return fmt.Sprintf("%08s,S", l[1:])
	}
	return fmt.Sprintf("%08s,N", l)
}

func lonE7nmea(e7 IntString) string {
	l := e7toDegMin(e7)
	if l[0] == '-' {
		// padding with 0s to 9 chars
		return fmt.Sprintf("%09s,W", l[1:])
	}
	return fmt.Sprintf("%09s,E", l)
}

var timeReplacer = strings.NewReplacer(":", "", "Z", "")

// CRC calculates the CRC checksum for a NMEA sentence
// by XORing each byte between the $ and * characters
func CRC(s string) string {
	var crc uint8 = 0
	for i := 0; i < len(s); i++ {
		crc ^= uint8(s[i])
	}
	return fmt.Sprintf("%02X", crc)
}

// AccuracyToHDOP converts accuracy to HDOP
// this conversion is only approximate
// there is no exact conversion from accuracy to HDOP
// so I use the formula: HDOP = accuracy / 4
func AccuracyToHDOP(accuracy IntString) string {
	acc, _ := strconv.ParseFloat(string(accuracy), 32)
	return fmt.Sprintf("%.1f", acc/4)
}

// NMEA convert location to GPGGA and GPRMC NMEA sentences
func NMEA(l Location) string {
	// convert timestamp to NMEA format
	td := strings.Split(l.Timestamp, "T")
	// convert date to NMEA format
	ds := strings.Split(td[0], "-")
	d := ds[2] + ds[1] + ds[0][2:]
	// convert time to NMEA format
	t := timeReplacer.Replace(td[1])
	// latitude and longitude in NMEA format
	lat, lon := latE7nmea(l.LatitudeE7), lonE7nmea(l.LongitudeE7)
	// convert latitude to NMEA format
	gpgga := fmt.Sprintf("GPGGA,%s,%s,%s,1,04,%s,0,M,,,,0000", t, lat, lon, AccuracyToHDOP(l.Accuracy))
	gprmc := fmt.Sprintf("GPRMC,%s,A,%s,%s,0.0,0.0,%s,,,A", t, lat, lon, d)
	return fmt.Sprintf("$%s*%s\n$%s*%s", gpgga, CRC(gpgga), gprmc, CRC(gprmc))
}

func NewNMEAWriter(w io.Writer) Writer {
	// compile the templates
	locTemplate := template.New("nmea").Funcs(funcMap)
	locTemplate = template.Must(locTemplate.Parse(nmeaLocTemplate))

	return &TemplateWriter{
		w:        bufio.NewWriter(w),
		location: locTemplate,
	}
}
