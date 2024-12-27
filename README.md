# Go Shop API

A simple e-commerce REST API built with Go, Gin and GORM.

## Features

- User authentication (signup/login) with JWT
- Role-based access control (Admin/Customer)
- Product management (CRUD operations)
- Order management with status tracking
- Input validation
- Pagination
- Error handling

## Tech Stack

- [Go](https://golang.org/) - Programming language
- [Gin](https://gin-gonic.com/) - Web framework
- [GORM](https://gorm.io/) - ORM library
- [JWT-Go](https://github.com/golang-jwt/jwt) - JWT implementation
- PostgreSQL - Database

## API Endpoints

### Authentication

- `POST /auth/signup` - Register a new user
- `POST /auth/login` - Login user

### Products (Admin only)

- `POST /products` - Create product
- `GET /products` - List all products
- `GET /products/:id` - Get product details
- `PUT /products/:id` - Update product
- `DELETE /products/:id` - Delete product

### Orders

- `POST /orders` - Create order (Auth required)
- `GET /orders` - List user orders (Auth required)
- `GET /orders/:id` - Get order details (Auth required)
- `POST /orders/:id/cancel` - Cancel order (Auth required)
- `PUT /orders/:id/status` - Update order status (Admin only)

## Project Structure

├── controllers/
│ ├── usersController.go
│ ├── productsController.go
│ └── ordersController.go
├── models/
│ ├── userModel.go
│ ├── productModel.go
│ └── orderModel.go
├── initializers/
│ └── connectToDb.go
| |** loadEnvVariables.go
| |** syncDb.go
├── middlewares/
│ └── requireAuth.go
| |** requireAdmin.go
├── validators/
│ └── password.go
├── services/
│ └── usersService.go
└── libs/
│ └── errors.go
│ └── response.go
└── main.go
|** .env
|** .env.example
|** .gitignore
|** go.mod
|** go.sum
|\_\_ README.md

## Potential Improvements

Given that this was a simple project, there are many potential improvements that could be made:

- Implement a service layer to separate business logic from controllers
- Use UUIDs for IDs instead of auto-incrementing integers
- Audit trails to keep track of who creates products, cancels orders, updates order statuses or edit products etc.

## Getting Started

1. Clone the repository

[git clone https://github.com/roronoazor/goShopAPI.git](https://github.com/roronoazor/goShopAPI.git)

2. Install dependencies

```
    go mod download
```

3. Set up environment variables
   For convenience, i have pushed my .env file to the repo. You can use it as a reference. if you dont want to use the .env.example file

```
    cp .env.example .env
```

4. Edit .env with your database credentials and JWT secret key

5. Run the application

```
    go run main.go
```

6. Use Postman or curl to test the API endpoints

7. API Documentation

->
