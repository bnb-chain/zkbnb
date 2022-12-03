package price

type status struct {
	Timestamp    string  `json:"timestamp"`
	ErrorCode    int     `json:"error_code"`
	ErrorMessage *string `json:"error_message"`
	Elapsed      int     `json:"elapsed"`
	CreditCount  int     `json:"credit_count"`
}

type Quote struct {
	Price            float64 `json:"price"`
	Volume24H        float64 `json:"volume_24h"`
	PercentChange1H  float64 `json:"percent_change_1h"`
	PercentChange24H float64 `json:"percent_change_24h"`
	PercentChange7D  float64 `json:"percent_change_7d"`
	MarketCap        float64 `json:"market_cap"`
	LastUpdated      string  `json:"last_updated"`
}

type QuoteLatest struct {
	ID                float64          `json:"id"`
	Name              string           `json:"name"`
	Symbol            string           `json:"symbol"`
	Slug              string           `json:"slug"`
	CirculatingSupply float64          `json:"circulating_supply"`
	TotalSupply       float64          `json:"total_supply"`
	MaxSupply         float64          `json:"max_supply"`
	DateAdded         string           `json:"date_added"`
	NumMarketPairs    float64          `json:"num_market_pairs"`
	CMCRank           float64          `json:"cmc_rank"`
	LastUpdated       string           `json:"last_updated"`
	Quote             map[string]Quote `json:"quote"`
}

type currencyPrice struct {
	Status status      `json:"status"`
	Data   interface{} `json:"data"`
}
