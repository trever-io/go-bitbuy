package bitbuy

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	WALLETS_ENDPOINT = "/api/v1/wallets"
)

/*
 {
        "id": null,
        "displayName": null,
        "symbol": "BCH",
        "balance": 2.846220260000000000,
        "reservedBalance": 0,
        "availableBalance": 2.846220260000000000,
        "fiatCurrencySymbol": "CAD",
        "fiatBalance": 456.39,
        "fiatReservedBalance": "0.00",
        "fiatAvailableBalance": "456.39"
    },
*/

type Wallet struct {
	Id                   *string     `json:"id"`
	DisplayName          *string     `json:"displayName"`
	Symbol               string      `json:"symbol"`
	Balance              json.Number `json:"balance"`
	ReservedBalance      json.Number `json:"reservedBalance"`
	AvailableBalance     json.Number `json:"availableBalance"`
	FiatCurrencySymbol   string      `json:"fiatCurrencySymbol"`
	FiatBalance          json.Number `json:"fiatBalance"`
	FiatReservedBalance  string      `json:"fiatReservedBalance"`
	FiatAvailableBalance string      `json:"fiatAvailableBalance"`
}

func (c *client) GetWallets(ctx context.Context) ([]*Wallet, error) {
	data, err := c.getRequest(ctx, WALLETS_ENDPOINT)
	if err != nil {
		return nil, fmt.Errorf("error getting wallets: %w", err)
	}

	tmp := make([]*Wallet, 0)
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling get wallets response: %w", err)
	}

	return tmp, nil
}
