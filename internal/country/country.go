package country

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Country struct {
	ID        int
	Name      string
	Continent string
}

type CountryStore struct {
	byID   map[int]Country
	byName map[string]Country
}

func LoadCountries(csvPath string) (*CountryStore, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open countries CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read countries CSV: %w", err)
	}

	byID := make(map[int]Country)
	byName := make(map[string]Country)

	for _, row := range records {
		if len(row) < 3 {
			continue
		}

		id, err := strconv.Atoi(row[0])
		if err != nil {
			continue
		}

		name := sanitizeName(row[1])
		continent := row[2]

		country := Country{
			ID:        id,
			Name:      row[1],
			Continent: continent,
		}

		byID[id] = country
		byName[name] = country
	}

	return &CountryStore{byID: byID, byName: byName}, nil
}

func (cs *CountryStore) GetByID(id int) (string, bool) {
	c, ok := cs.byID[id]
	return c.Name, ok
}

func (cs *CountryStore) GetByName(name string) (int, bool) {
	c, ok := cs.byName[sanitizeName(name)]
	return c.ID, ok
}

func (cs *CountryStore) GetContByName(name string) (string, bool) {
	c, ok := cs.byName[sanitizeName(name)]
	return c.Continent, ok
}

func sanitizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
