package price

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/cache"
	"github.com/bnb-chain/zkbnb/types"
)

const (
	fetchTimeout  = 3 * time.Second
	fetchLimit    = 100
	fetchInterval = 60 * time.Minute
)

type Fetcher interface {
	GetCurrencyPrice(ctx context.Context, l2Symbol string) (price float64, err error)
	Stop()
}

func NewFetcher(memCache *cache.MemCache, assetModel asset.AssetModel, oracleUrl, oracleApiKey,oracleApiSecret string) Fetcher {
	f := &fetcher{
		memCache:   memCache,
		assetModel: assetModel,
		oracleUrl:     oracleUrl,
		oracleApiKey:   oracleApiKey,
		oracleApiSecret: oracleApiSecret,
		quitCh:     make(chan struct{}),
	}
	go f.loop()
	return f
}

type fetcher struct {
	memCache   *cache.MemCache
	assetModel asset.AssetModel
	oracleUrl     string
	oracleApiKey   string
	oracleApiSecret string

	quitCh chan struct{}
}

func (f *fetcher) loop() {
	ticker := time.NewTicker(fetchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			total, err := f.assetModel.GetAssetsTotalCount()
			if err != nil {
				logx.Errorf("failed to get all assets, err: %v", err)
				continue
			}
			for i := 0; i < int(total); i += fetchLimit {
				assets, err := f.assetModel.GetAssets(int64(fetchLimit), int64(i))
				if err != nil {
					logx.Errorf("failed to get all assets, err: %v", err)
					continue
				}

				for _, asset := range assets {
					func() {
						ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
						defer cancel()
						assetPrice, err := f.GetCurrencyPrice(ctx, asset.AssetSymbol)
						if err != nil {
							logx.Errorf("failed to get all assets, err: %v", err)
							return
						}
						f.memCache.SetPrice(asset.AssetSymbol, assetPrice)
					}()
				}
			}
		case <-f.quitCh:
			return
		}
	}
}

func (f *fetcher) Stop() {
	close(f.quitCh)
}

func (f *fetcher) GetCurrencyPrice(_ context.Context, symbol string) (float64, error) {
	return f.memCache.GetPriceWithFallback(symbol, func() (interface{}, error) {
		binanceOraclePrice, err := f.getSymbolPrice(symbol)
		if err != nil {
			if err == types.BinanceOracleNotListedErr {
				return 0.0, nil
			}
			return 0.0, err
		}
		price := float64(binanceOraclePrice.Price)/math.Pow10(binanceOraclePrice.Scale)
		return price, nil
	})
}


func (f *fetcher) getSymbolPrice(symbol string) (*BinanceOraclePrice, error) {
	client := &http.Client{}
	timestamp := time.Now().Unix()
	param := make(map[string]interface{})
	param["sign"] = false
	param["symbols"] = strings.ToUpper(symbol) + "/USD"
	reqdata,_:=json.Marshal(param)
	request, err := http.NewRequest("POST", f.oracleUrl, bytes.NewReader(reqdata))
	if err != nil {
		return nil, types.HttpErrFailToRequest
	}

	needSignMessage := "sign=false&symbols="+strings.ToUpper(symbol) + "/USD" + "&x-api-timestamp=" + fmt.Sprintf("%d",timestamp)
	signature := GenHmacSha256(needSignMessage,f.oracleApiSecret)
	request.Header.Add("x-api-key", f.oracleApiKey)
	request.Header.Add("x-api-timestamp", fmt.Sprintf("%d",timestamp))
	request.Header.Add("x-api-signature", signature)
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json;charset=utf-8")
	resp, err := client.Do(request)
	if err != nil {
		return nil, types.HttpErrClientDo
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.IoErrFailToRead
	}

	//数据处理下
	binanceOracleResp := &BinanceOracleResp{}
	if err = json.Unmarshal(body, &binanceOracleResp); err != nil {
		return nil, types.JsonErrUnmarshal
	}
	datas := binanceOracleResp.Data
	if len(datas) == 0 {
		return nil,types.BinanceOracleNotListedErr
	}

	return &datas[0], nil
}


func GenHmacSha256(message string,secret string) string {
	h := hmac.New(sha256.New,[]byte(secret))
	h.Write([]byte(message))

	return hex.EncodeToString(h.Sum(nil))
}