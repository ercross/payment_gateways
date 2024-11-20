package db

import "time"

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	Country   Country
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Gateway struct {
	ID                  int
	Name                string
	DataFormatSupported string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Country struct {
	ID        int
	Name      string
	Code      string
	Currency  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID          int
	Amount      float64
	Type        string
	Status      string
	Currency    string
	UserID      int
	GatewayName string
	CountryName string
	CreatedAt   time.Time
}

type GatewayPriority struct {
	Gateway   Gateway
	CountryID int
	Priority  int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserAccount struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
