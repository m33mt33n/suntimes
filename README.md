# Suntimes

A simple frontend for api.sunrise-sunset.org, which displays information about
various sun related events for given date.

# Features
- Pretty prints times for below events:
  - Sunrise
  - Sunset
  - Solar noon
  - Day length
  - Civil twilight
  - Nautical twilight
  - Astronomical twilight
- Supports auto-detection of user's location using `ip-api.com`

# Installation
```shell
# Using git
git clone "https://github.com/m33mt33n/suntimes.git"
cd suntimes
CGO_ENABLED=0 go build -ldflags="-s -w" -v .
```

# Usage
Usage of suntimes:
  -city string
      city name to be used (default "Unknown")
  -coordinates string
      coordinates in lat,lon format (default "24.85468,67.02071")
  -date %Y-%m-%d
      date to get times for in %Y-%m-%d format (default "<today's date>")
  -detect-location
      detect location by using ip address
  -timezone string
      timezone to be used by default it will use $TZ environment variable. (default "<system's $TZ>")

# Examples
```shell
# Use auto-detection feature
suntimes -detect-location=true

# Or specify locatiom explicitly if auto-detection not worked for your location
suntimes -city=Lahore -coordinates='31.566,74.314' -timezone='Asia/Karachi'
```
