DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gateways') THEN
        CREATE TABLE gateways (
                          id SERIAL PRIMARY KEY,
                          name VARCHAR(255) NOT NULL UNIQUE,
                          data_format_supported VARCHAR(50) NOT NULL,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'countries') THEN
        CREATE TABLE countries (
                           id SERIAL PRIMARY KEY,
                           name VARCHAR(255) NOT NULL UNIQUE,
                           code CHAR(2) NOT NULL UNIQUE,
                           currency CHAR(3) NOT NULL,
                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gateway_countries') THEN
        CREATE TABLE gateway_countries (
                                   gateway_id INT NOT NULL,
                                   country_id INT NOT NULL,
                                   PRIMARY KEY (gateway_id, country_id),
                                   FOREIGN KEY (country_id) REFERENCES countries(id) ON DELETE CASCADE,
                                   FOREIGN KEY (gateway_id) REFERENCES gateways(id) ON DELETE CASCADE
);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gateway_priority') THEN
        CREATE TABLE gateway_priority (
                                  country_id INT NOT NULL,
                                  gateway_id INT NOT NULL,
                                  priority INT NOT NULL,
                                  is_active BOOLEAN DEFAULT TRUE,
                                  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                  PRIMARY KEY (country_id, gateway_id, priority, is_active),
                                  FOREIGN KEY (country_id) REFERENCES countries(id) ON DELETE CASCADE,
                                  FOREIGN KEY (gateway_id) REFERENCES gateways(id) ON DELETE CASCADE
        );
    END IF;
END $$;


DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'transactions') THEN
        CREATE TABLE transactions (
                              id SERIAL PRIMARY KEY,
                              amount DECIMAL(10, 2) NOT NULL,
                              type VARCHAR(50) NOT NULL,
                              status VARCHAR(50) NOT NULL,
                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                              currency VARCHAR(50) NOT NULL,
                              gateway_name VARCHAR(75) NOT NULL,
                              country_name VARCHAR(75) NOT NULL,
                              user_id INT NOT NULL
);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(255) NOT NULL UNIQUE,
                       email VARCHAR(255) NOT NULL UNIQUE,
                       password VARCHAR(255) NOT NULL,
                       country_id INT,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_accounts') THEN
        CREATE TABLE user_accounts (
                               id SERIAL PRIMARY KEY,
                               user_id INT NOT NULL UNIQUE,                  -- Foreign key to users table
                               balance DECIMAL(18, 2) NOT NULL DEFAULT 0.0, -- User's current balance
                               currency CHAR(3) NOT NULL,                  -- Currency code (e.g., USD, EUR)
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Account creation timestamp
                               updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Last update timestamp
                               CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
        );
    END IF;
END $$;