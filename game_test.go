package nbascrape

import (
	"strings"
	"testing"
	"time"

	"github.com/antchfx/htmlquery"
)

func (g Game) Equal(other Game) bool {
	return g.IsHome == other.IsHome &&
		g.Opponent == other.Opponent &&
		g.Id == other.Id &&
		g.Tipoff == other.Tipoff &&
		g.UnixTipoff == other.UnixTipoff
}

func TestNewGameFromRow(t *testing.T) {
	docs := []string{
		`<html><table><tr>
			<td>Mon Nov 21</td>
			<td>vs Houston Rockets   </td>
			<td>9:30 PM </td></tr></table>
		</html>`,
		`<html><table><tr>
			<td>Wed Jan 4</td>
			<td>@ Los Angeles Lakers</td>
			<td>3:00 AM </td></tr></table>
		</html>`,
		`<html><table><tr>
			<td>Sat Apr 15</td>
			<td>vs Orlando Magic   </td>
			<td>6:00 PM </td></tr></table>
		</html>`,
	}
	expected := []Game{
		{
			Id:         -1,
			Opponent:   "Houston Rockets",
			IsHome:     true,
			Tipoff:     time.Date(2022, 11, 21, 21, 30, 0, 0, EST),
			UnixTipoff: time.Date(2022, 11, 21, 21, 30, 0, 0, EST).Unix(),
		},
		{
			Id:         -1,
			Opponent:   "Los Angeles Lakers",
			IsHome:     false,
			Tipoff:     time.Date(2023, 1, 4, 3, 0, 0, 0, EST),
			UnixTipoff: time.Date(2023, 1, 4, 3, 0, 0, 0, EST).Unix(),
		},
		{
			Id:         -1,
			Opponent:   "Orlando Magic",
			IsHome:     true,
			Tipoff:     time.Date(2023, 4, 15, 18, 0, 0, 0, EST),
			UnixTipoff: time.Date(2023, 4, 15, 18, 0, 0, 0, EST).Unix(),
		},
	}

	for i := range expected {
		exp := expected[i]
		s := docs[i]

		doc, err := htmlquery.Parse(strings.NewReader(s))
		if err != nil {
			t.Fail()
		}

		row, err := htmlquery.Query(doc, "//tr")
		if err != nil {
			t.Fail()
		}

		g, err := NewGameFromRow(row)
		if err != nil {
			t.Fail()
		}

		if !g.Equal(exp) {
			t.Fail()
		}
	}
}
