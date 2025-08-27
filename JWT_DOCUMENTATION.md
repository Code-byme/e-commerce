# JWT Authentication System Documentation

## Overview

This e-commerce Go application uses JWT (JSON Web Tokens) for secure authentication. The system provides user registration, login, and protected route access with token-based authentication.

## 🔐 JWT Configuration

### Environment Variables

The JWT system uses the following environment variable:

```bash
JWT_SECRET=AANmjnkH0q6f5g9KASpjq6r6Arj0OHHnNhdYPrChFvX8pf1okVESyCPrex9SMBQcLBPNEEvMdGDLyiMPhVkcXg==
```

**⚠️ Security Note**: The JWT secret has been generated using OpenSSL for security. In production, always use a strong, unique secret key.

### Token Structure

JWT tokens contain the following claims:

```json
{
  "user_id": 1,
  "email": "user@example.com",
  "role": "customer",
  "exp": 1756374756,
  "iat": 1756288356,
  "nbf": 1756288356
}
```

- `user_id`: Unique user identifier
- `email`: User's email address
- `role`: User role (customer, admin, etc.)
- `exp`: Token expiration time (24 hours from creation)
- `iat`: Token issued at time
- `nbf`: Token not valid before time

## 🚀 API Endpoints

### Authentication Endpoints

#### 1. Register User
```http
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "customer",
      "created_at": "2024-01-01T00:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 2. Login User
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "customer",
      "created_at": "2024-01-01T00:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### Protected Endpoints

#### 3. Get User Profile
```http
GET /api/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "data": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "customer",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

## 🛠️ JWT CLI Tool

A command-line tool is provided for JWT token management and testing.

### Building the Tool
```bash
go build -o bin/jwt cmd/jwt/main.go
```

### Usage Examples

#### Generate a Test Token
```bash
./bin/jwt -action=generate -user-id=1 -email="test@example.com" -role="customer" -pretty
```

#### Validate a Token
```bash
./bin/jwt -action=validate -token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." -pretty
```

#### Decode a Token (without validation)
```bash
./bin/jwt -action=decode -token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." -pretty
```

### CLI Options

- `-action`: Action to perform (`generate`, `validate`, `decode`)
- `-user-id`: User ID for token generation
- `-email`: Email for token generation
- `-role`: Role for token generation
- `-token`: JWT token to validate or decode
- `-pretty`: Pretty print JSON output

## 🔧 Implementation Details

### File Structure

```
├── internal/
│   ├── services/
│   │   └── auth_service.go      # Core authentication logic
│   ├── handlers/
│   │   └── auth.go              # HTTP handlers
│   └── database/
│       └── migrations.go        # Database schema
├── pkg/
│   ├── middleware/
│   │   └── auth.go              # JWT middleware
│   └── utils/
│       └── jwt_generator.go     # JWT utilities
└── cmd/
    └── jwt/
        └── main.go              # CLI tool
```

### Key Components

1. **AuthService**: Handles user registration, login, and token generation
2. **AuthMiddleware**: Protects routes by validating JWT tokens
3. **JWTGenerator**: Utility for token generation and validation
4. **Database Migrations**: Creates users table with proper schema

### Security Features

- **Password Hashing**: bcrypt with default cost
- **Token Expiration**: 24-hour token lifetime
- **Input Validation**: Request validation with proper error messages
- **SQL Injection Protection**: Parameterized queries
- **Role-based Access**: Extensible role system

## 🧪 Testing

### Automated Testing
```bash
./test-auth.sh
```

### Manual Testing with curl

#### Register a new user
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

#### Login and get token
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### Access protected endpoint
```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🔒 Security Best Practices

1. **Strong JWT Secret**: Use a cryptographically secure random secret
2. **Token Expiration**: Set reasonable expiration times
3. **HTTPS Only**: Always use HTTPS in production
4. **Input Validation**: Validate all user inputs
5. **Error Handling**: Don't expose sensitive information in error messages
6. **Rate Limiting**: Implement rate limiting for auth endpoints
7. **Logging**: Log authentication events for security monitoring

## 🚀 Production Deployment

### Environment Setup
1. Generate a new JWT secret for production
2. Set up proper database credentials
3. Configure HTTPS
4. Set up monitoring and logging

### JWT Secret Generation
```bash
openssl rand -base64 64
```

### Environment Variables
```bash
export JWT_SECRET="your-production-secret"
export DB_HOST="your-db-host"
export DB_PASSWORD="your-db-password"
# ... other variables
```

## 📝 Error Handling

### Common Error Responses

#### Invalid Token
```json
{
  "error": "Invalid or expired token"
}
```

#### Missing Authorization Header
```json
{
  "error": "Authorization header is required"
}
```

#### Invalid Credentials
```json
{
  "error": "Invalid email or password"
}
```

#### User Already Exists
```json
{
  "error": "User with this email already exists"
}
```

## 🔄 Token Refresh

Currently, the system uses 24-hour tokens. For enhanced security, consider implementing:

1. **Refresh Tokens**: Short-lived access tokens with longer-lived refresh tokens
2. **Token Rotation**: Automatic token refresh
3. **Token Blacklisting**: Ability to invalidate tokens

## 📚 Additional Resources

- [JWT.io](https://jwt.io/) - JWT debugger and documentation
- [RFC 7519](https://tools.ietf.org/html/rfc7519) - JWT specification
- [Go JWT Library](https://github.com/golang-jwt/jwt) - Official Go JWT library
