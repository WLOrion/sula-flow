package adapters

import (
	"encoding/json"
	"os"

	"github.com/WLOrion/sula-flow/internal/domain"
)

type JSONRepository struct {
	FilePath string
}

func NewJSONRepository(path string) *JSONRepository {
	return &JSONRepository{FilePath: path}
}

func (r *JSONRepository) Save(transfers []domain.Player) error {
	data, err := json.MarshalIndent(transfers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.FilePath, data, 0644)
}
