# Payroll

## Installation

Clone this repository and run `go mod tidy` in the root directory to download all dependencies.

## Setup

1. Rename `.env.example` to `.env`.
2. Fill in the required environment variables such as database credentials, secret key, etc.

## Testing Environment

Testing is done using a manually configured connection string.

Update the following line in your test files:
```go
connectionString := "postgresql://postgres:localdb123@localhost/test_db_1?sslmode=disable"
```
To run all tests:
```bash
go test ./... -count=1
```
The -count=1 flag disables test caching so results are always fresh.

## Running the Application

Make sure you have an empty database set up based on your .env configuration.

Execute init.sql to prepare the schema.

Optionally, execute data.sql to insert seed data.

To start the application:

```bash
go run .
```
or build
```
go build .
```

For how to use api, i provide the collection_curl

