package nbascrape

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	R  *mux.Router
	DB *sql.DB
}

func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getTz(r *http.Request) string {
	tz := r.URL.Query().Get("tz")
	if tz == "" {
		tz = "CET"
	}
	return tz
}

func (s Server) HandleGameRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	loc, err := time.LoadLocation(getTz(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	game, err := GetGame(id)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if game == nil {
		http.Error(w, "no matching game found", http.StatusNotFound)
	} else {
		game.Tipoff = game.Tipoff.In(loc)
		json.NewEncoder(w).Encode(game)
	}
}

func (s Server) HandleGamesRequest(w http.ResponseWriter, r *http.Request) {
	allGames, err := GetGames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	loc, err := time.LoadLocation(getTz(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i := range allGames {
		allGames[i].Tipoff = allGames[i].Tipoff.In(loc)
	}

	json.NewEncoder(w).Encode(struct {
		Games []*Game `json:"games"`
	}{allGames})
}

func (s Server) HandleTeamsRequest(w http.ResponseWriter, r *http.Request) {
	allTeams, err := AllTeams()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type TeamsResponse struct {
		Teams []*Team `json:"teams"`
	}

	json.NewEncoder(w).Encode(TeamsResponse{Teams: allTeams})
}

func (s Server) HandleTeamRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	row := s.DB.QueryRow("SELECT * FROM TEAMS WHERE id = $1", id)
	if err = row.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var team Team
	row.Scan(&team.Id, &team.Name)

	json.NewEncoder(w).Encode(team)
}

func (s Server) HandleTeamGamesRequest(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	loc, err := time.LoadLocation(getTz(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var team Team
	row := s.DB.QueryRow("SELECT * FROM teams WHERE id = $1", tid)
	if err = row.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	row.Scan(&team.Id, &team.Name)

	games, err := GetGamesForTeam(tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// convert timezone
	for i, _ := range games {
		games[i].Tipoff = games[i].Tipoff.In(loc)
	}

	json.NewEncoder(w).Encode(struct {
		TeamData Team    `json:"team"`
		Games    []*Game `json:"games"`
	}{team, games})
}

func (s Server) Error(w http.ResponseWriter, err string, code int) {
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, err)
}

func NewServer() *Server {
	s := &Server{R: mux.NewRouter(), DB: GetDatabase()}
	s.R.Use(JSONMiddleware)
	s.R.HandleFunc("/games", s.HandleGamesRequest)
	s.R.HandleFunc("/games/{id:[0-9]+}", s.HandleGameRequest)

	s.R.HandleFunc("/teams", s.HandleTeamsRequest)
	s.R.HandleFunc("/teams/{id:[0-9]+}", s.HandleTeamRequest)
	s.R.HandleFunc("/teams/{id:[0-9]+}/games", s.HandleTeamGamesRequest)

	s.R.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			s.Error(w, "resource not found", http.StatusNotFound)
		},
	)
	s.R.MethodNotAllowedHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			s.Error(w, "method not allowed", http.StatusNotFound)
		},
	)

	return s
}
