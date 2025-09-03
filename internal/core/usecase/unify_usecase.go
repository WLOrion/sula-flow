package usecase

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/WLOrion/sula-flow/internal/country"
	"github.com/WLOrion/sula-flow/internal/domain"
)

type UnifiedPlayer struct {
	PlayerID    int               `json:"player_id"`
	PlayerName  string            `json:"player_name"`
	PlayerUrl   string            `json:"player_url"`
	Nationality string            `json:"nationality"`
	Transfers   []domain.Transfer `json:"transfers"`
}

type UnifyUsecase struct {
	countryStore *country.CountryStore
}

func NewUnifyUsecase(cs *country.CountryStore) *UnifyUsecase {
	return &UnifyUsecase{
		countryStore: cs,
	}
}

// UnifyTransfers unifica todas as transfers de um país, salvando o resultado em JSON
func (uc *UnifyUsecase) UnifyTransfers(countryID int) (int, int, error) {
	countryName, ok := uc.countryStore.GetByID(countryID)
	if !ok {
		return 0, 0, fmt.Errorf("invalid country id %d", countryID)
	}

	countryPath := sanitizeName(countryName)
	transfersDir := filepath.Join("transfers", countryPath)

	files, err := os.ReadDir(transfersDir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read transfers dir: %w", err)
	}

	allTransfers := []domain.Player{}
	firstYear, lastYear := 0, 0

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !strings.HasPrefix(name, "transfer_") || !strings.HasSuffix(name, ".json") {
			continue
		}

		parts := strings.Split(name, "_")
		if len(parts) < 3 {
			continue
		}

		yearPart := strings.TrimSuffix(parts[len(parts)-1], ".json")
		var year int
		fmt.Sscanf(yearPart, "%d", &year)
		if firstYear == 0 || year < firstYear {
			firstYear = year
		}
		if year > lastYear {
			lastYear = year
		}

		path := filepath.Join(transfersDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return 0, 0, err
		}

		var players []domain.Player
		if err := json.Unmarshal(data, &players); err != nil {
			return 0, 0, err
		}

		for _, p := range players {
			allTransfers = append(allTransfers, p)
		}
	}

	// unifica por jogador
	unifiedMap := map[int]UnifiedPlayer{}
	for _, p := range allTransfers {
		up, exists := unifiedMap[p.PlayerID]
		if !exists {
			up = UnifiedPlayer{
				PlayerID:    p.PlayerID,
				PlayerName:  p.PlayerName,
				PlayerUrl:   p.PlayerUrl,
				Nationality: p.Nationality,
				Transfers:   []domain.Transfer{},
			}
		}

		p.Transfer.From.Continent, ok = uc.countryStore.GetContByName(p.Transfer.From.Country)
		if !ok {
			return 0, 0, fmt.Errorf("invalid country founded: %s", p.Transfer.From.Country)
		}

		p.Transfer.To.Continent, ok = uc.countryStore.GetContByName(p.Transfer.To.Country)
		if !ok {
			return 0, 0, fmt.Errorf("invalid country founded: %s", p.Transfer.From.Country)
		}

		up.Transfers = append(up.Transfers, p.Transfer)
		unifiedMap[p.PlayerID] = up
	}

	unifiedPlayers := []UnifiedPlayer{}
	for _, up := range unifiedMap {
		unifiedPlayers = append(unifiedPlayers, up)
	}

	// cria pasta de saída
	outDir := filepath.Join("unified_transfers", fmt.Sprintf("%d_%d", firstYear, lastYear))
	if err := os.MkdirAll(outDir, fs.ModePerm); err != nil {
		return 0, 0, err
	}

	outFile := filepath.Join(outDir, fmt.Sprintf("%s.json", countryPath))
	data, _ := json.MarshalIndent(unifiedPlayers, "", "  ")
	if err := os.WriteFile(outFile, data, 0644); err != nil {
		return 0, 0, err
	}

	return len(allTransfers), len(unifiedPlayers), nil
}
