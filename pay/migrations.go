package pay

import "github.com/cristosal/orm"

var migrations = []orm.Migration{
	{
		Name:        "customer table",
		Description: "create customers table",
		Up: `
		CREATE TABLE {{ .Schema }}.customer (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT Null,
			email VARCHAR(255) NOT Null,
			provider VARCHAR(32) NOT NUll,
			provider_id VARCHAR(255) NOT Null,
			UNIQUE (provider, provider_id)
		);`,
		Down: "DROP TABLE {{ .Schema }}.customer",
	},
	{
		Name:        "plan table",
		Description: "create plan table",
		Up: `
		CREATE TABLE {{ .Schema }}.plan (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			provider VARCHAR(255) NOT NULL,
			provider_id VARCHAR(255) NOT NULL,
			active BOOL NOT NULL,
			UNIQUE (provider, provider_id)
		);`,
		Down: "DROP TABLE {{ .Schema }}.plan",
	},
	{
		Name:        "price table",
		Description: "create price table",
		Up: `
		CREATE TABLE {{ .Schema }}.price (
			id SERIAL PRIMARY KEY,
			plan_id INT NOT NULL,
			provider VARCHAR(255) NOT NULL,
			provider_id VARCHAR(255) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			amount INT NOT NULL DEFAULT 0,
			schedule VARCHAR(32) NOT NULL,
			trial_days INT NOT NULL DEFAULT 0,
			FOREIGN KEY (plan_id) REFERENCES {{ .Schema }}.plan (id),
			UNIQUE (provider, provider_id)
		);`,
		Down: "DROP TABLE {{ .Schema }}.price",
	},
	{
		Name:        "subscription table",
		Description: "create subscription table",
		Up: `
		CREATE TABLE {{ .Schema }}.subscription (
			id SERIAL PRIMARY KEY,
			customer_id INT NOT NULL,
			price_id INT NOT NULL,
			provider VARCHAR(255) NOT NULL,
			provider_id VARCHAR(255) NOT NULL,
			active BOOL NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL,
			FOREIGN KEY (customer_id) REFERENCES {{ .Schema }}.customer(id),
			FOREIGN KEY (price_id) REFERENCES {{ .Schema }}.price(id),
			UNIQUE (provider, provider_id)
		);`,
		Down: "DROP TABLE {{ .Schema }}.subscription",
	},
	{
		Name:        "webhook event table",
		Description: "create webhook events table",
		Up: `CREATE TABLE {{ .Schema }}.webhook_event (
			id SERIAL PRIMARY KEY,
			provider VARCHAR(255) NOT NULL,
			provider_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			payload JSONB NOT NULL
		);`,
		Down: "DROP TABLE {{ .Schema }}.webhook_event",
	},
	{
		Name:        "subscription_user table",
		Description: "creates a subscription user table",
		Up: `CREATE TABLE {{ .Schema }}.subscription_user (
				username TEXT NOT NULL,
				subscription_id INT NOT NULL,
				FOREIGN KEY (subscription_id) REFERENCES {{ .Schema }}.subscription (id) ON DELETE CASCADE,
				PRIMARY KEY (username, subscription_id)
			)`,
		Down: "DROP TABLE {{ .Schema }}.subscription_user",
	},
}
