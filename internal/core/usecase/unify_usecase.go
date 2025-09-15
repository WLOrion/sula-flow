package usecase

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/WLOrion/sula-flow/internal/country"
	"github.com/WLOrion/sula-flow/internal/domain"
	"github.com/WLOrion/sula-flow/internal/utils"
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

		// p.Transfer.From.Continent, ok = uc.countryStore.GetContByName(p.Transfer.From.Country)
		// if !ok {
		// 	return 0, 0, fmt.Errorf("invalid country founded: %s, %d", p.Transfer.From.Country, p.PlayerID)
		// }

		// p.Transfer.To.Continent, ok = uc.countryStore.GetContByName(p.Transfer.To.Country)
		// if !ok {
		// 	return 0, 0, fmt.Errorf("invalid country founded: %s, %d", p.Transfer.From.Country, p.PlayerID)
		// }

		unifiedMap[p.PlayerID] = up
	}

	unifiedPlayers := []UnifiedPlayer{}
	for _, up := range unifiedMap {
		contName, cok := uc.countryStore.GetContByName(countryName)
		if !cok {
			return 0, 0, fmt.Errorf("invalid country name: %d", countryID)
		}

		fmt.Printf("------ Name: %s, ID: %d ------\n", up.PlayerName, up.PlayerID)
		up.Transfers, err = uc.fetchTransferHistory(up.PlayerID, countryName, contName)
		if err != nil {
			return 0, 0, err
		}

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

func (uc *UnifyUsecase) fetchTransferHistory(playerID int, countryName string, continentName string) ([]domain.Transfer, error) {
	url := fmt.Sprintf("https://www.transfermarkt.com/ceapi/transferHistory/list/%d", playerID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to request transfer history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transfer history: %w", err)
	}

	transfersJSON, ok := raw["transfers"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected transfers format")
	}

	var transfers []domain.Transfer
	for _, t := range transfersJSON {
		tr := t.(map[string]interface{})

		// season
		season, _ := tr["season"].(string)

		// fee
		feeStr, _ := tr["fee"].(string)
		feeEUR, isLoan := utils.ParseFee(feeStr) // ⚡ usar a mesma função do transfermarkt_scraper.go

		// from
		from := uc.parseClub(tr["from"], countryName, continentName)

		// to
		to := uc.parseClub(tr["to"], countryName, continentName)

		transfers = append(transfers, domain.Transfer{
			From:   from,
			To:     to,
			FeeEUR: feeEUR,
			IsLoan: isLoan,
			Season: season,
		})
	}

	return transfers, nil
}

func (uc *UnifyUsecase) parseClub(raw interface{}, countryName string, continentName string) domain.Club {
	m := raw.(map[string]interface{})
	clubName, _ := m["clubName"].(string)
	href, _ := m["href"].(string)
	flag, _ := m["countryFlag"].(string)

	// if clubName == "Without Club" {
	// 	return domain.Club{
	// 		ClubID:    515,
	// 		ClubName:  "Without Club",
	// 		Country:   "OOF",
	// 		Continent: "OOF",
	// 	}
	// }

	// if clubName == "Retired" {
	// 	return domain.Club{
	// 		ClubID:    123,
	// 		ClubName:  "Retired",
	// 		Country:   "OOF",
	// 		Continent: "OOF",
	// 	}
	// }

	// if clubName == "Unknown" {
	// 	return domain.Club{
	// 		ClubID:    75,
	// 		ClubName:  "Unknown",
	// 		Country:   "OOF",
	// 		Continent: "OOF",
	// 	}
	// }

	// if clubName == "Career break" {
	// 	return domain.Club{
	// 		ClubID:    2113,
	// 		ClubName:  "Career break",
	// 		Country:   "OOF",
	// 		Continent: "OOF",
	// 	}
	// }

	// if clubName == "Ban" {
	// 	return domain.Club{
	// 		ClubID:    2077,
	// 		ClubName:  "Ban",
	// 		Country:   "OOF",
	// 		Continent: "OOF",
	// 	}
	// }

	// if clubName == "Own Youth" {
	// 	return domain.Club{
	// 		ClubID:    12604,
	// 		ClubName:  "Own Youth",
	// 		Country:   "OOF",
	// 		Continent: "OOF",
	// 	}
	// }

	clubID := extractClubID(href)
	countryID := extractCountryID(flag)
	ctrName, cont := uc.countryInfos(countryID)

	return domain.Club{
		ClubID:    clubID,
		ClubName:  clubName,
		Country:   ctrName,
		Continent: cont,
	}
}

func (uc *UnifyUsecase) countryInfos(id int) (string, string) {
	name, ok := uc.countryStore.GetByID(id)

	if !ok {
		panic(fmt.Sprintf("Cannot parse country id: %d\n", id))
	}

	cont, cok := uc.countryStore.GetContByName(name)

	if !cok {
		panic(fmt.Sprintf("Cannot parse country name: %s\n", name))
	}

	return name, cont
}

func extractClubID(href string) int {
	re := regexp.MustCompile(`/verein/(\d+)`)
	matches := re.FindStringSubmatch(href)
	if len(matches) == 2 {
		id, _ := strconv.Atoi(matches[1])
		return id
	}
	return 0
}

func extractCountryID(flagURL string) int {
	re := regexp.MustCompile(`/verysmall/(\d+)\.png`)
	matches := re.FindStringSubmatch(flagURL)

	// fmt.Printf("URL: %s, %v\n", flagURL, matches)
	if len(matches) == 2 {
		id, _ := strconv.Atoi(matches[1])
		return id
	}

	return -1
}
