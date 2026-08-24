# Golang Project

## Prerequisites
- Go 1.21+

## Setup
1. Clone the project
2. Install dependencies:
    ```
    go mod tidy
    ```
## Setup project
 - Clone project
 - Run command
    ```
    go mod tidy
    ```

## Run
1. From the project root, run:
    ```
    go run main.go
    ```
2. The server will start on the port defined by `SERVER_PORT`.

## Database

- Migrate schema:
```
go run main.go migrate
```
- Seed database:
```
go run main.go seed
```

- Migrate and seed together:
```
go run main.go migrate seed
```
## Environment
- Copy environment from config.env.example
- Rename copied config.env.example into config.env

## Important Notes
- Currently all time are using time.now so it'll based on machine local time.