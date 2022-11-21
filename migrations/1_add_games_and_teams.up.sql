PRAGMA foreign_keys = ON;

-- Create teams
CREATE TABLE teams (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(50)
);

-- Create games
CREATE TABLE games (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  team_id INTEGER,
  tipoff INTEGER,
  opponent VARCHAR(50),
  is_home BOOLEAN,
  FOREIGN KEY(team_id) REFERENCES teams(id)
);
