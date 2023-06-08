# hist2geo
Small and durty program to extract GPX data from Google Location History.

## Usage
1. Download your location history from [Google Takeout](https://takeout.google.com/settings/takeout) as `zip` archive.
2. Run `hist2geo` with the downloaded archive as argument. For example to extract data for January 1, 2023 run:
```bash
hist2geo -s 2023-01-01 takeout-20230501T000000Z-001.zip
``` 
The output will be written to `history_2023-01-01.gpx` file.

## Installation

You can download the latest binary from [releases](github.com/kpym/hist2geo/releases) page.

Or you can install it from source:
```bash
go install github.com/kpym/hist2geo@latest
```

## Why? 

I use my travel history to geotag my photos. I often use a tracking application to record my positions, but occasionally (often?) I forget to launch it. In this case, Google Takout (Location history) helps me by extracting my tracks in GPX format.

## Inspiration. 

This software is strongly inspired by [location-history-json-converter](https://github.com/Scarygami/location-history-json-converter). But as this pyton application is rather slow, I decided to make one in go, which is 10x faster. The original application is also more complete, I only implemented the features I needed.

## License

[MIT](LICENSE)

