package price

import "github.com/zecrey-labs/zecrey-legend/pkg/zerror"

var (
	cachePriceSymbolPrefix = "cache:zecrey-legend:cachePriceSymbolPrefix"
	coinMarketCap          = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol="
)

var (
	ErrNewHttpRequest = zerror.New(40000, "http.NewRequest err")
	ErrHttpClientDo   = zerror.New(40001, "http.Client.Do err")
	ErrIoutilReadAll  = zerror.New(40002, "ioutil.ReadAll err")
	ErrJsonUnmarshal  = zerror.New(40003, "json.Unmarshal err")
	ErrJsonMarshal    = zerror.New(40004, "json.Marshal err")
	ErrTypeAssertion  = zerror.New(40005, "type assertion err")
	ErrSetCache       = zerror.New(40006, "set cache err")
)

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
	status status      `json:"status"`
	data   interface{} `json:"data"`
}
