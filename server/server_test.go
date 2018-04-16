package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paddyquinn/messari/database"
	"github.com/paddyquinn/messari/database/models"
	log "github.com/sirupsen/logrus"
)

func TestRegisterEndpoint(t *testing.T) {
	// Hide logs.
	log.SetLevel(log.FatalLevel)

	// Set up router for testing.
	gin.SetMode(gin.TestMode)
	mockDatabase := &database.Mock{}
	mockRouter := setUpMockRouter(mockDatabase)

	// Run the tests.
	testInvalidJSON(t, mockRouter, registerEndpoint, "{\"error\":\"unexpected EOF\"}")
	testNullTeam(t, mockRouter, registerEndpoint, "{\"error\":\"team cannot be null\"}")
	testNormalizationError(t, mockRouter, registerEndpoint, "{\"error\":\"date must be an ISO-8601 date in the past\"}")
	testRegisterUserError(t, mockRouter, mockDatabase)
	testRegisterDatabaseError(t, mockRouter, mockDatabase)
	testRegisterSuccess(t, mockRouter, mockDatabase)
}

func testNullTeam(t *testing.T, mockRouter *gin.Engine, endpoint, expectedResponse string) {
	testNull(t, mockRouter, endpoint, expectedResponse)
}

func testRegisterUserError(t *testing.T, mockRouter *gin.Engine, mockDatabase *database.Mock) {
	// Create the response recorder.
	recorder := httptest.NewRecorder()

	// Create a crypto asset with null data and marshal it into bytes to be sent to our mock router.
	date := "2017-12-25"
	invalidCryptoAsset := &models.CryptoAsset{Team: []string{}, FoundedDate: &date}
	buffer, err := json.Marshal(invalidCryptoAsset)

	// Fail the test if there is an error marshalling the valid crypto asset.
	if err != nil {
		t.Fatal("unexpected error marshaling to JSON")
	}
	// Prepare the HTTP request and mock database call.
	req := httptest.NewRequest("POST", registerEndpoint, bytes.NewReader(buffer))
	mockDatabase.On("Insert", invalidCryptoAsset).Return("", database.NewNullConstraintError("name"))

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the correct mock calls were made.
	mockDatabase.AssertExpectations(t)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusBadRequest, recorder.Code)
	assertResponseBody(t, "{\"error\":\"name cannot be null\"}", recorder.Body.String())
}

func testRegisterDatabaseError(t *testing.T, mockRouter *gin.Engine, mockDatabase *database.Mock) {
	// Create the response recorder.
	recorder := httptest.NewRecorder()

	// Create a valid crypto asset and marshal it into bytes to be sent to our mock router.
	name := "bitcoin"
	symbol := "btc"
	description := "The original cryptocurrency"
	var icoAmount float64
	blockReward := 12.5
	fundingStatus := "no-ico"
	foundedDate := "2009-01-03"
	coinType := "currency"
	website := "https://bitcoin.org/en/"
	validCryptoAsset := &models.CryptoAsset{
		Name:          &name,
		Symbol:        &symbol,
		Description:   &description,
		Team:          []string{"Satoshi Nakomoto", "Hal Finney", "Nick Szabo"},
		ICOAmount:     &icoAmount,
		BlockReward:   &blockReward,
		FundingStatus: &fundingStatus,
		FoundedDate:   &foundedDate,
		CoinType:      &coinType,
		Website:       &website,
	}
	buffer, err := json.Marshal(validCryptoAsset)

	// Fail the test if there is an error marshalling the valid crypto asset.
	if err != nil {
		t.Fatal("unexpected error marshaling to JSON")
	}
	// Prepare the HTTP request and mock database call.
	req := httptest.NewRequest("POST", registerEndpoint, bytes.NewReader(buffer))
	mockDatabase.On("Insert", validCryptoAsset).Return("", errors.New("mock database error"))

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the correct mock calls were made.
	mockDatabase.AssertExpectations(t)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusInternalServerError, recorder.Code)
	assertResponseBody(t, "{\"error\":\"internal server error\"}", recorder.Body.String())
}

func testRegisterSuccess(t *testing.T, mockRouter *gin.Engine, mockDatabase *database.Mock) {
	// Create the response recorder.
	recorder := httptest.NewRecorder()

	// Create a valid crypto asset and marshal it into bytes to be sent to our mock router.
	name := "aragon"
	symbol := "ant"
	description := "A decentralized autonomous organization"
	var icoAmount float64 = 27709391
	var blockReward float64
	fundingStatus := "post-ico"
	foundingDate := "2017-05-17"
	coinType := "governance"
	website := "https://aragon.one/"
	validCryptoAsset := &models.CryptoAsset{
		Name:          &name,
		Symbol:        &symbol,
		Description:   &description,
		Team:          []string{"Luis Cuende", "Jorge Izquierdo"},
		ICOAmount:     &icoAmount,
		BlockReward:   &blockReward,
		FundingStatus: &fundingStatus,
		FoundedDate:   &foundingDate,
		CoinType:      &coinType,
		Website:       &website,
	}
	buffer, err := json.Marshal(validCryptoAsset)

	// Fail the test if there is an error marshalling the valid crypto asset.
	if err != nil {
		t.Fatal("unexpected error marshaling to JSON")
	}

	// Prepare the HTTP request and mock database call.
	reader := bytes.NewReader(buffer)
	req := httptest.NewRequest("POST", registerEndpoint, reader)
	mockDatabase.On("Insert", validCryptoAsset).Return("1", nil)

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the correct mock calls were made.
	mockDatabase.AssertExpectations(t)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusOK, recorder.Code)
	assertResponseBody(t, "{\"id\":\"1\"}", recorder.Body.String())
}

func TestSearchEndpoint(t *testing.T) {
	// Hide logs.
	log.SetLevel(log.FatalLevel)

	// Set up router for testing.
	gin.SetMode(gin.TestMode)
	mockDatabase := &database.Mock{}
	mockRouter := setUpMockRouter(mockDatabase)

	// Run tests.
	testSearchDatabaseError(t, mockRouter, mockDatabase)
	testSearchSuccess(t, mockRouter, mockDatabase)
}

func testSearchDatabaseError(t *testing.T, mockRouter *gin.Engine, mockDatabase *database.Mock) {
	// Create the response recorder.
	recorder := httptest.NewRecorder()

	// Prepare the HTTP request and mock database call.
	req := httptest.NewRequest("GET", searchEndpoint, nil)
	mockDatabase.On("Select", []string(nil), []string(nil), []string(nil), []string(nil), "", "").Return(nil,
		errors.New("mock database error"))

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the correct mock calls were made.
	mockDatabase.AssertExpectations(t)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusInternalServerError, recorder.Code)
	assertResponseBody(t, "{\"error\":\"internal server error\"}", recorder.Body.String())
}

func testSearchSuccess(t *testing.T, mockRouter *gin.Engine, mockDatabase *database.Mock) {
	// Create the response recorder
	recorder := httptest.NewRecorder()

	// Create a crypto asset to be returned.
	aragonID := "2"
	aragonName := "aragon"
	aragonSymbol := "ant"
	aragonDescription := "A decentralized autonomous organization"
	var aragonICOAmount float64 = 27709391
	var aragonBlockReward float64
	aragonFundingStatus := "post-ico"
	aragonFoundingDate := "2017-05-17"
	aragonCoinType := "governance"
	aragonWebsite := "https://aragon.one/"
	aragon := &models.CryptoAsset{
		ID:            &aragonID,
		Name:          &aragonName,
		Symbol:        &aragonSymbol,
		Description:   &aragonDescription,
		Team:          []string{"Luis Cuende", "Jorge Izquierdo"},
		ICOAmount:     &aragonICOAmount,
		BlockReward:   &aragonBlockReward,
		FundingStatus: &aragonFundingStatus,
		FoundedDate:   &aragonFoundingDate,
		CoinType:      &aragonCoinType,
		Website:       &aragonWebsite,
	}

	// Create a second crypto asset to be returned.
	storjID := "3"
	storjName := "storj"
	storjSymbol := "storj"
	storjDescription := "Decentralized storage"
	var storjICOAmount float64 = 125000000
	var storjBlockReward float64
	storjFundingStatus := "post-ico"
	storjFoundedDate := "2017-05-19"
	storjCoinType := "storage"
	storjWebsite := "https://storj.io/"
	storj := &models.CryptoAsset{
		ID:            &storjID,
		Name:          &storjName,
		Symbol:        &storjSymbol,
		Description:   &storjDescription,
		Team:          []string{"Ben Golub", "Shawn Wilkinson", "John Quinn", "Philip Hutchins", "Matthew May"},
		ICOAmount:     &storjICOAmount,
		BlockReward:   &storjBlockReward,
		FundingStatus: &storjFundingStatus,
		FoundedDate:   &storjFoundedDate,
		CoinType:      &storjCoinType,
		Website:       &storjWebsite,
	}

	// Prepare the HTTP request and mock database call. Note that the mock call expects the query string to be parsed such
	// that query parameters can be split by commas or ampersands but single value query parameters (startDate, endDate)
	// take only the first value.
	req := httptest.NewRequest("GET", "/search?fundingStatus=post-ico&fundingStatus=active-ico"+
		"&coinType=governance,storage&startDate=2017-05-17,2017-05-18&endDate=2017-05-19&endDate=2017-05-18", nil)
	mockDatabase.On("Select", []string(nil), []string(nil), []string{"post-ico", "active-ico"},
		[]string{"governance", "storage"}, "2017-05-17", "2017-05-19").Return([]*models.CryptoAsset{aragon, storj}, nil)

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the correct mock calls were made.
	mockDatabase.AssertExpectations(t)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusOK, recorder.Code)
	assertResponseBody(t, "[{\"id\":\"2\",\"name\":\"Aragon\",\"symbol\":\"ANT\","+
		"\"description\":\"A decentralized autonomous organization\",\"team\":[\"Luis Cuende\",\"Jorge Izquierdo\"],"+
		"\"icoAmount\":27709391,\"blockReward\":0,\"fundingStatus\":\"POST-ICO\",\"foundedDate\":\"2017-05-17\","+
		"\"coinType\":\"Governance\",\"website\":\"https://aragon.one/\"},{\"id\":\"3\",\"name\":\"Storj\","+
		"\"symbol\":\"STORJ\",\"description\":\"Decentralized storage\",\"team\":[\"Ben Golub\",\"Shawn Wilkinson\","+
		"\"John Quinn\",\"Philip Hutchins\",\"Matthew May\"],\"icoAmount\":125000000,\"blockReward\":0,"+
		"\"fundingStatus\":\"POST-ICO\",\"foundedDate\":\"2017-05-19\",\"coinType\":\"Storage\","+
		"\"website\":\"https://storj.io/\"}]", recorder.Body.String())
}

func TestUpdateEndpoint(t *testing.T) {
	// Hide logs.
	log.SetLevel(log.FatalLevel)

	// Set up router for testing.
	gin.SetMode(gin.TestMode)
	mockDatabase := &database.Mock{}
	mockRouter := setUpMockRouter(mockDatabase)

	// Run tests.
	testInvalidJSON(t, mockRouter, updateEndpoint, "false")
	testNullID(t, mockRouter, updateEndpoint, "false")
	testNormalizationError(t, mockRouter, updateEndpoint, "false")
	testUpdateUserUpdate(t, mockRouter, mockDatabase)
	testUpdateDatabaseError(t, mockRouter, mockDatabase)
	testUpdateSuccess(t, mockRouter, mockDatabase)
}

func testNullID(t *testing.T, mockRouter *gin.Engine, endpoint, expectedResponse string) {
	testNull(t, mockRouter, endpoint, expectedResponse)
}

func testUpdateUserUpdate(t *testing.T, mockRouter *gin.Engine, mockDatabase *database.Mock) {
	recorder := httptest.NewRecorder()

	// Create an empty crypto asset update and marshal it into bytes to be sent to our mock router.
	id := "1"
	emptyCryptoAssetUpdate := &models.CryptoAsset{
		ID: &id,
	}
	buffer, err := json.Marshal(emptyCryptoAssetUpdate)

	// Fail the test if there is an error marshalling the valid crypto asset.
	if err != nil {
		t.Fatal("unexpected error marshaling to JSON")
	}

	// Prepare the HTTP request and mock database call.
	reader := bytes.NewReader(buffer)
	req := httptest.NewRequest("POST", "/update", reader)
	mockDatabase.On("Update", 1, emptyCryptoAssetUpdate).Return(database.NewEmptyUpdateError())

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the correct mock calls were made.
	mockDatabase.AssertExpectations(t)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusBadRequest, recorder.Code)
	assertResponseBody(t, "false", recorder.Body.String())
}

func testUpdateDatabaseError(t *testing.T, mockRouter *gin.Engine, mockDatabase *database.Mock) {
	// Create the response recorder.
	recorder := httptest.NewRecorder()

	// Create a valid crypto asset update and marshal it into bytes to be sent to our mock router.
	id := "3"
	symbol := "stj"
	validCryptoAssetUpdate := &models.CryptoAsset{
		ID:     &id,
		Symbol: &symbol,
	}
	buffer, err := json.Marshal(validCryptoAssetUpdate)

	// Fail the test if there is an error marshalling the valid crypto asset.
	if err != nil {
		t.Fatal("unexpected error marshaling to JSON")
	}
	// Prepare the HTTP request and mock database call.
	req := httptest.NewRequest("POST", updateEndpoint, bytes.NewReader(buffer))
	mockDatabase.On("Update", 3, validCryptoAssetUpdate).Return(errors.New("mock database error"))

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the correct mock calls were made.
	mockDatabase.AssertExpectations(t)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusInternalServerError, recorder.Code)
	assertResponseBody(t, "false", recorder.Body.String())
}

func testUpdateSuccess(t *testing.T, mockRouter *gin.Engine, mockDatabase *database.Mock) {
	recorder := httptest.NewRecorder()

	// Create a valid crypto asset update and marshal it into bytes to be sent to our mock router.
	id := "1"
	blockReward := 6.25
	validCryptoAssetUpdate := &models.CryptoAsset{
		ID:          &id,
		BlockReward: &blockReward,
	}
	buffer, err := json.Marshal(validCryptoAssetUpdate)

	// Fail the test if there is an error marshalling the valid crypto asset.
	if err != nil {
		t.Fatal("unexpected error marshaling to JSON")
	}

	// Prepare the HTTP request and mock database call.
	reader := bytes.NewReader(buffer)
	req := httptest.NewRequest("POST", "/update", reader)
	mockDatabase.On("Update", 1, validCryptoAssetUpdate).Return(nil)

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the correct mock calls were made.
	mockDatabase.AssertExpectations(t)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusOK, recorder.Code)
	assertResponseBody(t, "true", recorder.Body.String())
}

func assertResponseBody(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("unexpected response body\n\nexpected: %s\nactual: %s", expected, actual)
	}
}

func assertResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Fatalf("unexpected response code\n\nexpected: %d\nactual: %d", expected, actual)
	}
}

func setUpMockRouter(mock *database.Mock) *gin.Engine {
	server := NewServer(mock)
	return server.initializeRouter()
}

func testInvalidJSON(t *testing.T, mockRouter *gin.Engine, endpoint, expectedResponse string) {
	// Create the response recorder.
	recorder := httptest.NewRecorder()

	// Prepare the HTTP request.
	req := httptest.NewRequest("POST", endpoint, strings.NewReader("{"))

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusBadRequest, recorder.Code)
	assertResponseBody(t, expectedResponse, recorder.Body.String())
}

func testNormalizationError(t *testing.T, mockRouter *gin.Engine, endpoint, expectedResponse string) {
	// Create the response recorder.
	recorder := httptest.NewRecorder()

	// Create a crypto asset with an invalid date and marshal it into bytes to be sent to our mock router.
	id := "4"
	nonISO8601Date := "12/25/2017"
	invalidCryptoAsset := &models.CryptoAsset{ID: &id, Team: []string{}, FoundedDate: &nonISO8601Date}
	buffer, err := json.Marshal(invalidCryptoAsset)

	// Fail the test if there is an error marshalling the valid crypto asset.
	if err != nil {
		t.Fatal("unexpected error marshaling to JSON")
	}
	// Prepare the HTTP request.
	req := httptest.NewRequest("POST", endpoint, bytes.NewReader(buffer))

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusBadRequest, recorder.Code)
	assertResponseBody(t, expectedResponse, recorder.Body.String())
}

func testNull(t *testing.T, mockRouter *gin.Engine, endpoint, expectedResponse string) {
	// Create the response recorder.
	recorder := httptest.NewRecorder()

	// Prepare the HTTP request.
	req := httptest.NewRequest("POST", endpoint, strings.NewReader("{}"))

	// Make the request.
	mockRouter.ServeHTTP(recorder, req)

	// Assert the expected HTTP response code and body.
	assertResponseCode(t, http.StatusBadRequest, recorder.Code)
	assertResponseBody(t, expectedResponse, recorder.Body.String())
}
