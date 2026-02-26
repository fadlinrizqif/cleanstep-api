# Cleansteap API

## Overview
Cleanstep API adalah API berbasis RESTFul API yang berfungsi sebagai backend untuk e-commerce. Disini client bisa memgambil atau membuat order untuk berdasar struktur yang sudah ditentukan. e-commerce API ini berfokus pada usaha pembersih sepatu. 

## Features
- User Authentication (Login/Logout)
- JWT-based authentication with refresh token
- Provide product data for client
- Client can POST order

## Tech Stack
- Go (net/http)
- Gin Web Framework
- Goose
- SQLC
- JWT Authentication
- Cookies Base Authentication

## Project Structure
├── internal
│   ├── app           # api config struct
│   ├── auth          # JWT & refresh token logic
│   ├── database      # query database (SQLC)
│   ├── dto           # Data Transfer Object for respond struct
│   ├── handler       # HTTP handlers
│   └── middleware    # authentication middleware
└── sql
    ├── queries       # query to database 
    └── schema        # migration for database using Goose

## Environment Variables
| Variable | Description |
|--------|------------|
| DB_URL | PostgreSQL connection string |
| SERVER_SECRET | JWT signing secret |

## Routes
| Method | Path | Description |
|------|------|------------|
| POST | /api/signup | Register New user |
| POST | /api/login | Authenticate user |
| GET | /logout | Logout user |
| POST | /api/admin/products | Add new product to the database |
| POST | /api/admin/products/bulk | Add multiple new product to the database |
| GET | /api/products | Get product data shortlink |
| GET | /api/product/{productID} | Get product by id |
| POST | /api/orders | Create new order |

## Security Notes
- Password Hashed
- JWT stored in HttpOnly cookies
- Refresh token stored in database 
- CSRF mitigated via SameSite cookie
