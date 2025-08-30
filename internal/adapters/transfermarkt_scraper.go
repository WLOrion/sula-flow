package adapters

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/WLOrion/sula-flow/internal/domain"

	"github.com/PuerkitoBio/goquery"
)

type IScraper interface {
	Scrape(region string, fromYear, toYear int) ([]domain.Player, error)
}

type TransfermarktScraper struct{}

func NewTransfermarktScraper() IScraper {
	return &TransfermarktScraper{}
}

func (s *TransfermarktScraper) Scrape(region string, fromYear, toYear int) ([]domain.Player, error) {
	players := []domain.Player{}

	for year := fromYear; year <= toYear; year++ {
		url := fmt.Sprintf("https://www.transfermarkt.com/transfers/saisontransfers/statistik/top/plus/1/galerie/0?saison_id=%d&transferfenster=alle&land_id=26&ausrichtung=&spielerposition_id=&altersklasse=&leihe=", year)
		doc, err := fetchDocument(url)
		if err != nil {
			return nil, err
		}

		doc.Find("table.items tbody tr").Each(func(_ int, row *goquery.Selection) {
			cols := row.Find("td")
			if cols.Length() < 17 {
				return
			}

			// PLAYER
			playerID := 0
			playerName := ""
			cols.Eq(1).Find("table.inline-table tbody tr").Each(func(i int, tr *goquery.Selection) {
				if i == 0 {
					link := tr.Find("td.hauptlink a")
					if link.Length() > 0 {
						href, exists := link.Attr("href")
						if exists {
							parts := strings.Split(href, "/")
							if len(parts) > 0 {
								idStr := parts[len(parts)-1]
								idInt, _ := strconv.Atoi(idStr)
								playerID = idInt
							}
							playerName = strings.TrimSpace(link.Text())
						}
					}
				}
			})

			if playerID == 0 || playerName == "" {
				return
			}

			// FROM CLUB
			fromClub := domain.Club{}
			cols.Eq(10).Find("td.hauptlink a").Each(func(_ int, a *goquery.Selection) {
				href, exists := a.Attr("href")
				if exists {
					parts := strings.Split(href, "/")
					for i, p := range parts {
						if p == "verein" && i+1 < len(parts) {
							id, _ := strconv.Atoi(parts[i+1])
							fromClub.ClubID = id
							break
						}
					}
				}
				fromClub.ClubName = strings.TrimSpace(a.Text())
			})

			// TO CLUB
			toClub := domain.Club{}
			cols.Eq(14).Find("td.hauptlink a").Each(func(_ int, a *goquery.Selection) {
				href, exists := a.Attr("href")
				if exists {
					parts := strings.Split(href, "/")
					for i, p := range parts {
						if p == "verein" && i+1 < len(parts) {
							id, _ := strconv.Atoi(parts[i+1])
							toClub.ClubID = id
							break
						}
					}
				}
				toClub.ClubName = strings.TrimSpace(a.Text())
			})

			// FEE
			feeText := strings.TrimSpace(cols.Eq(16).Text())
			fee, isLoan := parseFee(feeText)

			players = append(players, domain.Player{
				PlayerID:    playerID,
				PlayerName:  playerName,
				Nationality: region,
				Transfer: domain.Transfer{
					From:   fromClub,
					To:     toClub,
					FeeEUR: fee,
					IsLoan: isLoan,
					Season: fmt.Sprintf("%d/%d", year, year+1),
				},
			})
		})
	}

	return players, nil
}

func parseFee(feeText string) (float64, bool) {
	feeTextLower := strings.ToLower(strings.TrimSpace(feeText))
	isLoan := false

	switch {
	case feeTextLower == "" || feeTextLower == "-":
		return 0, false
	case strings.Contains(feeTextLower, "free transfer"):
		return 0, false
	case strings.Contains(feeTextLower, "loan fee"):
		isLoan = true
		feeText = strings.ReplaceAll(feeTextLower, "loan fee:", "")
	case strings.Contains(feeTextLower, "loan transfer"):
		isLoan = true
		return 0, isLoan
	}

	feeText = strings.ReplaceAll(feeText, "€", "")
	feeText = strings.ReplaceAll(feeText, "$", "")
	feeText = strings.TrimSpace(feeText)

	multiplier := 1.0
	if strings.Contains(feeText, "th.") || strings.Contains(feeText, "k") {
		multiplier = 1_000
		feeText = strings.ReplaceAll(feeText, "th.", "")
		feeText = strings.ReplaceAll(feeText, "k", "")
	} else if strings.Contains(feeText, "m") {
		multiplier = 1_000_000
		feeText = strings.ReplaceAll(feeText, "m", "")
	} else if strings.Contains(feeText, "bn") || strings.Contains(feeText, "b") {
		multiplier = 1_000_000_000
		feeText = strings.ReplaceAll(feeText, "bn", "")
		feeText = strings.ReplaceAll(feeText, "b", "")
	}

	feeText = strings.ReplaceAll(feeText, ",", ".")
	feeText = strings.ReplaceAll(feeText, " ", "")
	if feeText == "" {
		return 0, isLoan
	}

	value, err := strconv.ParseFloat(feeText, 64)
	if err != nil {
		return 0, isLoan
	}
	return value * multiplier, isLoan
}

func fetchDocument(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
