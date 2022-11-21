package nbascrape

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	SQLITE_FILENAME string
	dbx             *sqlx.DB
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

// Get a pointer to the database handle used by the application
func GetDatabase() *sqlx.DB {
	if dbx == nil {
		dbx, _ = sqlx.Connect("sqlite3", SQLITE_FILENAME)
	}
	return dbx
}

// Runs migrations to properly configure sql database
func Migrate() error {
	database := GetDatabase()
	driver, err := sqlite.WithInstance(database.DB, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "sqlite", driver)
	if err != nil {
		return err
	}
	log.Println("Running migration.")
	if err = m.Up(); err != nil {
		return err
	}
	log.Println("Done.")

	return nil
}

// Inserts a Game into the database
func InsertGame(g *Game) error {
	sql := `INSERT INTO games
				("tipoff", "opponent", "is_home", "team_id")
				VALUES ($1, $2, $3, $4)
				RETURNING id;`
	row := dbx.QueryRow(sql, g.Tipoff.UTC().Unix(), g.Opponent, g.IsHome, g.TeamId)
	if err := row.Err(); err != nil {
		return err
	}

	var id int
	return row.Scan(&id)
}

// GetGame returns a game given that game's id
func GetGame(id int) (*Game, error) {
	g := Game{}
	err := dbx.Get(&g, "SELECT * FROM games WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	g.Tipoff = time.Unix(g.UnixTipoff, 0)
	return &g, nil
}

// GetGames returns all games in the database
func GetGames() ([]*Game, error) {
	games := make([]*Game, 0, NumberOfGames)
	rows, err := dbx.Queryx("SELECT * FROM games")
	if err != nil {
		return nil, err
	}

	g := Game{}
	for rows.Next() {
		err := rows.StructScan(&g)
		if err != nil {
			return nil, err
		}
		g.Tipoff = time.Unix(g.UnixTipoff, 0)
		games = append(games, &g)
	}
	return games, nil
}

// GetGamesForTeam returns all games for a team given that team's id
func GetGamesForTeam(teamId int) ([]*Game, error) {
	games := make([]*Game, 0, NumberOfGames)
	rows, err := dbx.Queryx("SELECT * FROM games WHERE team_id = $1", teamId)
	if err != nil {
		return nil, err
	}

	g := Game{}
	for rows.Next() {
		err = rows.StructScan(&g)
		if err != nil {
			return nil, err
		}
		games = append(games, &g)
	}

	return games, nil
}

// GetTeam selects a team from the database given that team's id
func GetTeam(id int) (*Team, error) {
	t := Team{}
	err := dbx.Get(&t, "SELECT * FROM teams WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// AllTeams selects all teams from the database
func AllTeams() ([]*Team, error) {
	// return allTeams, nil
	ts := make([]*Team, 0, NumberOfTeams)
	rows, err := dbx.Queryx("SELECT * FROM teams")
	if err != nil {
		return nil, err
	}

	t := Team{}
	for rows.Next() {
		err := rows.StructScan(&t)
		if err != nil {
			return nil, err
		}
		ts = append(ts, &t)
	}

	return ts, nil
}
