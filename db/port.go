package db

import "errors"

var ErrDataNotFound = errors.New("data not found")

type Repository interface {
	CreateUser(User) error
	GetUserByID(int) (User, error)
	GetCountries() ([]Country, error)
	CreateTransaction(Transaction) (int, error)
	InsertGatewayPriority(GatewayPriority) error
	GetGatewayPriorities(countryID int) ([]GatewayPriority, error)
	GetGatewayByName(string) (Gateway, error)
	GetUserCountryByUserID(int) (Country, error)
	GetUserAccount(userID int) (*UserAccount, error)
	UpdateUserBalance(userID int, amount float64) error
	GetTransactionByID(int) (Transaction, error)
	UpdateTransactionStatus(id int, newStatus string) error
}

type Mock struct{}

func (m *Mock) CreateTransaction(tx Transaction) (int, error) { return 1, nil }

func (m *Mock) CreateUser(User) error                       { return nil }
func (m *Mock) GetUserByID(int) (User, error)               { return User{}, nil }
func (m *Mock) GetCountries() ([]Country, error)            { return []Country{}, nil }
func (m *Mock) InsertGatewayPriority(GatewayPriority) error { return nil }
func (m *Mock) GetGatewayPriorities(countryID int) ([]GatewayPriority, error) {
	return make([]GatewayPriority, 0), nil
}
func (m *Mock) GetGatewayByName(string) (Gateway, error)               { return Gateway{}, nil }
func (m *Mock) GetUserCountryByUserID(int) (Country, error)            { return Country{}, nil }
func (m *Mock) GetUserAccount(userID int) (*UserAccount, error)        { return &UserAccount{}, nil }
func (m *Mock) UpdateUserBalance(userID int, amount float64) error     { return nil }
func (m *Mock) GetTransactionByID(int) (Transaction, error)            { return Transaction{}, nil }
func (m *Mock) UpdateTransactionStatus(id int, newStatus string) error { return nil }
