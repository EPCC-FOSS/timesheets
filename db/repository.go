package db

import (
	"calendar_utility_node_for_timesheets/models"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "modernc.org/sqlite"
	"path/filepath"
)

type Repository struct {
	Conn *sql.DB
}

func NewRepository(dbFolder string) (*Repository, error) {
	dbPath := filepath.Join(dbFolder, "school_timesheets.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Profile table. Schedule stored as JSON blob.
	profilQuery := `
	CREATE TABLE IF NOT EXISTS profile (
		id INTEGER PRIMARY KEY CHECK (id = 1), --ensure single row for 1 profile
		first_name TEXT,
		last_name TEXT,
		data_json TEXT --store rest of data as JSON blob
	);`

	//Timesheet table. Daily entries stored as JSON blob.
	timesheetQuery := `
	CREATE TABLE IF NOT EXISTS timesheet (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		month INTEGER,
		year INTEGER,
		total_worked REAL,
		entries_json TEXT, --map[string]DailyEntry
		UNIQUE (month, year) -- Prevent duplicate sheets for the same month
	);`

	if _, err := conn.Exec(profilQuery); err != nil {
		return nil, fmt.Errorf("profile table init: %w", err)
	}

	if _, err := conn.Exec(timesheetQuery); err != nil {
		return nil, fmt.Errorf("timesheet table init: %w", err)
	}

	return &Repository{Conn: conn}, nil
}

/* PROFILE METHODS */

func (r *Repository) SaveProfile(p *models.Profile) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	// Upser: Inset or replace

	query := `INSERT OR REPLACE INTO profile (id, first_name, last_name, data_json) VALUES (1, ?, ?, ?)`

	_, err = r.Conn.Exec(query, p.FirstName, p.LastName, string(data))
	return err
}

func (r *Repository) GetProfile() (*models.Profile, error) {
	row := r.Conn.QueryRow(`SELECT data_json FROM profile WHERE id = 1`)
	var dataStr string

	// Scan the result into dataStr
	if err := row.Scan(&dataStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No profile found
		}
		return nil, err
	}

	// Unmarshal JSON data into Profile struct
	var p models.Profile
	if err := json.Unmarshal([]byte(dataStr), &p); err != nil {
		return nil, err
	}

	// Return the populated Profile struct
	return &p, nil
}

/* TIMESHEET METHODS */

func (r *Repository) SaveTimesheet(t models.Timesheet) error {
	// Marshal entries to JSON
	entriesData, err := json.Marshal(t.Entries)
	if err != nil {
		return err
	}

	// Insert or update timesheet
	query := `
	INSERT INTO timesheets (month, year, total_worked, entries_json)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(month, year) DO UPDATE SET
		total_worked = excluded.total_worked,
		entries_json = excluded.entries_json;
	`
	// Execute the query
	_, err = r.Conn.Exec(query, t.Month, t.Year, t.TotalWorked, string(entriesData))
	return err
}

func (r *Repository) GetTimesheets() ([]models.Timesheet, error) {
	rows, err := r.Conn.Query(`SELECT id, month, year, total_worked, entried_json FROM timesheets ORDER BY year DESC, month DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sheets []models.Timesheet
	for rows.Next() {
		var t models.Timesheet
		var jsonBlob string
		if err := rows.Scan(&t.ID, &t.Month, &t.Year, &t.TotalWorked, &jsonBlob); err != nil {
			return nil, err
		}

		//Unmarshal entries JSON
		json.Unmarshal([]byte(jsonBlob), &t.Entries)
		sheets = append(sheets, t)
	}

	return sheets, nil
}

// Helper to extract timesheet by month and year
func (r *Repository) GetTimesheetByDate(month int, year int) (*models.Timesheet, error) {
	// Get timesheet from db
	query := `SELECT id, month, year, total_worked, entried_json FROM timesheets WHERE month = ? AND year = ?`
	row := r.Conn.QueryRow(query, month, year)

	// Timesheet and blob
	var t models.Timesheet
	var jsonBlob string

	// error handling for query
	if err := row.Scan(&t.ID, &t.Month, &t.Year, &t.TotalWorked, &jsonBlob); err != nil {
		// no rows found
		if err == sql.ErrNoRows {
			return nil, nil
		}
		// Any other error
		return nil, err
	}

	//Error handling for json
	if err := json.Unmarshal([]byte(jsonBlob), &t.Entries); err != nil {
		return nil, err
	}

	//Return timesheet, no error
	return &t, nil
}
