package database

import (
	"bytes"
	"database/sql"
	//"errors"
	"fmt"
	"os"
	"strconv"
	//"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/paddyquinn/messari/database/models"
	"strings"
)

const (
	dbFile      = "database/data/sqlite"
	emptyString = ""
)

// SQLite is an implementation of the database interface to connect to a SQLite database.
type SQLite struct {
	connection *sql.DB
}

// NewSQLite creates a new SQLite database connection. If the SQLite file does not exist, it is created and the
// crypto_asset and team_member tables are created.
func NewSQLite() (*SQLite, error) {
	// Open a database connection to a sqlite db file with foreign keys enabled. Note: not all sqlite binaries support
	// foreign keys.
	conn, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=1", dbFile))
	if err != nil {
		return nil, err
	}

	// Create tables if the database file did not previously exist.
	if _, err = os.Stat(dbFile); os.IsNotExist(err) {
		_, err = conn.Exec("CREATE TABLE crypto_asset(id INTEGER PRIMARY KEY, name TEXT NOT NULL, " +
			"symbol TEXT UNIQUE NOT NULL, description TEXT NOT NULL, icoAmount REAL NOT NULL, blockReward REAL NOT NULL, " +
			"fundingStatus TEXT NOT NULL, foundedDate TEXT NOT NULL, coinType TEXT NOT NULL, website TEXT NOT NULL);")
		if err != nil {
			// TODO: what if this errors?
			os.Remove(dbFile)
			return nil, err
		}

		_, err = conn.Exec("CREATE TABLE team_member(cryptoAssetId INTEGER NOT NULL, name TEXT NOT NULL, " +
			"FOREIGN KEY(cryptoAssetId) REFERENCES crypto_asset(id));")
		if err != nil {
			// TODO: what if this errors?
			os.Remove(dbFile)
			return nil, err
		}
	}

	return &SQLite{connection: conn}, nil
}

// Insert inserts the crypto asset into the crypto_asset table and its team members into the team_member table.
func (s *SQLite) Insert(cryptoAsset *models.CryptoAsset) (string, error) {
	// Begin a SQL transaction to guarantee all inserts are executed or a rollback occurs.
	transaction, err := s.connection.Begin()
	if err != nil {
		return emptyString, err
	}

	// Insert the crypto asset into the database.
	result, err := transaction.Exec("INSERT INTO crypto_asset(name, symbol, description, icoAmount, blockReward, "+
		"fundingStatus, foundedDate, coinType, website) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)",
		cryptoAsset.Name, cryptoAsset.Symbol, cryptoAsset.Description, cryptoAsset.ICOAmount, cryptoAsset.BlockReward,
		cryptoAsset.FundingStatus, cryptoAsset.FoundedDate, cryptoAsset.CoinType, cryptoAsset.Website)
	if err != nil {
		transaction.Rollback()

		// If the error is a constraint error return an error type that we've created and the server package knows how to
		// handle. Otherwise, just return the SQLite driver error. If the error returned from the SQLite driver is a null
		// constraint error it will be of the form 'NOT NULL constraint failed: crypto_asset.name' so the offending field is
		// derived by taking all of the characters after the last appearance of a period.
		sqliteErr := err.(sqlite3.Error)
		switch sqliteErr.ExtendedCode {
		case sqlite3.ErrConstraintNotNull:
			errString := sqliteErr.Error()
			nullField := errString[strings.LastIndex(errString, ".")+1:]
			return emptyString, NewNullConstraintError(nullField)
		case sqlite3.ErrConstraintUnique:
			return emptyString, NewUniqueConstraintError(*cryptoAsset.Symbol)
		}

		return emptyString, sqliteErr
	}

	id64, err := result.LastInsertId()
	if err != nil {
		transaction.Rollback()
		return emptyString, err
	}
	id := int(id64)

	// Insert team members into the team_member table.
	err = insertTeamMembers(transaction, id, cryptoAsset.Team)
	if err != nil {
		transaction.Rollback()
		return emptyString, err
	}

	// Commit the transaction.
	err = transaction.Commit()
	if err != nil {
		return emptyString, err
	}

	// Return the new id of the inserted crypto asset.
	return strconv.Itoa(id), nil
}

// Select searches for crypto assets given the passed parameters.
func (s *SQLite) Select(names, symbols, fundingStatuses, coinTypes []string, startDate,
	endDate string) ([]*models.CryptoAsset, error) {

	// Create and execute the select statement.
	stmt := _createSelectStatement(names, symbols, fundingStatuses, coinTypes, startDate, endDate)
	rows, err := s.connection.Query(stmt.sql, stmt.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Instantiate a map from id to crypto asset and then loop through each row. The LEFT JOIN causes there to be a one
	// row per team member for each asset so the first time we add the crypto asset to the map and then each subsequent
	// time we append the team member to the list of team member's for that already created asset.
	cryptoAssetMap := make(map[int]*models.CryptoAsset)
	for rows.Next() {
		// Retrieve the values from the current row.
		var (
			id                                                                       int
			name, symbol, description, fundingStatus, foundedDate, coinType, website string
			icoAmount, blockReward                                                   float64
			fkID                                                                     *int
			teamMember                                                               *string
		)
		err := rows.Scan(&id, &name, &symbol, &description, &icoAmount, &blockReward, &fundingStatus, &foundedDate,
			&coinType, &website, &fkID, &teamMember)
		if err != nil {
			return nil, err
		}

		// Find the given crypto asset by id. If it cannot be found, create it and put it in the map. Then add the team
		// member from that row to the asset's list of team members given the team member is not null.
		cryptoAsset, found := cryptoAssetMap[id]
		if !found {
			idString := strconv.Itoa(id)
			cryptoAsset = &models.CryptoAsset{
				ID:            &idString,
				Name:          &name,
				Symbol:        &symbol,
				Description:   &description,
				Team:          []string{},
				ICOAmount:     &icoAmount,
				BlockReward:   &blockReward,
				FundingStatus: &fundingStatus,
				FoundedDate:   &foundedDate,
				CoinType:      &coinType,
				Website:       &website,
			}
			cryptoAssetMap[id] = cryptoAsset
		}
		if teamMember != nil {
			cryptoAsset.Team = append(cryptoAsset.Team, *teamMember)
		}
	}

	// Translate the crypto asset map to an array.
	cryptoAssets := make([]*models.CryptoAsset, len(cryptoAssetMap))
	idx := 0
	for _, cryptoAsset := range cryptoAssetMap {
		cryptoAssets[idx] = cryptoAsset
		idx++
	}

	// Return the array of crypto assets.
	return cryptoAssets, nil
}

// Update updates a crypto asset with the fields it contains. If the passed crypto asset has a team array then all of
// the old team members are deleted from the team_member table and all of the new members are inserted.
func (s *SQLite) Update(id int, cryptoAsset *models.CryptoAsset) error {
	// Create the update statement. If there is nothing to update given the passed asset, return an empty update error.
	updateCryptoAssetStatement := _createUpdateStatement(id, cryptoAsset)
	if updateCryptoAssetStatement == nil && cryptoAsset.Team == nil {
		return NewEmptyUpdateError()
	}

	// Begin a SQL transaction to guarantee all updates and inserts are executed or a rollback occurs.
	transaction, err := s.connection.Begin()
	if err != nil {
		return err
	}

	if updateCryptoAssetStatement != nil {
		// Execute the update statement.
		result, err := transaction.Exec(updateCryptoAssetStatement.sql, updateCryptoAssetStatement.args...)
		if err != nil {
			transaction.Rollback()
			return err
		}

		// Get the number of rows affected. If the number of rows affected is not 1, the only possibility is that it is 0,
		// indicating that no rows were affected because there were no assets with the id passed. Note that this is not
		// returned as an error from transaction.Exec(...). In this case, return an UnknownIDError.
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected != 1 {
			return NewUnknownIDError(id)
		}
	}

	if cryptoAsset.Team != nil {
		// Do not check the rows affected here because it is possible an asset has no team members.
		_, err = transaction.Exec("DELETE FROM team_member WHERE cryptoAssetId = ?;", id)
		if err != nil {
			transaction.Rollback()
			return err
		}

		// Insert the team members into the team_member table.
		err = insertTeamMembers(transaction, id, cryptoAsset.Team)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}

	// Commit the transaction and return.
	err = transaction.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Close closes the connection to the SQLite database.
func (s *SQLite) Close() {
	s.connection.Close()
}

// statement represents a SQL statement and its arguments.
type statement struct {
	sql  string
	args []interface{}
}

// _createSelectStatement creates a select statement that joins the crypto_asset and team_member tables and is used to
// search the database. Note that this function starts with an underscore because it is idiomatic in Go to write test
// functions as TestFunctionName. However, since this is a private function. TestcreateSelectStatement looked awkward
// so an underscore was introduced to make the function name clear.
func _createSelectStatement(names, symbols, fundingStatuses, coinTypes []string, startDate, endDate string) *statement {
	var args []interface{}
	isFirstClause := true
	sqlBuffer := bytes.NewBufferString("SELECT * FROM crypto_asset ca LEFT JOIN team_member ON id = cryptoAssetId")

	if ok := createConditionalClause(sqlBuffer, &isFirstClause, len(names), "ca.name"); ok {
		for _, name := range names {
			args = append(args, name)
		}
	}

	if ok := createConditionalClause(sqlBuffer, &isFirstClause, len(symbols), "symbol"); ok {
		for _, symbol := range symbols {
			args = append(args, symbol)
		}
	}

	if ok := createConditionalClause(sqlBuffer, &isFirstClause, len(fundingStatuses), "fundingStatus"); ok {
		for _, fundingStatus := range fundingStatuses {
			args = append(args, fundingStatus)
		}
	}

	if ok := createConditionalClause(sqlBuffer, &isFirstClause, len(coinTypes), "coinType"); ok {
		for _, coinType := range coinTypes {
			args = append(args, coinType)
		}
	}

	if ok := createDateClause(sqlBuffer, &isFirstClause, startDate, ">="); ok {
		args = append(args, startDate)
	}

	if ok := createDateClause(sqlBuffer, &isFirstClause, endDate, "<="); ok {
		args = append(args, endDate)
	}

	sqlBuffer.WriteRune(';')

	return &statement{sql: sqlBuffer.String(), args: args}
}

// createClauseKeyword determines whether this is the first conditional clause in the select statement and writes
// 'WHERE' if so and 'AND' if not.
func createClauseKeyword(sqlBuffer *bytes.Buffer, isFirstClause *bool) {
	if *isFirstClause {
		sqlBuffer.WriteString(" WHERE")
		*isFirstClause = false
	} else {
		sqlBuffer.WriteString(" AND")
	}
}

// createConditionalClause writes a conditional clause to the SQL statement if there are values to compare to and
// returns true, and returns false otherwise.
func createConditionalClause(sqlBuffer *bytes.Buffer, isFirstClause *bool, numValues int, column string) bool {
	if numValues > 0 {
		createClauseKeyword(sqlBuffer, isFirstClause)
		sqlBuffer.WriteString(" (")

		condition := fmt.Sprintf("%s = ?", column)
		for i := 0; i < numValues; i++ {
			if i != 0 {
				sqlBuffer.WriteString(" OR ")
			}
			sqlBuffer.WriteString(condition)
		}

		sqlBuffer.WriteRune(')')

		return true
	}

	return false
}

// createDateClause writes a conditional clause comparing dates to the SQL statement if there is a date passed with
// non-zero length and returns true, and returns false otherwise.
// TODO: will this comparator work??
func createDateClause(sqlBuffer *bytes.Buffer, isFirstClause *bool, date, comparator string) bool {
	if len(date) > 0 {
		createClauseKeyword(sqlBuffer, isFirstClause)
		sqlBuffer.WriteString(fmt.Sprintf(" foundedDate %s ?", comparator))

		return true
	}

	return false
}

// _createUpdateStatement creates an update statement that updates a crypto asset in the database by id. Note that this
// function starts with an underscore because it is idiomatic in Go to write test functions as TestFunctionName.
// However, since this is a private function. TestcreateUpdateStatement looked awkward so an underscore was introduced
// to make the function name clear.
func _createUpdateStatement(id int, cryptoAsset *models.CryptoAsset) *statement {
	var args []interface{}
	sqlBuffer := bytes.NewBufferString("UPDATE crypto_asset SET ")
	setClauseWritten := false

	if ok := createSetClause(sqlBuffer, &setClauseWritten, cryptoAsset.Name == nil, "name"); ok {
		args = append(args, *cryptoAsset.Name)
	}

	if ok := createSetClause(sqlBuffer, &setClauseWritten, cryptoAsset.Symbol == nil, "symbol"); ok {
		args = append(args, *cryptoAsset.Symbol)
	}

	if ok := createSetClause(sqlBuffer, &setClauseWritten, cryptoAsset.Description == nil,
		"description"); ok {

		args = append(args, *cryptoAsset.Description)
	}

	if ok := createSetClause(sqlBuffer, &setClauseWritten, cryptoAsset.ICOAmount == nil, "icoAmount"); ok {
		args = append(args, *cryptoAsset.ICOAmount)
	}

	if ok := createSetClause(sqlBuffer, &setClauseWritten, cryptoAsset.BlockReward == nil, "blockReward"); ok {
		args = append(args, *cryptoAsset.BlockReward)
	}

	if ok := createSetClause(sqlBuffer, &setClauseWritten, cryptoAsset.FundingStatus == nil, "fundingStatus"); ok {

		args = append(args, *cryptoAsset.FundingStatus)
	}

	if ok := createSetClause(sqlBuffer, &setClauseWritten, cryptoAsset.FoundedDate == nil, "foundedDate"); ok {

		args = append(args, *cryptoAsset.FoundedDate)
	}

	if ok := createSetClause(sqlBuffer, &setClauseWritten, cryptoAsset.CoinType == nil, "coinType"); ok {
		args = append(args, *cryptoAsset.CoinType)
	}

	if ok := createSetClause(sqlBuffer, &setClauseWritten, cryptoAsset.Website == nil, "website"); ok {
		args = append(args, *cryptoAsset.Website)
	}

	var updateCryptoAssetClause *statement
	if setClauseWritten {
		sqlBuffer.WriteString(" WHERE id = ?;")
		args = append(args, id)
		updateCryptoAssetClause = &statement{sql: sqlBuffer.String(), args: args}
	}

	return updateCryptoAssetClause
}

// createSetClause writes a set clause to the SQL buffer is the value passed is non-nil.
func createSetClause(sqlBuffer *bytes.Buffer, setClauseWritten *bool, isNil bool, column string) bool {
	if !isNil {
		if *setClauseWritten {
			sqlBuffer.WriteString(", ")
		}
		*setClauseWritten = true
		sqlBuffer.WriteString(fmt.Sprintf("%s = ?", column))
		return true
	}

	return false
}

// Insert team members inserts each team member from the array into the team_member table as part of a SQL transaction.
func insertTeamMembers(transaction *sql.Tx, id int, team []string) error {
	// Prepare the insert into the relation table which will be used multiple times.
	stmt, err := transaction.Prepare("INSERT INTO team_member(cryptoAssetId, name) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert each team member into the relation table. If it is a foreign key constraint error an unknown id error is
	// returned because the id could not be found in the crypto_asset table.
	for _, teamMember := range team {
		_, err := stmt.Exec(id, teamMember)
		if err != nil {
			sqliteError := err.(sqlite3.Error)
			if sqliteError.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return NewUnknownIDError(id)
			}
			return sqliteError
		}
	}

	return nil
}
