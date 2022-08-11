package price

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/bnb-chain/zkbas/errorcode"
)

const cacheKey = "p:"

type Fetcher interface {
	GetCurrencyPrice(ctx context.Context, l2Symbol string) (price float64, err error)
}

func NewFetcher(cache *gocache.Cache) Fetcher {
	return &fetcher{
		cache: cache,
		//todo: put into config files
		cmcUrl:   "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=",
		cmcToken: "cfce503f-dd3d-4847-9570-bbab5257dac8",
	}
}

type fetcher struct {
	cache    *gocache.Cache
	cmcUrl   string
	cmcToken string
}

/*
	Func: GetCurrencyPrice
	Params: currency string
	Return: fetcher float64, err error
	Description: get currency fetcher cache by currency symbol
*/
func (f *fetcher) GetCurrencyPrice(ctx context.Context, l2Symbol string) (float64, error) {
	var price float64
	cached, hit := f.cache.Get(cacheKey + l2Symbol)
	if hit {
		price = cached.(float64)
		return price, nil
	}

	quoteMap, err := f.getLatestQuotes(l2Symbol)
	if err != nil {
		return 0, err
	}
	q, ok := quoteMap[l2Symbol]
	if !ok {
		return 0, errorcode.AppErrQuoteNotExist
	}
	price = q.Quote["USD"].Price
	f.cache.Set(cacheKey+l2Symbol, price, time.Millisecond*500)
	return price, nil
}

func (f *fetcher) getLatestQuotes(l2Symbol string) (map[string]QuoteLatest, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s%s", f.cmcUrl, l2Symbol)
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errorcode.HttpErrFailToRequest
	}
	reqest.Header.Add("X-CMC_PRO_API_KEY", f.cmcToken)
	reqest.Header.Add("Accept", "application/json")
	resp, err := client.Do(reqest)
	if err != nil {
		return nil, errorcode.HttpErrClientDo
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errorcode.IoErrFailToRead
	}
	currencyPrice := &currencyPrice{}
	if err = json.Unmarshal(body, &currencyPrice); err != nil {
		return nil, errorcode.JsonErrUnmarshal
	}
	ifcs, ok := currencyPrice.Data.(interface{})
	if !ok {
		return nil, errors.New("type conversion error")
	}
	quotesLatest := make(map[string]QuoteLatest, 0)
	for _, coinObj := range ifcs.(map[string]interface{}) {
		b, err := json.Marshal(coinObj)
		if err != nil {
			return nil, errorcode.JsonErrMarshal
		}
		quoteLatest := &QuoteLatest{}
		err = json.Unmarshal(b, quoteLatest)
		if err != nil {
			return nil, errorcode.JsonErrUnmarshal
		}
		quotesLatest[quoteLatest.Symbol] = *quoteLatest
	}
	return quotesLatest, nil
}
