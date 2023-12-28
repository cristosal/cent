# cent

payment micro-service using https://github.com/cristosal/cent/pay

## Features

- Local database copy of all customer, subscription, plan and price data
- Sync and integration with Stripe via webhooks
- Web portal for viewing and managing entities
- [NATS](https://nats.io) pub sub, and request reply integration
- Entire API exposed via CLI make any request and view data directly from the terminal.

## Installation

Make sure you have go installed in your system, then run

`go install github.com/cristosal/cent`

Now you will have the cent CLI installed and the `cent` command available

## Getting Started

To start the microservice the command is `cent serve` you will need to pass the flags to pass in your database connection string and nats url.

```bash
cent serve \
    --stripe-api-key=$STRIPE_API_KEY \
    --stripe-webhook-secret=$STRIPE_WEBHOOK_SECRET \
    --pg=$CONNECTION_STRING \
    --addr="localhost:8080"

```

Once the service is started it will automatically sync with stripe and make a local copy of all your customers, plans, prices, and subscriptions

type in `cent -h` to view all available commands. They are pretty straightforward for the most part.
