package main

import (
	"testing"
)

// Test CRC
func TestCRC(t *testing.T) {
	data := []struct {
		input    string
		expected string
	}{
		{"GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,", "47"},
		{"GPRMC,000253.034,V,5038.664,N,00303.215,E,0.00,0.00,310522,,", "1C"},
	}

	for _, d := range data {
		if actual := CRC(d.input); actual != d.expected {
			t.Errorf("CRC(%q) = %q, expected %q", d.input, actual, d.expected)
		}
	}
}

// Test e7toDegMin
func TestE7toDegMin(t *testing.T) {
	data := []struct {
		input    IntString
		expected string
	}{
		{"485000000", "4830.000"}, // 48.5 = 48°30.000'
		{"11310000", "107.860"},   // 1.131 = 1°07.860'
		{"-10000000", "-100.000"}, // -1.0 = -1°00.000'
	}

	for _, d := range data {
		if actual := e7toDegMin(d.input); actual != d.expected {
			t.Errorf("e7toDegMin(%q) = %q, expected %q", d.input, actual, d.expected)
		}
	}
}

// Test latE7nmea
func TestLatE7nmea(t *testing.T) {
	data := []struct {
		input    IntString
		expected string
	}{
		{"485000000", "4830.000,N"},
		{"11310000", "0107.860,N"},
		{"-10000000", "0100.000,S"},
	}

	for _, d := range data {
		if actual := latE7nmea(d.input); actual != d.expected {
			t.Errorf("latE7nmea(%q) = %q, expected %q", d.input, actual, d.expected)
		}
	}
}

// Test lonE7nmea
func TestLonE7nmea(t *testing.T) {
	data := []struct {
		input    IntString
		expected string
	}{
		{"485000000", "04830.000,E"},   //    48.5 = 48°30.000' East
		{"11310000", "00107.860,E"},    //   1.131 = 1°07.860' East
		{"-10000000", "00100.000,W"},   //    -1.0 = 1°00.000' West
		{"-1401500000", "14009.000,W"}, // -140.15 = 140°09.000' West
	}

	for _, d := range data {
		if actual := lonE7nmea(d.input); actual != d.expected {
			t.Errorf("lonE7nmea(%q) = %q, expected %q", d.input, actual, d.expected)
		}
	}
}

// Test NMEA
func TestNMEA(t *testing.T) {
	data := []struct {
		input    Location
		expected string
	}{
		{
			Location{Timestamp: "2021-05-31T00:02:53Z", LatitudeE7: "485000000", LongitudeE7: "11310000", Accuracy: "14"},
			"$GPGGA,000253,4830.000,N,00107.860,E,1,04,14,0,M,,,,0000*23\n$GPRMC,000253,A,4830.000,N,00107.860,E,0.0,0.0,310521,,,A*77",
		},
	}

	for _, d := range data {
		if actual := NMEA(d.input); actual != d.expected {
			t.Errorf("NMEA(%q) = %q, expected %q", d.input, actual, d.expected)
		}
	}
}
