package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/hosackm/nbascrape"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

const (
	NBA_BASE_URL    = "https://www.espn.com"
	SCHEDULE_AFFIX  = "/nba/team/schedule/_/name"
	TEAM_LINK_AFFIX = "/nba/team/_/name"
	WARRIORS_PATH   = "/gsw/gs-warriors"
	SCRAPE_PATH     = "/nba/standings"
	PORT            = 8080
)

var rootCmd = &cobra.Command{
	Use:   "nbascrape",
	Short: "Scrape nba games or look them up.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Inside root command. You probably wanted to add a command to this one.")
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve the nbascrape api",
	Run: func(cmd *cobra.Command, args []string) {
		s := nbascrape.NewServer()
		addr := fmt.Sprintf(":%d", PORT)
		log.Println("Listening on", addr)
		log.Fatal(http.ListenAndServe(addr, s.R))
	},
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage or interact with the database.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Inside db root command. You probably wanted to add a command to this one.")
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database to most recent migration.",
	Run: func(cmd *cobra.Command, args []string) {

		err := nbascrape.CreateTables()
		if err != nil {
			log.Fatal(err)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List entries in the database.",
	Run: func(cmd *cobra.Command, args []string) {
		db := nbascrape.GetDatabase()

		rows, err := db.Query("SELECT * FROM teams")
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var team nbascrape.Team
			err := rows.Scan(&team.Id, &team.Name)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Team [%d]: %s\n", team.Id, team.Name)
		}

		games, err := nbascrape.GetGames()
		if err != nil {
			log.Fatal(err)
		}

		for _, g := range games {
			log.Printf("    Game: %#v\n", g)
		}
	},
}

var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "Scrape the links for each NBA team",
	Run: func(cmd *cobra.Command, args []string) {
		doc, err := htmlquery.LoadURL("https://www.espn.com/nba/standings")
		if err != nil {
			log.Fatal(err)
		}

		links, err := htmlquery.QueryAll(doc, "//a[@class='AnchorLink']")
		if err != nil {
			log.Fatal(err)
		}

		rawLinks := make(map[string]bool)
		for _, l := range links {
			href := htmlquery.SelectAttr(l, "href")
			if strings.HasPrefix(href, TEAM_LINK_AFFIX) {
				rawLinks[href] = true
			}
		}

		linkMap := make(map[string]string)
		for k := range rawLinks {
			// https://www.espn.com/nba/team/schedule/_/name/gsw/gs-warriors
			link := strings.Replace(k, TEAM_LINK_AFFIX, SCHEDULE_AFFIX, 1)
			// https://www.espn.com/nba/team/_/name/gsw/gs-warriors
			path := strings.Split(link, SCHEDULE_AFFIX)[1]
			// /gsw/gs-warriors
			name := strings.Split(path, "/")[2]
			name = strings.Title(strings.Replace(name, "-", " ", -1))
			// gs-warriors
			linkMap[name] = path
		}

		type Pair [2]string
		counter := make(chan int)
		pairs := make(chan Pair)
		for k, v := range linkMap {
			go func(k, v string) {
				DownloadGames(k, v)
				pairs <- Pair{k, v}
				counter <- 1
			}(k, v)
		}

		count := 0
		for i := 0; i < len(linkMap); i++ {
			<-pairs
			count += <-counter
		}
	},
}

func DownloadGames(name, path string) error {
	url := strings.Join([]string{NBA_BASE_URL, SCHEDULE_AFFIX, path}, "")
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := htmlquery.QueryAll(doc, "//tr")
	if err != nil {
		return err
	}
	if rows == nil {
		return errors.New("no rows returned from htmlquery")
	}

	db := nbascrape.GetDatabase()

	// insert team, get id, set id on game after NewGameFromRow
	log.Println("Inserting team:", name)
	sql := `INSERT INTO teams(name) VALUES ($1) RETURNING id;`
	row := db.QueryRow(sql, name)
	if err = row.Err(); err != nil {
		return err
	}

	var id int
	err = row.Scan(&id)
	if err != nil {
		return err
	}

	n := 0
	for _, row := range rows {
		g, err := nbascrape.NewGameFromRow(row)
		if err != nil || g == nil {
			continue
		}
		n++

		g.TeamId = id
		err = nbascrape.InsertGame(g)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Inserted %d games for %s\n", n, name)

	return nil
}

func main() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(scrapeCmd)
	dbCmd.AddCommand(listCmd)
	dbCmd.AddCommand(migrateCmd)
	rootCmd.Execute()
}
