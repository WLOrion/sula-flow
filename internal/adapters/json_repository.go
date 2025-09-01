package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type JSONRepository interface {
	Save(path string, data interface{}) error
	Load(path string, dest interface{}) error
}

type jsonRepository struct{}

func NewJSONRepository() JSONRepository {
	return &jsonRepository{}
}

// Salva os dados em JSON criando diretórios se necessário
func (r *jsonRepository) Save(path string, data interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Carrega dados de um arquivo JSON
func (r *jsonRepository) Load(path string, dest interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(dest)
}
