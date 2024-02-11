# global-entry-notifier
Script to notify of available Global Entry enrollment appointments at a particular location. Only supports running on MacOS.

## Prerequisites

* MacOS (for notifications).
* Install [Go](https://go.dev) (if building from source).

## Usage

```shell
$ ./global-entry-notifier -h
Usage of ./global-entry-notifier:
  -days int
        number of days from today to filter slots; use 1 for current day (default 1)
  -interval duration
        polling interval for available slots e.g. 1m, 1h, 1h10m, 1d, 1d1h10m (default 1m0s)
  -limit int
        number of slots to notify (default 1)
  -location-id int
        ID of the Global Entry Enrollment Center; defaults to AUS airport (default 7820)
```

1. Find your desired enrollment
   center [here](https://ttp.cbp.dhs.gov/schedulerapi/locations/?temporary=false&inviteOnly=false&operational=true&serviceName=Global%20Entry)
   and copy its `id` field. For example: ID
   of San Francisco Global Entry Enrollment Center is `5446`.
2. Use the binary from [releases page](https://github.com/rajathagasthya/global-entry-notifier/releases) or build binary from source using `go build -o global-entry-notifier ./main.go`.
3. Run the binary with the arguments shown above. Use `Ctrl+C` to exit.
