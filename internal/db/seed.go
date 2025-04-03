package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/lib/pq"
)

const (
	seedDir = "./cmd/cli/seed/json"
)

var (
	ErrInvalidNumberOfValues = errors.New("invalid number of values")
)

type Seed struct {
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
	Values  [][]any  `json:"values"`
}

func SeedDatabase(db *sql.DB) {
	files, err := os.ReadDir(seedDir)
	if err != nil {
		log.Println("error reading seed directory:", err)
		return
	}

	for _, file := range files {
		f := strings.Split(file.Name(), ".")

		if file.IsDir() || f[1] != "json" {
			continue
		}

		seedFile, err := os.ReadFile(filepath.Join(seedDir, file.Name()))
		if err != nil {
			log.Println("error reading seed file:", err)
			return
		}

		var seed Seed
		if err := json.Unmarshal(seedFile, &seed); err != nil {
			log.Println("error unmarshaling seed file:", err)
			return
		}

		if err := seed.Insert(db); err != nil {
			log.Println("error inserting seed:", err)
			return
		}

		log.Println("seed inserted successfully")
	}
}

func (s *Seed) Insert(db *sql.DB) error {
	query := s.buildInsertQuery()

	for _, values := range s.Values {
		if len(values) != len(s.Columns) {
			return ErrInvalidNumberOfValues
		}

		// Convert any slice of strings to pq.Array
		for i, value := range values {
			if v, ok := value.([]any); ok {
				values[i] = pq.Array(v)
			}
		}

		_, err := db.Exec(query, values...)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (s *Seed) buildInsertQuery() string {
	placeholders := make([]string, len(s.Columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1) // Use $n placeholders
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		s.Table,
		strings.Join(s.Columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return query
}
