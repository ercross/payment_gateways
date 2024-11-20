-- Seed gateways table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM gateways WHERE name = 'Stripe') THEN
        INSERT INTO gateways (name, data_format_supported)
        VALUES
        ('Stripe', 'JSON');
END IF;
END $$;

-- Seed countries table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM countries WHERE name = 'United States') THEN
        INSERT INTO countries (name, code, currency)
        VALUES
        ('United States', 'US', 'USD');
END IF;

    IF NOT EXISTS (SELECT 1 FROM countries WHERE name = 'Canada') THEN
        INSERT INTO countries (name, code, currency)
        VALUES
        ('Canada', 'CA', 'CAD');
END IF;
END $$;

-- Seed gateway_countries table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM gateway_countries WHERE gateway_id = (SELECT id FROM gateways WHERE name = 'Stripe') AND country_id = (SELECT id FROM countries WHERE name = 'United States')) THEN
        INSERT INTO gateway_countries (gateway_id, country_id)
        VALUES
        ((SELECT id FROM gateways WHERE name = 'Stripe'), (SELECT id FROM countries WHERE name = 'United States'));
END IF;

    IF NOT EXISTS (SELECT 1 FROM gateway_countries WHERE gateway_id = (SELECT id FROM gateways WHERE name = 'Stripe') AND country_id = (SELECT id FROM countries WHERE name = 'Canada')) THEN
        INSERT INTO gateway_countries (gateway_id, country_id)
        VALUES
        ((SELECT id FROM gateways WHERE name = 'Stripe'), (SELECT id FROM countries WHERE name = 'Canada'));
END IF;
END $$;

-- Seed gateway_priority table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM gateway_priority WHERE gateway_id = (SELECT id FROM gateways WHERE name = 'Stripe') AND country_id = (SELECT id FROM countries WHERE name = 'United States') AND priority = 1) THEN
        INSERT INTO gateway_priority (gateway_id, country_id, priority, is_active)
        VALUES
        ((SELECT id FROM gateways WHERE name = 'Stripe'), (SELECT id FROM countries WHERE name = 'United States'), 1, TRUE);
END IF;

    IF NOT EXISTS (SELECT 1 FROM gateway_priority WHERE gateway_id = (SELECT id FROM gateways WHERE name = 'Stripe') AND country_id = (SELECT id FROM countries WHERE name = 'Canada') AND priority = 1) THEN
        INSERT INTO gateway_priority (gateway_id, country_id, priority, is_active)
        VALUES
        ((SELECT id FROM gateways WHERE name = 'Stripe'), (SELECT id FROM countries WHERE name = 'Canada'), 1, TRUE);
END IF;
END $$;

-- Seed users table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'john.doe@example.com') THEN
        INSERT INTO users (username, email, password, country_id)
        VALUES
        ('john_doe', 'john.doe@example.com', 'hashed_password', (SELECT id FROM countries WHERE name = 'United States'));
END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'jane.doe@example.com') THEN
        INSERT INTO users (username, email, password, country_id)
        VALUES
        ('jane_doe', 'jane.doe@example.com', 'hashed_password', (SELECT id FROM countries WHERE name = 'Canada'));
END IF;
END $$;

-- Seed user_accounts table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM user_accounts WHERE user_id = (SELECT id FROM users WHERE email = 'john.doe@example.com')) THEN
        INSERT INTO user_accounts (user_id, balance, currency)
        VALUES
        ((SELECT id FROM users WHERE email = 'john.doe@example.com'), 1000.00, 'USD');
END IF;

    IF NOT EXISTS (SELECT 1 FROM user_accounts WHERE user_id = (SELECT id FROM users WHERE email = 'jane.doe@example.com')) THEN
        INSERT INTO user_accounts (user_id, balance, currency)
        VALUES
        ((SELECT id FROM users WHERE email = 'jane.doe@example.com'), 500.00, 'CAD');
END IF;
END $$;

-- Seed transactions table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM transactions WHERE user_id = (SELECT id FROM users WHERE email = 'john.doe@example.com') AND amount = 200.00 AND status = 'completed') THEN
        INSERT INTO transactions (amount, type, status, currency, gateway_name, country_name, user_id)
        VALUES
        (200.00, 'withdrawal', 'completed', 'USD', (SELECT id FROM gateways WHERE name = 'Stripe'), (SELECT id FROM countries WHERE name = 'United States'), (SELECT id FROM users WHERE email = 'john.doe@example.com'));
END IF;

    IF NOT EXISTS (SELECT 1 FROM transactions WHERE user_id = (SELECT id FROM users WHERE email = 'jane.doe@example.com') AND amount = 150.00 AND status = 'completed') THEN
        INSERT INTO transactions (amount, type, status, currency, gateway_name, country_name, user_id)
        VALUES
        (150.00, 'deposit', 'completed', 'CAD', (SELECT id FROM gateways WHERE name = 'Stripe'), (SELECT id FROM countries WHERE name = 'Canada'), (SELECT id FROM users WHERE email = 'jane.doe@example.com'));
END IF;
END $$;
