package gas

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/shopspring/decimal"
)

type Speed int

const (
	SafeLow Speed = iota
	Average
	Fast
	Fastest
)

type Prices struct {
	Fast    int64 `json:"fast"`
	Fastest int64 `json:"fastest"`
	SafeLow int64 `json:"safeLow"`
	Average int64 `json:"average"`
}

func GetPrice(speed Speed) (*big.Int, error) {
	resp, err := http.Get("https://ethgasstation.info/api/ethgasAPI.json")

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad http status code: %s: %s", resp.Status, string(body))
	}

	var prices Prices
	j := json.NewDecoder(resp.Body)

	if err = j.Decode(&prices); err != nil {
		return nil, err
	}

	var price int64

	switch speed {
	case SafeLow:
		price = prices.SafeLow
	case Average:
		price = prices.Average
	case Fast:
		price = prices.Fast
	case Fastest:
		price = prices.Fastest
	default:
		return nil, errors.New("invalid speed")
	}

	if price == 0 {
		return nil, errors.New("invalid price")
	}

	return ToWei(price, 8), nil
}

// ToWei decimals to wei
func ToWei(iamount interface{}, decimals int) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case int:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}
