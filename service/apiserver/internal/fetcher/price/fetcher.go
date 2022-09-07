package price

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/cache"
	"github.com/bnb-chain/zkbnb/types"
)

type Fetcher interface {
	GetCurrencyPrice(ctx context.Context, l2Symbol string) (price float64, err error)
}

func NewFetcher(memCache *cache.MemCache, cmcUrl, cmcToken string) Fetcher {
	return &fetcher{
		memCache: memCache,
		cmcUrl:   cmcUrl,
		cmcToken: cmcToken,
	}
}

type fetcher struct {
	memCache *cache.MemCache
	cmcUrl   string
	cmcToken string
}

func (f *fetcher) GetCurrencyPrice(_ context.Context, symbol string) (float64, error) {
	return f.memCache.GetPriceWithFallback(symbol, func() (interface{}, error) {
		quoteMap, err := f.getLatestQuotes(symbol)
		if err != nil {
			if err == types.CmcNotListedErr {
				return 0.0, nil
			}
			return 0.0, err
		}
		q, ok := quoteMap[symbol]
		if !ok {
			return 0.0, nil
		}
		price := q.Quote["USD"].Price
		return price, nil
	})
}

func (f *fetcher) getLatestQuotes(symbol string) (map[string]QuoteLatest, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s%s", f.cmcUrl, symbol)
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, types.HttpErrFailToRequest
	}
	reqest.Header.Add("X-CMC_PRO_API_KEY", f.cmcToken)
	reqest.Header.Add("Accept", "application/json")
	resp, err := client.Do(reqest)
	if err != nil {
		return nil, types.HttpErrClientDo
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.IoErrFailToRead
	}
	currencyPrice := &currencyPrice{}
	if err = json.Unmarshal(body, &currencyPrice); err != nil {
		return nil, types.JsonErrUnmarshal
	}
	dataMap, ok := currencyPrice.Data.(map[string]interface{})
	if !ok { //the currency not listed on cmc
		return nil, types.CmcNotListedErr
	}
	quotesLatest := make(map[string]QuoteLatest, 0)
	for _, coinObj := range dataMap {
		b, err := json.Marshal(coinObj)
		if err != nil {
			return nil, types.JsonErrMarshal
		}
		quoteLatest := &QuoteLatest{}
		err = json.Unmarshal(b, quoteLatest)
		if err != nil {
			return nil, types.JsonErrUnmarshal
		}
		quotesLatest[quoteLatest.Symbol] = *quoteLatest
	}
	return quotesLatest, nil
}
