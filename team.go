package nbascrape

type Team struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func ScanTeam(s Scanner) (*Team, error) {
	var t Team
	err := s.Scan(&t.Id, &t.Name)
	if err != nil {
		return nil, err
	}

	return &t, err
}

func GetTeam(id int) (*Team, error) {
	row := GetDatabase().QueryRow("SELECT * FROM teams WHERE id = $1", id)
	if err := row.Err(); err != nil {
		return nil, err
	}

	t, err := ScanTeam(row)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func AllTeams() ([]*Team, error) {
	db := GetDatabase()
	rows, err := db.Query("SELECT * FROM teams")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	allTeams := make([]*Team, 0, 30)
	for rows.Next() {
		t, err := ScanTeam(rows)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, t)
	}

	return allTeams, nil
}
