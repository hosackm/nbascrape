package nbascrape

const (
	NumberOfTeams = 30
)

type Team struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
