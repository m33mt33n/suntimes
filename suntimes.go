//┌───────────────────────────────────────────────────────────────────────────┐
//	File           suntimes.go
//	Description    A simple frontend for api.sunrise-sunset.org.
//	Version        0.1.0 alpha
//	Author         Moin Khan <m33mt33n>
//	Source         https://github.com/m33mt33n/suntimes
//	License        GNU General Public License v3.0 or later (see LICENSE)
//	Created        October 31, 2025 18:50
//	Last Updated   November 02, 2025 17:20
//└───────────────────────────────────────────────────────────────────────────┘

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const report string = `
Sunrise                       {{.Sunrise}}
Sunset                        {{.Sunset}}
Solar noon                    {{.Solar_noon}}
Day length                    {{.Day_length}}
Twilight
- Civil (beg)                 {{.Civil_twilight_begin}}
- Civil (end)                 {{.Civil_twilight_end}}
- Nautical (beg)              {{.Nautical_twilight_begin}}
- Nautical (end)              {{.Nautical_twilight_end}}
- Astronomical (beg)          {{.Astronomical_twilight_begin}}
- Astronomical (end)          {{.Astronomical_twilight_end}}
`

var offline bool = false

type Timestamp string
type Duration string

type Times struct {
	Sunrise                     Timestamp `json:"sunrise"`
	Sunset                      Timestamp `json:"sunset"`
	Solar_noon                  Timestamp `json:"solar_noon"`
	Day_length                  Duration  `json:"day_length"`
	Civil_twilight_begin        Timestamp `json:"civil_twilight_begin"`
	Civil_twilight_end          Timestamp `json:"civil_twilight_end"`
	Nautical_twilight_begin     Timestamp `json:"nautical_twilight_begin"`
	Nautical_twilight_end       Timestamp `json:"nautical_twilight_end"`
	Astronomical_twilight_begin Timestamp `json:"astronomical_twilight_begin"`
	Astronomical_twilight_end   Timestamp `json:"astronomical_twilight_end"`
}

type SunAPIResp struct {
	Times  Times  `json:"results"`
	status string `json:"status"`
	tzid   string `json:"tzid"`
}

type Location struct {
	City      string
	Timezone  string
	Latitude  float64
	Longitude float64
}

func (ts *Timestamp) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ts = new(Timestamp)
		return
	}
	timestamp, err := time.Parse("2006-01-02T15:04:05-07:00", s)
	check(err)
	*ts = Timestamp(timestamp.Format("15:04:05"))
	return
}

func (du *Duration) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		du = new(Duration)
		return
	}
	duration, err := strconv.Atoi(s)
	check(err)
	d := time.Duration(duration) * time.Second
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	sec := d / time.Second
	//*du = Duration(fmt.Sprintf("%02d:%02d:%02d", h, m, sec))
	*du = Duration(fmt.Sprintf("%dh %dm %ds", h, m, sec))
	return
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func get_data_from_api(url string, api string) []uint8 {
	var (
		data_bytes []uint8
		err        error
	)
	if offline {
		data_bytes, err = os.ReadFile(get_dummy_data_file(api))
	} else {
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		check(err)
		req.Header.Set(
			"User-Agent", "Opera/9.80 (X11; Linux i686; U; ru) Presto/2.8.131 Version/11.11",
		)
		response, err := client.Do(req)
		check(err)
		data_bytes, err = io.ReadAll(response.Body)
	}
	check(err)
	return data_bytes
}

// source: https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
func file_exists(fpath string) bool {
	info, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func get_dummy_data_file(api string) string {
	var (
		envname string = fmt.Sprintf("_dummy_data_%s", api)
		fpath   string
		ok      bool
	)
	if fpath, ok = os.LookupEnv(envname); !ok {
		log.Fatal(errors.New(fmt.Sprintf("environment variable `%s` is not set!\n", envname)))
	}
	if !file_exists(fpath) {
		log.Fatal(errors.New(fmt.Sprintf("file not exists: `%s`\n", fpath)))
	}
	return fpath
}

func get_location_from_ip() Location {
	// get location from ip address
	var (
		data map[string]any
		loc  Location
	)
	data_bytes := get_data_from_api(
		"http://ip-api.com/json", "ipapi",
	)
	err := json.Unmarshal(data_bytes, &data)
	check(err)
	loc.City = data["city"].(string)
	loc.Timezone = data["timezone"].(string)
	loc.Latitude = data["lat"].(float64)
	loc.Longitude = data["lon"].(float64)
	return loc
}

func get_times(loc Location, date string) SunAPIResp {
	// get data from sunrise-sunset.org api
	// NOTE: times calculated by api are based on coordinates, so there could be a minor difference
	// in values for a single city when different coordinates provided.
	var data SunAPIResp
	data_bytes := get_data_from_api(
		fmt.Sprintf(
			"https://api.sunrise-sunset.org/json?lat=%f&lng=%f&tzid=%s&formatted=0&date=%s",
			loc.Latitude, loc.Longitude, loc.Timezone, date,
		),
		"suntimes",
	)
	err := json.Unmarshal(data_bytes, &data)
	check(err)
	return data
}

func main() {
	var (
		city            string
		timezone        string
		coordinates     string
		date            string
		detect_location bool
	)
	if value, ok := os.LookupEnv("_suntimes_offline"); ok && value == "1" {
		offline = true
	}
	default_date := time.Now().Format("2006-01-02")
	var (
		default_timezone string
		ok               bool
	)
	if default_timezone, ok = os.LookupEnv("TZ"); !ok {
		default_timezone = "UTC"
	}
	flag.StringVar(&city, "city", "Unknown", "city name to be used")
	flag.StringVar(&timezone, "timezone", default_timezone, "timezone to be used by default it will use $TZ environment variable.")
	flag.StringVar(&coordinates, "coordinates", "24.85468,67.02071", "coordinates in lat,lon format")
	flag.StringVar(&date, "date", default_date, "date to get times for in `%Y-%m-%d` format")
	flag.BoolVar(&detect_location, "detect-location", false, "detect location by using ip address")
	flag.Parse()
	var (
		loc Location
		err error
	)
	if detect_location {
		loc = get_location_from_ip()
	} else {
		loc.City = city
		loc.Timezone = timezone
		lat_lon := strings.Split(coordinates, ",")
		if loc.Latitude, err = strconv.ParseFloat(lat_lon[0], 64); err != nil {
			check(err)
		} else {
			if loc.Latitude < -90.0 || loc.Latitude > 90.0 {
				log.Fatal(errors.New(fmt.Sprintf("invalid latitude: %f\n", loc.Latitude)))
			}
		}
		if loc.Longitude, err = strconv.ParseFloat(lat_lon[1], 64); err != nil {
			check(err)
		} else {
			if loc.Longitude < -180.0 || loc.Longitude > 180.0 {
				log.Fatal(errors.New(fmt.Sprintf("invalid longitude: %f\n", loc.Longitude)))
			}
		}
	}
	var date_obj time.Time
	if date_obj, err = time.Parse("2006-01-02", date); err != nil {
		check(err)
	}
	data := get_times(loc, date)
	fmt.Printf(
		"%s %.3f,%.3f (%s)\n%s\n",
		loc.City, loc.Latitude, loc.Longitude, loc.Timezone,
		date_obj.Format("Monday, Jan 02, 2006"),
	)
	t := template.Must(template.New("report").Parse(report))
	err = t.Execute(os.Stdout, data.Times)
	check(err)
}
