package nbascrape

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var (
	EST         = time.Now().Location()
	CET         = time.Now().Location()
	SeasonStart = time.Date(2022, 8, 18, 0, 0, 0, 0, time.UTC)
	SeasonEnd   = time.Date(2023, 4, 9, 0, 0, 0, 0, time.UTC)
)

func init() {
	EST, _ = time.LoadLocation("EST")
	CET, _ = time.LoadLocation("CET")
}

type Game struct {
	Id         int       `json:"id"`
	Opponent   string    `json:"opponent"`
	Tipoff     time.Time `json:"tipoff"`
	IsHome     bool      `json:"is_home"`
	TeamId     int       `json:"team_id"`
	UnixTipoff int64     `json:"-"`
}

func NewGameFromRow(row *html.Node) (*Game, error) {
	cols, err := htmlquery.QueryAll(row, "//td")
	if err != nil {
		log.Fatal(err)
	}

	var colStrings []string
	for _, col := range cols[:3] {
		colStrings = append(colStrings, htmlquery.InnerText(col))
	}

	// ie. "Tue, Oct 18"
	var weekdayS string
	var monthS string
	var day int
	numScanned, err := fmt.Sscanf(colStrings[0], "%s %s %d", &weekdayS, &monthS, &day)
	if numScanned != 3 || err != nil {
		return nil, errors.New("couldn't scan date from string")
	}
	monthLUT := map[string]int{
		"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4,
		"May": 5, "Jun": 6, "Jul": 7, "Aug": 8,
		"Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	}
	month := time.Month(monthLUT[monthS])

	// ie. "@ Portland Trailblazers"
	// or "vs Portland Trailblazers"
	var versus string
	var isHome bool
	if strings.HasPrefix(colStrings[1], "vs ") {
		versus = strings.Replace(colStrings[1], "vs ", "", 1)
		isHome = true
	} else {
		versus = strings.Replace(colStrings[1], "@ ", "", 1)
		isHome = false
	}
	versus = strings.Trim(versus, " ")

	// ie. "9:30 PM "
	var t time.Time
	if strings.HasPrefix(colStrings[2], "W") || strings.HasPrefix(colStrings[2], "L") {
		t = time.Date(2022, month, day, 0, 0, 0, 0, EST)
	} else {
		var timeWithColon string
		var ampm string

		numScanned, err := fmt.Sscanf(colStrings[2], "%s %s ", &timeWithColon, &ampm)
		if numScanned != 2 || err != nil {
			return nil, errors.New("couldn't scan date from string")
		}
		timeParts := strings.Split(timeWithColon, ":")
		hour, _ := strconv.Atoi(timeParts[0])
		minute, _ := strconv.Atoi(timeParts[1])

		if ampm == "PM" {
			hour += 12
		}

		// espn report them in eastern standard time
		t = time.Date(2022, month, day, hour, minute, 0, 0, EST)
	}

	if t.Before(SeasonStart) {
		t = time.Date(t.Year()+1, t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, EST)
	}

	return &Game{Id: -1, Tipoff: t, Opponent: versus, IsHome: isHome}, nil
}
