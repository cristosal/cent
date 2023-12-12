module github.com/cristosal/cent

require (
	github.com/a-h/templ v0.2.476
	github.com/cristosal/pay v1.0.0
	github.com/jackc/pgx/v5 v5.5.0
)

require (
	github.com/cristosal/orm v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/stripe/stripe-go/v74 v74.30.0 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/cristosal/pay => ../pay

replace github.com/cristosal/pgxx => ../orm

replace github.com/cristosal/orm => ../orm

go 1.21.4
