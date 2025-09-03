package domain

type Player struct {
	PlayerID    int      `json:"player_id"`
	PlayerName  string   `json:"player_name"`
	PlayerUrl   string   `json:"player_url"`
	Nationality string   `json:"nationality"`
	Transfer    Transfer `json:"transfer"`
}

type Club struct {
	ClubID    int    `json:"club_id"`
	ClubName  string `json:"club_name"`
	Country   string `json:"country"`
	Continent string `json:"continent,omitempty"`
}

type Transfer struct {
	From   Club    `json:"from"`
	To     Club    `json:"to"`
	FeeEUR float64 `json:"fee_eur"`
	IsLoan bool    `json:"is_loan"`
	Season string  `json:"season"`
}
