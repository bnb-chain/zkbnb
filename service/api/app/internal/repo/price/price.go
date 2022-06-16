package price

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
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
	// quote.Quote["USD"].Price
	err := UpdateCurrencyPriceBySymbol(ctx, l2Symbol, m.cache)
	if err != nil {
		errInfo := fmt.Sprintf("[PriceModel.GetCurrencyPrice.UpdateCurrencyPriceBySymbol] %s", err)
		logx.Error(errInfo)
		return 0, err
	}
	key := fmt.Sprintf("%s%v", cachePriceSymbolPrefix, l2Symbol)
	var returnObj interface{}
	_, err = m.cache.Get(ctx, key, &returnObj)
	if err != nil {
		errInfo := fmt.Sprintf("[PriceModel.GetCurrencyPrice.Getcache] %s %s", key, err)
		logx.Error(errInfo)
		return 0, err
	}
	_price, ok := returnObj.(float64)
	if !ok {
		return _price, ErrTypeAssertion
	}
	return _price, nil
}

/*
	Func: UpdateCurrencyPriceBySymbol
	Params:
	Return: err
	Description: update currency price cache by symbol
*/
func UpdateCurrencyPriceBySymbol(ctx context.Context, l2Symbol string, cache multcache.MultCache) error {
	latestQuotes, err := getQuotesLatest(l2Symbol)
	if err != nil {
		return err
	}
	for _, latestQuote := range latestQuotes {
		key := fmt.Sprintf("%s%s", cachePriceSymbolPrefix, latestQuote.Symbol)
		logx.Info(key, "   ", latestQuote.Quote["USD"].Price)
		if err := cache.Set(ctx, key, latestQuote.Quote["USD"].Price, 1); err != nil {
			return ErrSetCache.RefineError(err.Error())
		}
	}
	return nil
}

func getQuotesLatest(l2Symbol string) (quotesLatest []QuoteLatest, err error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s%s", coinMarketCap, l2Symbol)
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrNewHttpRequest.RefineError(err.Error())
	}
	reqest.Header.Add("X-CMC_PRO_API_KEY", "cfce503f-dd3d-4847-9570-bbab5257dac8")
	reqest.Header.Add("Accept", "application/json")
	resp, err := client.Do(reqest)
	if err != nil {
		return nil, ErrHttpClientDo.RefineError(err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrIoutilReadAll.RefineError(err.Error())
	}
	// TODO: currencyPrice's interface{} looks like not necessary, body could Unmarshal to a fixed type struct
	currencyPrice := &currencyPrice{}
	err = json.Unmarshal(body, &currencyPrice)
	if err != nil {
		return nil, ErrJsonUnmarshal.RefineError(err.Error() + string(body))
	}
	ifcs, ok := currencyPrice.Data.(interface{})
	if !ok {
		return nil, ErrTypeAssertion
	}
	for _, coinObj := range ifcs.(map[string]interface{}) {
		b, err := json.Marshal(coinObj)
		if err != nil {
			return nil, ErrJsonMarshal
		}
		quoteLatest := &QuoteLatest{}
		err = json.Unmarshal(b, quoteLatest)
		if err != nil {
			return nil, ErrJsonUnmarshal
		}
		quotesLatest = append(quotesLatest, *quoteLatest)
	}
	return quotesLatest, nil
}
