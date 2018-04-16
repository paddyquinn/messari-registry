package database

import (
	"testing"

	"github.com/paddyquinn/messari/database/models"
)

func Test_createSelectStatement(t *testing.T) {
	expectedArgs := []string{"Bitcoin", "Ethereum", "Bluzelle", "BTC", "ETH", "BLZ", "NO_ICO", "POST_ICO", "Currency",
		"Platform", "Storage", "2009-01-03", "2018-01-18"}
	expectedLen := len(expectedArgs)

	expectedSQLString := "SELECT * FROM crypto_asset ca LEFT JOIN team_member ON id = cryptoAssetId WHERE " +
		"(ca.name = ? OR ca.name = ? OR ca.name = ?) AND (symbol = ? OR symbol = ? OR symbol = ?) AND " +
		"(fundingStatus = ? OR fundingStatus = ?) AND (coinType = ? OR coinType = ? OR coinType = ?) AND " +
		"foundedDate >= ? AND foundedDate <= ?;"

	stmt := _createSelectStatement(expectedArgs[0:3], expectedArgs[3:6], expectedArgs[6:8], expectedArgs[8:11],
		expectedArgs[11], expectedArgs[12])

	if stmt.sql != expectedSQLString {
		t.Fatalf("unexpected SQL string\n\nexpected: %s\nactual: %s", expectedSQLString, stmt.sql)
	}

	argsLen := len(stmt.args)
	if argsLen != expectedLen {
		t.Fatalf("unexpected argument length\n\nexpected: %d\nactual: %d", expectedLen, argsLen)
	}

	for idx, expectedArg := range expectedArgs {
		actualArg, ok := stmt.args[idx].(string)
		if !ok {
			t.Fatalf("unexpected non-string argument at index %d", idx)
		}

		if actualArg != expectedArg {
			t.Fatalf("unexpected argument at index %d\n\nexpected: %s\nactual: %s", idx, expectedArg, actualArg)
		}
	}
}

func Test_createUpdateStatement(t *testing.T) {
	id := 1
	name := "Bitcoin"
	symbol := "BTC"
	description := "The original cryptocurrency"
	var icoAmount float64
	blockReward := 12.5
	fundingStatus := "NO-ICO"
	foundedDate := "2009-01-03"
	coinType := "Currency"
	website := "https://bitcoin.org/en/"
	cryptoAsset := &models.CryptoAsset{
		Name:          &name,
		Symbol:        &symbol,
		Description:   &description,
		ICOAmount:     &icoAmount,
		BlockReward:   &blockReward,
		FundingStatus: &fundingStatus,
		FoundedDate:   &foundedDate,
		CoinType:      &coinType,
		Website:       &website,
	}

	expectedSQLString := "UPDATE crypto_asset SET name = ?, symbol = ?, description = ?, icoAmount = ?, " +
		"blockReward = ?, fundingStatus = ?, foundedDate = ?, coinType = ?, website = ? WHERE id = ?;"

	expectedArgs := []interface{}{name, symbol, description, icoAmount, blockReward, fundingStatus, foundedDate, coinType,
		website, id}
	expectedLen := len(expectedArgs)

	stmt := _createUpdateStatement(id, cryptoAsset)

	if stmt.sql != expectedSQLString {
		t.Fatalf("unexpected SQL string\n\nexpected: %s\nactual: %s", expectedSQLString, stmt.sql)
	}

	argsLen := len(stmt.args)
	if argsLen != expectedLen {
		t.Fatalf("unexpected argument length\n\nexpected: %d\nactual: %d", expectedLen, argsLen)
	}

	for idx, expectedArg := range expectedArgs {
		if stmt.args[idx] != expectedArg {
			t.Fatalf("unexpected argument at index %d\n\nexpected: %v\nactual: %v", idx, expectedArg, stmt.args[idx])
		}
	}
}
