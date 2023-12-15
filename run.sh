#!/bin/bash

set -a
source .env
templ generate ./templates 
go run main.go serve \
    --stripe-api-key=$STRIPE_API_KEY \
    --stripe-webhook-secret=$STRIPE_WEBHOOK_SECRET \
    --pg=$CONNECTION_STRING \
    --addr="localhost:8080"
