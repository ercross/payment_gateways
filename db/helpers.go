package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ercross/payment_gateways/internal/services"
	"github.com/golang-migrate/migrate/v4"
	"os"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

func (p *DB) Migrate(migrationFilesDir string, dsn string) error {
	pwd, _ := os.Getwd()
	m, err := migrate.New(fmt.Sprintf("file://%s%s", pwd, migrationFilesDir), dsn)
	if err != nil {
		return fmt.Errorf("failed to obtain migrate instance: %w", err)
	}
	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("migration up failed: %w", err)
	}

	return nil
}

// New initializes the database connection
func New(dataSourceName string) (*DB, error) {
	var err error
	var db *sql.DB

	err = services.RetryOperation(func() error {
		db, err = sql.Open("postgres", dataSourceName)
		if err != nil {
			return err
		}
		return nil
	}, 5)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func (p *DB) CreateUser(user User) error {
	query := `INSERT INTO users (username, email, country_id, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := p.db.QueryRow(query, user.Username, user.Email, user.Country.ID, time.Now(), time.Now()).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %v", err)
	}
	return nil
}

func (p *DB) GetUserCountryByUserID(id int) (Country, error) {
	query := `
        SELECT c.id, c.name, c.code, c.currency, c.created_at, c.updated_at
        FROM users u
        INNER JOIN countries c ON u.country_id = c.id
        WHERE u.id = $1
    `

	var country Country
	err := p.db.QueryRow(query, id).Scan(
		&country.ID,
		&country.Name,
		&country.Code,
		&country.Currency,
		&country.CreatedAt,
		&country.UpdatedAt,
	)

	if err != nil {
		return Country{}, fmt.Errorf("failed to get user country: %w", err)
	}

	return country, nil

}

func (p *DB) GetTransactionByID(id int) (Transaction, error) {
	query := `
        SELECT id, amount, type, status, created_at, currency, gateway_name, country_name, user_id
        FROM transactions
        WHERE id = $1
    `

	var transaction Transaction
	err := p.db.QueryRow(query, id).Scan(
		&transaction.ID,
		&transaction.Amount,
		&transaction.Type,
		&transaction.Status,
		&transaction.CreatedAt,
		&transaction.Currency,
		&transaction.GatewayName,
		&transaction.CountryName,
		&transaction.UserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Transaction{}, ErrDataNotFound
		}
		return Transaction{}, fmt.Errorf("failed to get transaction: %w", err)
	}

	return transaction, nil
}

func (p *DB) UpdateTransactionStatus(id int, newStatus string) error {
	query := `
        UPDATE transactions
        SET status = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `

	result, err := p.db.Exec(query, newStatus, id)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no transaction found with id %d", id)
	}

	return nil
}

// GetUserByID queries the user by their ID
func (p *DB) GetUserByID(userID int) (User, error) {
	// SQL query to select the user by user_id
	query := `SELECT u.id, u.username, u.email, u.password, u.created_at, u.updated_at,
       			c.id, c.name, c.code, c.currency
              FROM users u
              JOIN countries c ON users.country_id = countries.id
              WHERE u.id = $1;`

	// Execute the query
	row := p.db.QueryRow(query, userID)

	// Prepare a User struct to hold the result
	var user User

	// Scan the result into the struct
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt,
		&user.UpdatedAt, &user.Country.ID, &user.Country.Name, &user.Country.Code, &user.Country.Currency)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no user is found
			return user, fmt.Errorf("user with ID %d not found", userID)
		}
		// If any other error occurs
		return user, fmt.Errorf("error querying user: %v", err)
	}

	// Return the user object
	return user, nil
}

func (p *DB) CreateGateway(gateway Gateway) error {
	query := `INSERT INTO gateways (name, data_format_supported, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	err := p.db.QueryRow(query, gateway.Name, gateway.DataFormatSupported, time.Now(), time.Now()).Scan(&gateway.ID)
	if err != nil {
		return fmt.Errorf("failed to insert gateway: %v", err)
	}
	return nil
}

func (p *DB) GetGateways() ([]Gateway, error) {
	rows, err := p.db.Query(`SELECT id, name, data_format_supported, created_at, updated_at FROM gateways`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gateways: %v", err)
	}
	defer rows.Close()

	var gateways []Gateway
	for rows.Next() {
		var gateway Gateway
		if err := rows.Scan(&gateway.ID, &gateway.Name, &gateway.DataFormatSupported, &gateway.CreatedAt, &gateway.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %v", err)
		}
		gateways = append(gateways, gateway)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return gateways, nil
}

func (p *DB) CreateCountry(country Country) error {
	query := `INSERT INTO countries (name, code, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	err := p.db.QueryRow(query, country.Name, country.Code, time.Now(), time.Now()).Scan(&country.ID)
	if err != nil {
		return fmt.Errorf("failed to insert country: %v", err)
	}
	return nil
}

func (p *DB) GetCountries() ([]Country, error) {
	rows, err := p.db.Query(`SELECT id, name, code, created_at, updated_at FROM countries`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries: %v", err)
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var country Country
		if err := rows.Scan(&country.ID, &country.Name, &country.Code, &country.CreatedAt, &country.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan country: %v", err)
		}
		countries = append(countries, country)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return countries, nil
}

func (p *DB) CreateTransaction(transaction Transaction) (int, error) {
	transaction.Status = "pending"
	query := `INSERT INTO transactions (amount, type, status, currency, gateway_name, country_name, user_id, created_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	err := p.db.QueryRow(query, transaction.Amount, transaction.Type, transaction.Status, transaction.Currency, transaction.GatewayName, transaction.CountryName, transaction.UserID, time.Now()).Scan(&transaction.ID)
	if err != nil {
		return -1, fmt.Errorf("failed to insert transaction: %v", err)
	}
	return transaction.ID, nil
}

func (p *DB) GetSupportedCountriesByGateway(gatewayID int) ([]Country, error) {
	query := `
		SELECT c.id AS country_id, c.name AS country_name
		FROM countries c
		JOIN gateway_countries gc ON c.id = gc.country_id
		WHERE gc.gateway_id = $1
		ORDER BY c.name
	`

	rows, err := p.db.Query(query, gatewayID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries for gateway %d: %v", gatewayID, err)
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var country Country
		if err := rows.Scan(&country.ID, &country.Name); err != nil {
			return nil, fmt.Errorf("failed to scan country: %v", err)
		}
		countries = append(countries, country)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %v", err)
	}

	return countries, nil
}

// InsertGatewayPriority inserts a new record into the gateway_priority table.
func (p *DB) InsertGatewayPriority(gp GatewayPriority) error {
	// Prepare the SQL query for inserting data
	query := `INSERT INTO gateway_priority (country_id, gateway_id, priority, is_active, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, NOW(), NOW())`

	// Execute the query with the provided values
	_, err := p.db.Exec(query, gp.CountryID, gp.Gateway.ID, gp.Priority, gp.IsActive)
	if err != nil {
		return fmt.Errorf("failed to insert gateway priority: %w", err)
	}

	return nil
}

// GetGatewayPriorities retrieves the list of gateway priorities for a specific country.
func (p *DB) GetGatewayPriorities(countryID int) ([]GatewayPriority, error) {
	// Query to fetch the gateway priorities for the given country
	query := `
		SELECT gp.country_id, gp.gateway_id, gp.priority, gp.is_active, gp.created_at, gp.updated_at,
		       g.name, g.data_format_supported
		FROM gateway_priority gp
		JOIN countries c ON gp.country_id = c.id
		JOIN gateways g ON gp.gateway_id = g.id
		WHERE gp.country_id = $1
		ORDER BY gp.priority ASC
	`

	// Prepare the query to execute and fetch results
	rows, err := p.db.Query(query, countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query gateway priorities: %w", err)
	}
	defer rows.Close()

	var gatewayPriorities []GatewayPriority

	// Iterate over the result set and map the data to GatewayPriority structs
	for rows.Next() {
		var gp GatewayPriority

		if err := rows.Scan(&gp.CountryID, &gp.Gateway.ID, &gp.Priority, &gp.IsActive, &gp.CreatedAt, &gp.UpdatedAt, &gp.Gateway.Name, &gp.Gateway.DataFormatSupported); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		// You can include the gateway name and data format in the response if necessary
		// For now, just return the GatewayPriority struct
		gatewayPriorities = append(gatewayPriorities, gp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return gatewayPriorities, nil
}

func (p *DB) GetGatewayByName(gatewayName string) (Gateway, error) {
	query := `
		SELECT id, name, data_format_supported, created_at, updated_at
		FROM gateways
		WHERE name = $1
		LIMIT 1;`

	row := p.db.QueryRow(query, gatewayName)

	var gateway Gateway

	err := row.Scan(&gateway.ID, &gateway.Name, &gateway.DataFormatSupported, &gateway.CreatedAt, &gateway.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			return gateway, fmt.Errorf("no gateway found with name: %s", gatewayName)
		}
		return gateway, fmt.Errorf("error querying gateway: %v", err)
	}

	return gateway, nil
}

// GetUserAccount fetches the user account details based on user ID.
func (p *DB) GetUserAccount(userID int) (*UserAccount, error) {
	query := `
        SELECT id, user_id, balance, currency, created_at, updated_at 
        FROM user_accounts 
        WHERE user_id = $1
    `
	row := p.db.QueryRow(query, userID)

	account := &UserAccount{}
	if err := row.Scan(&account.ID, &account.UserID, &account.Balance, &account.Currency, &account.CreatedAt, &account.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user account not found for user_id: %d", userID)
		}
		return nil, err
	}
	return account, nil
}

func (p *DB) UpdateUserBalance(userID int, amount float64) error {
	query := `
        UPDATE user_accounts 
        SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP
        WHERE user_id = $2
    `
	result, err := p.db.Exec(query, amount, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user account not found for user_id: %d", userID)
	}
	return nil
}
