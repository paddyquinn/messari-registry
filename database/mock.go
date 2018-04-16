package database

import (
	"github.com/paddyquinn/messari/database/models"
	"github.com/stretchr/testify/mock"
)

// Mock is a mock database.
type Mock struct {
	mock.Mock
}

// Insert mocks a crypto asset insert into the database.
func (m *Mock) Insert(cryptoAsset *models.CryptoAsset) (string, error) {
	args := m.Called(cryptoAsset)
	return args.String(0), args.Error(1)
}

// Select mocks a search for crypto assets from the database.
func (m *Mock) Select(names, symbols, fundingStatuses, coinTypes []string, startDate,
	endDate string) ([]*models.CryptoAsset, error) {

	args := m.Called(names, symbols, fundingStatuses, coinTypes, startDate, endDate)
	cryptoAssets, ok := args.Get(0).([]*models.CryptoAsset)
	if !ok {
		return nil, args.Error(1)
	}

	return cryptoAssets, args.Error(1)
}

// Update mocks an update to a crypto asset in the database.
func (m *Mock) Update(id int, cryptoAsset *models.CryptoAsset) error {
	args := m.Called(id, cryptoAsset)
	return args.Error(0)
}

// Close does nothing since there is no actual database to close.
func (m *Mock) Close() {}
