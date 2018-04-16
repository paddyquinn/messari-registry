package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paddyquinn/messari/database"
	"github.com/paddyquinn/messari/database/models"
	"github.com/paddyquinn/messari/util"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	// Miscellaneous constants.
	comma    = ","
	endpoint = "endpoint"
	errKey   = "error"

	// Error string constants.
	insertError         = "could not insert the crypto asset into the database"
	internalServerError = "internal server error"
	normalizeError      = "crypto asset normalization failed"
	nullTeamError       = "team cannot be null"
	parseError          = "unable to parse given crypto asset"
	selectError         = "error performing select query on the database"
	updateError         = "could not update the crypto asset"

	// Endpoint constants.
	registerEndpoint = "/register"
	searchEndpoint   = "/search"
	updateEndpoint   = "/update"
)

// Server is the main struct that responds to HTTP requests with responses from the database.
type Server struct {
	DB database.Interface
}

// NewServer creates a new server with the given database driver.
func NewServer(db database.Interface) *Server {
	return &Server{DB: db}
}

// Start runs the server. This function will loop infinitely if no error occurs.
func (s *Server) Start() error {
	// Initialize router.
	router := s.initializeRouter()

	// Run the router.
	if err := router.Run(); err != nil {
		return err
	}

	// Unreachable because router.Run() either errors out or loops forever.
	return nil
}

// initializeRouter registers the endpoints to route to the correct methods.
func (s *Server) initializeRouter() *gin.Engine {
	router := gin.Default()
	router.POST(registerEndpoint, s.register)
	router.GET(searchEndpoint, s.search)
	router.POST(updateEndpoint, s.update)
	return router
}

// register creates an entry in the database for the given crypto asset.
func (s *Server) register(ctx *gin.Context) {
	// Initialize the logger.
	logger := log.WithField(endpoint, registerEndpoint)

	// Parse the crypto asset passed in via the POST request.
	cryptoAsset, err := models.NewCryptoAsset(ctx.Request.Body)
	if err != nil {
		errString := err.Error()
		logger.WithField(errKey, errString).Error(parseError)
		ctx.JSON(http.StatusBadRequest, map[string]string{errKey: errString})
		return
	}

	// Ensure that a list of team members was passed in via the POST request. It is important to note that a null team is
	// different from an empty team. The register endpoint will reject JSON with no "team" key but will accept JSON of the
	// form {"team": []}. This allows for empty teams but forces the user to explicitly intend to pass in an empty team.
	if cryptoAsset.Team == nil {
		logger.Error(nullTeamError)
		ctx.JSON(http.StatusBadRequest, map[string]string{errKey: nullTeamError})
		return
	}

	// Normalize all of the fields in the newly created crypto asset struct. This will only fail if the date is not in
	// ISO-8601 format, the ICO amount or block reward are less than 0, or the id cannot be converted to an int, which is
	// irrelevant in this context.
	_, err = cryptoAsset.Normalize()
	if err != nil {
		errString := err.Error()
		logger.WithField(errKey, errString).Error(normalizeError)
		ctx.JSON(http.StatusBadRequest, map[string]string{errKey: errString})
		return
	}

	// Insert the crypto asset into the database. Note that the actual error string is only exposed to the user if a
	// user error that occurred. An internal database error is hidden behind a generic error message.
	id, err := s.DB.Insert(cryptoAsset)
	if err != nil {
		errString := err.Error()
		log.WithField(errKey, errString).Error(insertError)
		switch err.(type) {
		case *database.NullConstraintError, *database.UniqueConstraintError:
			ctx.JSON(http.StatusBadRequest, map[string]string{errKey: errString})
		default:
			ctx.JSON(http.StatusInternalServerError, map[string]string{errKey: internalServerError})
		}
		return
	}

	// Return the id back to the user.
	ctx.JSON(http.StatusOK, map[string]string{"id": id})
}

// search performs a search for crypto assets given the parameters passed in via the query string.
func (s *Server) search(ctx *gin.Context) {
	// Initialize the logger.
	logger := log.WithField(endpoint, searchEndpoint)

	// Get the crypto assets from the database.
	cryptoAssets, err := s.DB.Select(parseQueryString(ctx))
	if err != nil {
		logger.WithField(errKey, err.Error()).Error(selectError)
		ctx.JSON(http.StatusInternalServerError, map[string]string{errKey: internalServerError})
		return
	}

	// Format each crypto asset and return them all back to the user.
	for _, cryptoAsset := range cryptoAssets {
		cryptoAsset.Format()
	}
	ctx.JSON(http.StatusOK, cryptoAssets)
}

// update performs an update on a crypto asset given its id and the fields to update.
func (s *Server) update(ctx *gin.Context) {
	// Initialize the logger.
	logger := log.WithField(endpoint, updateEndpoint)

	// Parse the crypto asset passed in via the POST request.
	cryptoAsset, err := models.NewCryptoAsset(ctx.Request.Body)
	if err != nil {
		logger.WithField(errKey, err.Error()).Error(parseError)
		ctx.JSON(http.StatusBadRequest, false)
		return
	}

	// Ensure that the passed crypto asset has an id.
	if cryptoAsset.ID == nil {
		logger.Error("no id in POST request")
		ctx.JSON(http.StatusBadRequest, false)
		return
	}

	// Normalize all of the fields in the newly created crypto asset struct. This will only fail if the date is not in
	// ISO-8601 format or the id cannot be converted to an int.
	id, err := cryptoAsset.Normalize()
	if err != nil {
		logger.WithField(errKey, err.Error()).Error(normalizeError)
		ctx.JSON(http.StatusBadRequest, false)
		return
	}
	logger.WithField("id", id)

	// Update the crypto asset with the given id in the database. If an empty update or a
	err = s.DB.Update(id, cryptoAsset)
	if err != nil {
		logger.WithField(errKey, err.Error()).Error(updateError)
		switch err.(type) {
		case *database.EmptyUpdateError, *database.UnknownIDError:
			ctx.JSON(http.StatusBadRequest, false)
		default:
			ctx.JSON(http.StatusInternalServerError, false)
		}
		return
	}

	// Return a true boolean to the user.
	ctx.JSON(http.StatusOK, true)
}

// parseQueryString extracts the "name", "symbol", "fundingStatus", "coinType", "startDate", and "endDate" from the
// query string. Note that "startDate" and "endDate" will always take the first comma separated value while the rest
// will be an array and can have multiple values.
func parseQueryString(ctx *gin.Context) ([]string, []string, []string, []string, string, string) {
	return splitQueryArray(ctx.QueryArray("name")), splitQueryArray(ctx.QueryArray("symbol")),
		splitQueryArray(ctx.QueryArray("fundingStatus")), splitQueryArray(ctx.QueryArray("coinType")),
		parseDate(ctx.Query("startDate")), parseDate(ctx.Query("endDate"))
}

// parseDate takes the first date value and ensures it is in ISO-8601 format. Otherwise, an empty string is returned.
func parseDate(query string) string {
	date := strings.Split(query, comma)[0]
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return ""
	}
	return date
}

// splitQueryArray splits a query string parameter into its comma separated values.
func splitQueryArray(queryArray []string) []string {
	var commaSeparatedArray []string
	for _, queryValue := range queryArray {
		commaSeparatedValues := strings.Split(queryValue, comma)
		for _, commaSeparatedValue := range commaSeparatedValues {
			commaSeparatedArray = append(commaSeparatedArray, *util.Normalize(commaSeparatedValue))
		}
	}
	return commaSeparatedArray
}
