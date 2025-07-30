# Introduction

Welcome to my project! This is an example of a simple e-commerce CRUD API built with Golang, Gin, and Gorm. The database schema was designed using Gorm's auto-migration feature.

The project attempts to follow the Clean Architecture design pattern, although it is not yet fully implemented. Currently, the repository layer is separate, while the service and business layers are merged. and authentication middleware is used mock token for simplicity.

I have also included a Dockerfile and a docker-compose.yml file to run this project in a container.

I am not yet proficient in unit testing, so I used GitHub Copilot to generate test code. The tests primarily cover the handler and repository layers, where most of the logic resides.

To run a demo of this project, please follow the steps below.

# Api Endpoints

All endpoints require authentication with a Bearer token in the Authorization header: `Authorization: Bearer 12352`

You can test the endpoints using Postman Json collection that I have provided.

## Product Endpoints

| Method | Endpoint        | Description            |
| ------ | --------------- | ---------------------- |
| GET    | `/products`     | Get all products       |
| POST   | `/products`     | Create a new product   |
| DELETE | `/products/:id` | Delete a product by ID |

## Order Endpoints

| Method | Endpoint                             | Description                 |
| ------ | ------------------------------------ | --------------------------- |
| GET    | `/orders`                            | Get all orders              |
| GET    | `/orders/transaction/:transactionId` | Get order by transaction ID |
| POST   | `/orders`                            | Create a new order          |
| PUT    | `/orders/:id/status`                 | Update order status         |

# Running the Project

## start the project

```
sh run.sh
```

try demo with Postman collection that I have provided.

## stopping the project

```
sh down.sh
```
