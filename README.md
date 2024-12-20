# Stockk

Stockk is a RESTful service designed to track the stock of a restaurant. It provides APIs to manage orders, ingredients, and products, ensuring efficient stock management.

![CI Status](https://github.com/aradwann/stockk/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/aradwann/stockk)](https://goreportcard.com/report/github.com/aradwann/stockk)

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Running Tests](#running-tests)
- [License](#license)
- [Testing Workflow](#testing-workflow)

## Features

- **Manage orders, ingredients, and products**  
  - All updates to related entities are handled in a single transaction, ensuring strong consistency across the system.  
- **Track stock levels in real-time and alert merchants about low stock.**  
  - each order triggers a check for ingredients stock and enqueue a task of sending email notification about the low stock ingredients.  
- **Redis-based task distribution for background processing**  
  - tasks are enqueued to redis ensuring presistence in case app server is restarted and potential horizontal scaling,

  - email tasks are processed by a goroutine acts as worker, ensuring a concurrent performant excution of the tasks.
- Middleware for **structured** logging, CORS, and request handling.  

- Graceful server shutdown.  
- [end-to-end tested](./test/e2e_test.go) with test conatiners.

## Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/aradwann/stockk.git
    cd stockk
    ```

## Configuration

Configuration is managed through environment variables.

**Create a `.env` file in the root directory take `.env.dev` as base and modify it based on your needs.**

knowing that you have to set the following variable to have a successful alert email flow
you have to generate an [google app password](https://support.google.com/accounts/answer/185833?hl=en) for the mailer email and set it accordingly

```.env
EMAIL_SENDER_NAME=stockk
EMAIL_SENDER_ADDRESS= // mailer email
EMAIL_SENDER_PASSWORD= // mailer password
TEST_MERCHANT_EMAIL= // test merchant email (reciever of alerts)
```

## Usage

### Using Docker Compose

1. Ensure Docker and Docker Compose are installed on your machine.
2. Start the services using Docker Compose:

The application will be available at <http://localhost:8080>.

## API Endpoints

### Orders

- **Create Order**
  - `POST /api/v1/orders`
  - Request Body: `{ "product_id": "1", "quantity": 2 }`
  - Response: `201 Created`

### Health Check

- **Health Check**
  - `GET /health`
  - Response: `200 OK`

## Running Tests

To run the tests, use the following command:

```sh
go test ./...
```

## Testing Workflow

(for development purpose)

### Accessing pgAdmin4 Dashboard

prerequist: ensure .env file is in place and has proper email credientials

1. Ensure Docker and Docker Compose are installed on your machine.
2. Start the services using Docker Compose:

    ```sh
    docker-compose up --build
    ```

3. Open your web browser and go to [http://localhost:80](http://localhost:80).
4. Log in with the default credentials:
    - Email: `user@domain.com`
    - Password: `SuperSecret`

5. a server is added automatically via `servers.json` file:
    - DB server Password: `secret`
6. you can inspect and verify the DB changes

### Testing API Endpoints with cURL

#### Create Order

```sh
curl -X POST http://localhost:8080/api/v1/orders -H "Content-Type: application/json" -d '{"product_id": "1", "quantity": 2}'
```

## License

This project is licensed under the MIT License. See the [`LICENSE`](LICENSE ) file for details.
