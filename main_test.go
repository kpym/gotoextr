package main

import (
	"testing"
)

func TestNextDay(t *testing.T) {
	data := []struct {
		in  string
		out string
	}{
		{"2015-01-01", "2015-01-02"},
		{"2015-01-31", "2015-02-01"},
		{"2015-12-31", "2016-01-01"},
	}

	for _, d := range data {
		if nextDay(d.in) != d.out {
			t.Errorf("nextDay(%s) != %s", d.in, d.out)
		}
	}
}

func TestE7toDec(t *testing.T) {
	data := []struct {
		in  IntString
		out string
	}{
		{"0987654321", "098.7654321"},
		{"987654321", "98.7654321"},
		{"87654321", "8.7654321"},
		{"7654321", ".7654321"},
	}

	for _, d := range data {
		if e7toDec(d.in) != d.out {
			t.Errorf("e7toDec(%s) != %s", d.in, d.out)
		}
	}
}

func TestAcceptAccuracy(t *testing.T) {
	data := []struct {
		a   IntString
		max string
		out bool
	}{
		{"5", "5", true},
		{"5", "6", true},
		{"5", "4", false},
		{"15", "5", false},
		{"15", "10", false},
		{"15", "15", true},
		{"15", "20", true},
		{"15", "", false},
	}

	for _, d := range data {
		if acceptAccuracy(d.a, d.max) != d.out {
			t.Errorf("acceptAccuracy(%s, %s) != %t", d.a, d.max, d.out)
		}
	}
}

func TestSameDigits(t *testing.T) {
	data := []struct {
		a      IntString
		b      IntString
		expect int
	}{
		{"987654321", "987654321", 7},
		{"987654321", "987654021", 4},
		{"987654321", "987650321", 3},
		{"987654321", "980654321", 0},
		{"987654321", "907650321", -1},
		{"987654321", "87650321", -2},
	}

	for _, d := range data {
		if sameDigits(d.a, d.b) != d.expect {
			t.Errorf("samePrecision(%s, %s) = %d != %d", d.a, d.b, sameDigits(d.a, d.b), d.expect)
		}
	}
}
