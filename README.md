# leap

This is a public JSON API for leap seconds announced by the [IERS](https://en.wikipedia.org/wiki/International_Earth_Rotation_and_Reference_Systems_Service) in [Bulletin C](https://datacenter.iers.org/data/latestVersion/bulletinC.txt) since 1994. The API is available at [https://leap.webclock.io/leap.json](https://leap.webclock.io/leap.json).

The master branch of this repository contains a daemon that periodically checks for updates to Bulletin C and notifies me via [Gotify](https://gotify.net/) when a change is detected. Updates to the API may take up to 24 hours after a new bulletin is released. In rare cases, it may take longer due to personal circumstances. However, the API will remain available as long as Cloudflare's servers are operational, and updates will continue to be made as long as I do not get hit by a bus.

This API is planned for use in an ongoing project at [webclock.io](https://webclock.io).

## API structure

The API returns an array of objects representing individual leap seconds, ordered from most recent to least recent. Each object contains the following fields:
- `utc_date`: The [UTC](https://en.wikipedia.org/wiki/Coordinated_Universal_Time) date, in [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) format, at which the leap second is applied. Note that leap seconds are applied at 00:00:00 UTC on the specified date.
- `utc_to_tai_offset`: The [International Atomic Time (TAI)](https://en.wikipedia.org/wiki/International_Atomic_Time) offset in seconds, which can be used to convert between UTC and TAI (invert the value for the reverse conversion).

### Example Response
```json
[
  {
    "utc_date": "2017-01-01",
    "utc_to_tai_offset": -37
  },
  {
    "utc_date": "2015-07-01",
    "utc_to_tai_offset": -36
  },
  {
    "utc_date": "2012-07-01",
    "utc_to_tai_offset": -35
  }
]
