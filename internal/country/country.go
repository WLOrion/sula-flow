package country

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Country struct {
	idToName map[int]string
	nameToID map[string]int
}

// Carrega o CSV e retorna um objeto Country
func LoadCountries(path string) (*Country, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	idToName := make(map[int]string)
	nameToID := make(map[string]int)

	for i, row := range records {
		if i == 0 {
			continue // pula header
		}
		if len(row) < 2 {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(row[0]))
		if err != nil {
			continue
		}
		name := strings.TrimSpace(row[1])
		idToName[id] = name
		nameToID[strings.ToLower(name)] = id
	}

	return &Country{
		idToName: idToName,
		nameToID: nameToID,
	}, nil
}

// Retorna o nome do país dado o ID
func (c *Country) NameByID(id int) (string, error) {
	name, ok := c.idToName[id]
	if !ok {
		return "", fmt.Errorf("country ID %d not found", id)
	}
	return name, nil
}

// Retorna o ID do país dado o nome
func (c *Country) IDByName(name string) (int, error) {
	id, ok := c.nameToID[strings.ToLower(name)]
	if !ok {
		return 0, fmt.Errorf("country name %s not found", name)
	}
	return id, nil
}
