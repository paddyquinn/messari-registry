package models

import "testing"

func TestCryptoAsset_Format(t *testing.T) {
	name := "bitcoin"
	symbol := "btc"
	fundingStatus := "no-ico"
	coinType := "currency"
	cryptoAsset := &CryptoAsset{
		Name:          &name,
		Symbol:        &symbol,
		FundingStatus: &fundingStatus,
		CoinType:      &coinType,
	}

	cryptoAsset.Format()
	assertEquals(t, "name", "Bitcoin", *cryptoAsset.Name)
	assertEquals(t, "symbol", "BTC", *cryptoAsset.Symbol)
	assertEquals(t, "fundingStatus", "NO-ICO", *cryptoAsset.FundingStatus)
	assertEquals(t, "coinType", "Currency", *cryptoAsset.CoinType)
}

func TestCryptoAsset_Normalize(t *testing.T) {
	testNonNumericID(t)
	testNegativeICOAmount(t)
	testNegativeBlockReward(t)
	testNonISO8601Date(t)
	testFutureDate(t)
	testSuccess(t)
}

func testNonNumericID(t *testing.T) {
	idString := "a"
	cryptoAsset := &CryptoAsset{ID: &idString}
	id, err := cryptoAsset.Normalize()
	assertEquals(t, "id", -1, id)
	assertEquals(t, "error", "invalid id: a", err.Error())
}

func testNegativeICOAmount(t *testing.T) {
	icoAmount := -1.5
	cryptoAsset := &CryptoAsset{ICOAmount: &icoAmount}
	id, err := cryptoAsset.Normalize()
	assertEquals(t, "id", -1, id)
	assertEquals(t, "error", "ICO amount cannot be negative", err.Error())
}

func testNegativeBlockReward(t *testing.T) {
	blockReward := -1.5
	cryptoAsset := &CryptoAsset{BlockReward: &blockReward}
	id, err := cryptoAsset.Normalize()
	assertEquals(t, "id", -1, id)
	assertEquals(t, "error", "block reward cannot be negative", err.Error())
}

func testNonISO8601Date(t *testing.T) {
	date := "12/25/2017"
	cryptoAsset := &CryptoAsset{FoundedDate: &date}
	id, err := cryptoAsset.Normalize()
	assertEquals(t, "id", -1, id)
	assertEquals(t, "error", "date must be an ISO-8601 date in the past", err.Error())
}

func testFutureDate(t *testing.T) {
	date := "9999-12-31"
	cryptoAsset := &CryptoAsset{FoundedDate: &date}
	id, err := cryptoAsset.Normalize()
	assertEquals(t, "id", -1, id)
	assertEquals(t, "error", "date must be an ISO-8601 date in the past", err.Error())
}

func testSuccess(t *testing.T) {
	name := "   biTcoiN    "
	symbol := "   btc    "
	description := "   The original cryptocurrency   "
	fundingStatus := "no-ico"
	foundedDate := "   2009-01-03   "
	coinType := "   cuRRency "
	website := "   https://bitcoin.org/en/   "
	cryptoAsset := &CryptoAsset{
		Name:          &name,
		Symbol:        &symbol,
		Description:   &description,
		Team:          []string{"  Satoshi Nakomoto  ", "  Hal Finney  ", "  Nick Szabo  "},
		FundingStatus: &fundingStatus,
		FoundedDate:   &foundedDate,
		CoinType:      &coinType,
		Website:       &website,
	}

	id, err := cryptoAsset.Normalize()
	assertEquals(t, "id", 0, id)
	assertEquals(t, "error", err, nil)

	assertEquals(t, "name", "bitcoin", *cryptoAsset.Name)
	assertEquals(t, "symbol", "btc", *cryptoAsset.Symbol)
	assertEquals(t, "description", "The original cryptocurrency", *cryptoAsset.Description)
	assertTeamEquals(t, []string{"Satoshi Nakomoto", "Hal Finney", "Nick Szabo"}, cryptoAsset.Team)
	assertEquals(t, "fundingStatus", "no-ico", *cryptoAsset.FundingStatus)
	assertEquals(t, "foundedDate", "2009-01-03", *cryptoAsset.FoundedDate)
	assertEquals(t, "coinType", "currency", *cryptoAsset.CoinType)
	assertEquals(t, "website", "https://bitcoin.org/en/", *cryptoAsset.Website)
}

func assertEquals(t *testing.T, field string, expected, actual interface{}) {
	if expected != actual {
		t.Fatalf("%s comparison failed\n\nexpected: %v\nactual: %v", field, expected, actual)
	}
}

func assertTeamEquals(t *testing.T, expectedTeam, actualTeam []string) {
	if len(expectedTeam) != len(actualTeam) {
		t.Fatal("team length changed unexpectedly")
	}

	for idx, expectedMember := range expectedTeam {
		if expectedMember != actualTeam[idx] {
			t.Fatalf("team member comparison at index %d failed\n\nexpected: %s\nactual: %s", idx, expectedMember,
				actualTeam[idx])
		}

	}
}
