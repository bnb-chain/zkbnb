package price

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbas/errorcode"
	"io/ioutil"
	"net/http"

	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type price struct {
	cache multcache.MultCache
}

/*
	Func: GetCurrencyPrice
	Params: currency string
	Return: price float64, err error
	Description: get currency price cache by currency symbol
*/
func (m *price) GetCurrencyPrice(ctx context.Context, l2Symbol string) (float64, error) {
	f := func() (interface{}, error) {
		quoteMap, err := getQuotesLatest(l2Symbol)
		if err != nil {
			return 0, err
		}
		return &quoteMap, nil
	}
	var quoteType map[string]QuoteLatest
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyCurrencyPrice(), &quoteType, 10, f)
	if err != nil {
		return 0, err
	}
	res, _ := value.(*map[string]QuoteLatest)
	quoteMap := *res
	q, ok := quoteMap[l2Symbol]
	if !ok {
		return 0, errorcode.RepoErrQuoteNotExist
	}
	return q.Quote["USD"].Price, nil
}

func getQuotesLatest(l2Symbol string) (map[string]QuoteLatest, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s%s", coinMarketCap, l2Symbol)
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errorcode.RepoErrNewHttpRequest.RefineError(err.Error())
	}
	reqest.Header.Add("X-CMC_PRO_API_KEY", "cfce503f-dd3d-4847-9570-bbab5257dac8")
	reqest.Header.Add("Accept", "application/json")
	resp, err := client.Do(reqest)
	if err != nil {
		return nil, errorcode.RepoErrHttpClientDo.RefineError(err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errorcode.RepoErrIoutilReadAll.RefineError(err.Error())
	}
	currencyPrice := &currencyPrice{}
	if err = json.Unmarshal(body, &currencyPrice); err != nil {
		return nil, errorcode.RepoErrJsonUnmarshal.RefineError(err.Error() + string(body))
	}
	ifcs, ok := currencyPrice.Data.(interface{})
	if !ok {
		return nil, errorcode.RepoErrTypeAssertion
	}
	quotesLatest := make(map[string]QuoteLatest, 0)
	for _, coinObj := range ifcs.(map[string]interface{}) {
		b, err := json.Marshal(coinObj)
		if err != nil {
			return nil, errorcode.RepoErrJsonMarshal
		}
		quoteLatest := &QuoteLatest{}
		err = json.Unmarshal(b, quoteLatest)
		if err != nil {
			return nil, errorcode.RepoErrJsonUnmarshal
		}
		quotesLatest[quoteLatest.Symbol] = *quoteLatest
	}
	return quotesLatest, nil
}
