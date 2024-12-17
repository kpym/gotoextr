# Google Location History File Format

## Records.json (before 2024)

The file `Records.json` looks like this:

```json
{
  "locations": [
    {
      "latitudeE7": 506553765,
      "longitudeE7": 30632229,
      "accuracy": 24,
      "timestamp": "2012-01-27T21:14:42.352Z"
      ...
    },
    ...
  ]
}
```

- `latitudeE7` and `longitudeE7` represent coordinates multiplied by `1e7`.  
- `accuracy` is the location accuracy in meters.  
- `timestamp` is in **ISO 8601** format, which represents UTC date and time.

---

## Phone Location History (after 2024)

The file `<name>.json` exported from the phone looks like this:

```json
{
  ...
  "rawSignals": [
    ...
    {
      "position": {
        "LatLng": "50.6443831°, 3.0536723°",
        "accuracyMeters": 13,
        "timestamp": "2024-12-07T17:46:25.000+01:00",
        ...
      },
      ...
    },
    ...
  ]
}
```

- `LatLng` represents coordinates in **decimal degrees**.  
- `accuracyMeters` is the location accuracy in meters.  
- `timestamp` is in **RFC 3339** format, representing the local time with a timezone offset (`+01:00` indicates 1 hour ahead of UTC).

---

## Key Differences

1. **Coordinate Scaling**:  
   - `Records.json`: Coordinates are scaled (`latitudeE7`, `longitudeE7`).  
   - New format: Coordinates are in plain decimal degrees (`LatLng`).

2. **Timestamp Format**:  
   - `Records.json`: ISO 8601 format, always UTC.  
   - New format: RFC 3339 format, includes local time and timezone offset.

3. **File Naming**:  
   - Before 2024: `Records.json` (contained in a `zip` that was exported from [Google Takout](https://takeout.google.com/settings/takeout/custom/location_history)).  
   - After 2024: `<name>.json` (exported directly from the phone).
