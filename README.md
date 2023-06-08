# hist2geo
Small and durty program to extract GPX data from Google Location History.

## Usage

1. Download your location history from [Google Takeout](https://takeout.google.com/settings/takeout) as `zip` archive.
2. Run `hist2geo` with the downloaded archive as argument. For example to extract data for January 1, 2023 run:
```bash
hist2geo -s 2023-01-01 takeout-20230501T000000Z-001.zip
``` 
The output will be written to the file `history_2023-01-01.gpx`.

You can also manually extract `Records.json` and use it as parameter. Once extracted this file is quite big (several hundred MB).

### Help message

```
$ hist2geo.exe -h
hist2geo [version: x.y.z] extract history data from Google Location History.

Usage:
  hist2geo [-h] -s <start> [options] <input>

Options:
  -h --help              Show this screen.
  -s <start>             Start date in YYYY-MM-DD format
  -e <end>               End date in YYYY-MM-DD format [default: <start>]
  -a <accuracy>          Keeps only locations with accuracy less than <accuracy> meters [default: 40]
  -t <tp>                New track if coordinates have less than <tp> digits in common [default: 1]  
  -g <sp>                New segment if coordinates have less than <sp> digits in common [default: 2]
  -f <format>            Output format (gpx|kml|tcx|csv|nmea) [default: gpx]
  -o <output>            Output file name [default: history_<start>_<end>.<format>]
  <input>                Input file name (zip or json)

Examples:
  hist2geo -s 2012-01-01 -e 2012-01-31 -a 40 takeout.zip
```

## Installation

You can download the latest binary from the [releases](github.com/kpym/hist2geo/releases) page.

Or you can install it from source:
```bash
go install github.com/kpym/hist2geo@latest
```

## Why? 

I use my travel history to geotag my photos. I often use a tracking application to record my positions, but occasionally (often?) I forget to launch it. In this case, Google Takout (Location history) helps me by extracting my tracks in GPX format.

## Inspiration

This software is strongly inspired by [location-history-json-converter](https://github.com/Scarygami/location-history-json-converter) 🙏. But since the pyton application is rather slow, I decided to make one in go that is about  10x faster. The original application is also more complete, I only implemented the features I needed.

## License

[MIT](LICENSE)

