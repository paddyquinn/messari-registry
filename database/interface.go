package database

import "github.com/paddyquinn/messari/database/models"

// Interface represents an interface any database driver or mock need adhere to.
type Interface interface {
	Insert(cryptoAsset *models.CryptoAsset) (string, error)
	Select(names, symbols, fundingStatuses, coinTypes []string, startDate, endDate string) ([]*models.CryptoAsset, error)
	Update(id int, cryptoAsset *models.CryptoAsset) error
	Close()
}
