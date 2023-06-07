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

## License

[MIT](LICENSE)

