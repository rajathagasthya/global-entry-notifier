package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type Slot struct {
	LocationId     int  `json:"locationId"`
	StartTimestamp Time `json:"startTimestamp"`
	EndTimestamp   Time `json:"endTimestamp"`
	Active         bool `json:"active"`
}

type Time struct {
	time.Time `json:"-"`
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	pt, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return err
	}
	t.Time, err = time.ParseInLocation("2006-01-02T15:04", str, pt)
	return err
}

func getSlots(u *url.URL) ([]Slot, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("response error: %w", err)
	}
	defer resp.Body.Close()

	var slots []Slot
	err = json.NewDecoder(resp.Body).Decode(&slots)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return slots, nil
}

func filterSlots(slots []Slot, days int) []Slot {
	if len(slots) == 0 || days < 0 {
		return nil
	}
	var result []Slot
	deadline := time.Now().AddDate(0, 0, int(days))
	for _, s := range slots {
		if s.StartTimestamp.IsZero() {
			continue
		}
		if s.Active && (s.StartTimestamp.Before(deadline) || s.StartTimestamp.Equal(deadline)) {
			result = append(result, s)
		}
	}
	return result
}

func notify(slots []Slot) error {
	if len(slots) == 0 {
		return nil
	}
	for _, s := range slots {
		script := fmt.Sprintf("display notification \"%s\" with title \"Global Entry Slot Available\" sound name \"Purr\"", s.StartTimestamp.Format("Mon Jan 2 15:04"))
		cmd := exec.Command("/usr/bin/osascript", "-e", script)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to notify: %w", err)
		}
	}
	return nil
}

func main() {
	locationID := flag.Int("location-id", 7820, "ID of the Global Entry Enrollment Center; defaults to AUS airport")
	limit := flag.Int("limit", 1, "number of slots to notify")
	days := flag.Int("days", 1, "number of days from today to filter slots; use 1 for current day")
	interval := flag.Duration("interval", 1*time.Minute, "polling interval for available slots e.g. 1m, 1h, 1h10m, 1d, 1d1h10m")
	flag.Parse()

	if *locationID < 1 {
		log.Fatal("location-id cannot be < 1")
	}
	if *days < 1 {
		log.Fatal("days cannot be < 1")
	}
	if *limit < 1 {
		log.Fatal("limit cannot be < 1")
	}

	// List of locations can be found at https://ttp.cbp.dhs.gov/schedulerapi/locations/?temporary=false&inviteOnly=false&operational=true&serviceName=Global%20Entry
	u, err := url.Parse(fmt.Sprintf("https://ttp.cbp.dhs.gov/schedulerapi/slots?orderBy=soonest&limit=%d&locationId=%d", *limit, *locationID))
	if err != nil {
		log.Fatalf("error parsing url: %v", err)
	}

	fmt.Printf("Waiting for available appointments at location %d\n", *locationID)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	tickDuration := *interval
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()
	for {
		select {
		case <-sigs:
			log.Println("shutting down on signal")
			return
		case <-ticker.C:
			slots, err := getSlots(u)
			if err != nil {
				log.Fatal(err)
			}
			if len(slots) == 0 {
				break
			}
			if err := notify(filterSlots(slots, *days)); err != nil {
				log.Fatal(err)
			}
		}
	}
}
