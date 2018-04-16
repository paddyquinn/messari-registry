package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/paddyquinn/messari/util"
)

// CryptoAsset is a a representation of user input of a crypto asset
type CryptoAsset struct {
	ID            *string  `json:"id"`
	Name          *string  `json:"name"`
	Symbol        *string  `json:"symbol"`
	Description   *string  `json:"description"`
	Team          []string `json:"team"`
	ICOAmount     *float64 `json:"icoAmount"`
	BlockReward   *float64 `json:"blockReward"`
	FundingStatus *string  `json:"fundingStatus"`
	FoundedDate   *string  `json:"foundedDate"`
	CoinType      *string  `json:"coinType"`
	Website       *string  `json:"website"`
}

// NewCryptoAsset creates a new crypto asset from a request body (typically passed in via POST JSON).
func NewCryptoAsset(requestBody io.ReadCloser) (*CryptoAsset, error) {
	cryptoAsset := &CryptoAsset{}
	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(cryptoAsset)
	if err != nil {
		return nil, err
	}

	return cryptoAsset, nil
}

// Format formats the fields of a crypto asset to make them more human readable.
func (asset *CryptoAsset) Format() {
	if asset.Name != nil {
		asset.Name = capitalize(strings.TrimSpace(*asset.Name))
	}

	if asset.Symbol != nil {
		symbol := strings.ToUpper(strings.TrimSpace(*asset.Symbol))
		asset.Symbol = &symbol
	}

	if asset.FundingStatus != nil {
		fundingStatus := strings.ToUpper(strings.TrimSpace(*asset.FundingStatus))
		asset.FundingStatus = &fundingStatus
	}

	if asset.CoinType != nil {
		asset.CoinType = capitalize(strings.TrimSpace(*asset.CoinType))
	}
}

// Normalize normalizes all of the data within a crypto asset by trimming whitespace and lowercasing everything so that
// data that enters our database is consistent. This function returns an error if the crypto asset contains a
// non-numeric id string, the ICO amount or block reward are below 0, or the founded date is not ISO-8601 compliant.
func (asset *CryptoAsset) Normalize() (int, error) {
	var (
		id  int
		err error
	)

	if asset.ID != nil {
		id, err = strconv.Atoi(*asset.ID)
		if err != nil {
			return -1, fmt.Errorf("invalid id: %s", *asset.ID)
		}
	}

	if asset.Name != nil {
		asset.Name = util.Normalize(*asset.Name)
	}

	if asset.Symbol != nil {
		asset.Symbol = util.Normalize(*asset.Symbol)
	}

	// Descriptions are only trimmed because their capitalization could be ambiguous.
	if asset.Description != nil {
		description := strings.TrimSpace(*asset.Description)
		asset.Description = &description
	}

	// Team member names are only trimmed because team members may go by pseudonyms which may not follow a consistent
	// pattern.
	if asset.Team != nil {
		normalizedTeam := make([]string, len(asset.Team))
		for idx, name := range asset.Team {
			normalizedTeam[idx] = strings.TrimSpace(name)
		}
		asset.Team = normalizedTeam
	}

	if asset.ICOAmount != nil && *asset.ICOAmount < 0 {
		return -1, errors.New("ICO amount cannot be negative")
	}

	if asset.BlockReward != nil && *asset.BlockReward < 0 {
		return -1, errors.New("block reward cannot be negative")
	}

	if asset.FundingStatus != nil {
		asset.FundingStatus = util.Normalize(*asset.FundingStatus)
	}

	if asset.FoundedDate != nil {
		foundedDate := strings.TrimSpace(*asset.FoundedDate)

		// Ensure the date is ISO-8601 formatted and is not in the future.
		date, err := time.Parse("2006-01-02", foundedDate)
		if err != nil || date.After(time.Now()) {
			return -1, errors.New("date must be an ISO-8601 date in the past")
		}

		asset.FoundedDate = &foundedDate
	}

	if asset.CoinType != nil {
		asset.CoinType = util.Normalize(*asset.CoinType)
	}

	if asset.Website != nil {
		asset.Website = util.Normalize(*asset.Website)
	}

	return id, nil
}

// capitalize capitalizes the first letter of a string.
func capitalize(str string) *string {
	capitalizedStr := strings.Title(strings.ToLower(str))
	return &capitalizedStr
}
