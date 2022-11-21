package nbascrape

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var (
	SQLITE_FILENAME string
	db              *sql.DB
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	dir := filepath.Join(home, ".config", "nbascrape")
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		log.Println("Creating config folder:", dir)
		err = os.Mkdir(dir, 0700)
		if err != nil {
			log.Fatal(err)
		}
	}

	SQLITE_FILENAME = filepath.Join(dir, "nbascrape.db")
}

func GetDatabase() *sql.DB {
	if db == nil {
		db, _ = sql.Open("sqlite3", SQLITE_FILENAME)
	}
	return db
}

func CreateTables() error {
	database := GetDatabase()
	driver, err := sqlite.WithInstance(database, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "sqlite", driver)
	if err != nil {
		return err
	}
	log.Println("Running migration.")
	m.Up()
	log.Println("Done.")

	return nil
}

func InsertGame(g *Game) error {
	sql := `
INSERT INTO
  "games" ("tipoff", "opponent", "is_home", "team_id")
VALUES ($1, $2, $3, $4)
RETURNING id;`

	row := db.QueryRow(sql, g.Tipoff.UTC().Unix(), g.Opponent, g.IsHome, g.TeamId)
	if err := row.Err(); err != nil {
		return err
	}

	var id int
	return row.Scan(&id)
}

type Scanner interface {
	Scan(dest ...any) error
}

func ScanGame(s Scanner) (*Game, error) {
	g := Game{}

	err := s.Scan(&g.Id, &g.TeamId, &g.UnixTipoff, &g.Opponent, &g.IsHome)
	if err != nil {
		return nil, err
	}
	g.Tipoff = time.Unix(g.UnixTipoff, 0)

	return &g, nil
}

func GetGame(id int) (*Game, error) {
	row := db.QueryRow("SELECT * FROM games WHERE id = $1", id)
	if err := row.Err(); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, nil
	}

	g, err := ScanGame(row)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func GetGames() ([]*Game, error) {
	rows, err := db.Query("SELECT * FROM games")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]*Game, 0, 82*30)
	for rows.Next() {
		g, err := ScanGame(rows)
		if err != nil {
			return nil, nil
		}
		games = append(games, g)
	}

	return games, nil
}

func GetGamesForTeam(teamId int) ([]*Game, error) {
	rows, err := db.Query("SELECT * FROM games WHERE team_id = $1", teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]*Game, 0, 82)
	for rows.Next() {
		g, err := ScanGame(rows)
		if err != nil {
			return nil, err
		}

		games = append(games, g)
	}

	return games, nil
}
