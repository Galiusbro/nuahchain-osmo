# Authentication System Documentation

## Overview

The server implements a comprehensive authentication system supporting both web-based (email/password) and Telegram WebApp authentication. Each registered user automatically receives a Cosmos blockchain wallet with encrypted private keys stored securely on the server.

## Features

- **Web Registration**: Email/password based user registration
- **Telegram Integration**: Registration via Telegram WebApp
- **Automatic Wallet Creation**: Cosmos blockchain wallet generation for each user
- **Secure Key Storage**: Private keys encrypted using AES-GCM
- **JWT Tokens**: Stateless authentication with JSON Web Tokens
- **Session Management**: Token-based session handling
- **Protected Endpoints**: Middleware for endpoint protection

## Architecture

### Database Schema

The authentication system uses the following PostgreSQL tables:

- **users**: User accounts (email, telegram, credentials)
- **wallets**: Blockchain wallets with encrypted private keys
- **sessions**: JWT token sessions and tracking
- **telegram_auth**: Telegram authentication data

### Components

```
server/auth/
├── crypto.go        # Encryption/decryption utilities (AES-GCM)
├── jwt.go          # JWT token generation and validation
├── middleware.go   # Authentication middleware
├── models.go       # Data models
├── repository.go   # Database operations
├── service.go      # Business logic
└── wallet.go       # Cosmos wallet generation
```

## API Endpoints

### POST /api/auth/register

Register a new user via web (email/password).

**Request:**
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "created_at": "2025-11-02T23:00:00Z",
    "updated_at": "2025-11-02T23:00:00Z",
    "is_active": true
  },
  "wallet": {
    "id": 1,
    "user_id": 1,
    "address": "cosmos1cpy2zpd7tetxp63v4dayn9zywj2zw3p4a5rwcj",
    "created_at": "2025-11-02T23:00:00Z",
    "updated_at": "2025-11-02T23:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Registration successful"
}
```

**Status Codes:**
- `201 Created`: Registration successful
- `400 Bad Request`: Invalid input or user already exists
- `500 Internal Server Error`: Server error

### POST /api/auth/login

Login with email and password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "user": { ... },
  "wallet": { ... },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login successful"
}
```

**Status Codes:**
- `200 OK`: Login successful
- `401 Unauthorized`: Invalid credentials
- `400 Bad Request`: Missing required fields

### POST /api/auth/telegram

Register or login via Telegram WebApp.

**Request:**
```json
{
  "id": 123456789,
  "first_name": "John",
  "last_name": "Doe",
  "username": "johndoe",
  "auth_date": 1698950000,
  "hash": "telegram_hash_here"
}
```

**Response:**
```json
{
  "user": { ... },
  "wallet": { ... },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Telegram authentication successful"
}
```

**Status Codes:**
- `200 OK`: Authentication successful
- `400 Bad Request`: Invalid Telegram data
- `500 Internal Server Error`: Server error

### GET /api/auth/me

Get current user information (requires authentication).

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "created_at": "2025-11-02T23:00:00Z",
    "updated_at": "2025-11-02T23:00:00Z",
    "is_active": true
  },
  "wallet": {
    "id": 1,
    "address": "cosmos1cpy2zpd7tetxp63v4dayn9zywj2zw3p4a5rwcj"
  }
}
```

**Status Codes:**
- `200 OK`: Success
- `401 Unauthorized`: Missing or invalid token

## Security

### Encryption

Private keys are encrypted using **AES-GCM** (Advanced Encryption Standard in Galois/Counter Mode) before storage. The encryption key is derived from the `AUTH_MASTER_KEY` environment variable.

**Key Requirements:**
- Must be set via `AUTH_MASTER_KEY` environment variable
- Recommended: Generate using `openssl rand -base64 32`
- Should be kept secret and never committed to version control

### Password Hashing

Passwords are hashed using **bcrypt** with default cost (10 rounds). The hash is stored in the database, never the plain text password.

### JWT Tokens

JWT tokens use **HS256** (HMAC-SHA256) signing algorithm.

**Token Structure:**
```json
{
  "user_id": 1,
  "address": "cosmos1...",
  "exp": 1762187639,
  "iat": 1762101239,
  "nbf": 1762101239
}
```

**Configuration:**
- Default expiry: 24 hours
- Refresh token expiry: 7 days
- Secret key: Set via `JWT_SECRET` environment variable

## Configuration

### Environment Variables

Required environment variables:

```bash
# Authentication encryption key (REQUIRED)
export AUTH_MASTER_KEY="$(openssl rand -base64 32)"

# JWT secret key (REQUIRED for production)
export JWT_SECRET="$(openssl rand -base64 32)"

# Database configuration
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="serverdb"
```

### Default Values

- **Token Expiry**: 24 hours
- **Refresh Token Expiry**: 7 days
- **JWT Secret**: `your-secret-key-change-in-production` (development only)

## Wallet Generation

### Cosmos Wallet Creation

When a user registers, the system automatically:

1. Generates a BIP39 mnemonic (12/24 words)
2. Derives a Cosmos private key using HD wallet derivation path `44'/118'/0'/0/0`
3. Encrypts the private key using AES-GCM
4. Encrypts the mnemonic phrase
5. Stores both encrypted values in the database
6. Returns the Cosmos address to the user

**Derivation Path**: `m/44'/118'/0'/0/0`
- `44'`: BIP44 coin type
- `118'`: Cosmos coin type
- `0'`: Account
- `0`: Change
- `0`: Address index

## Testing

### Test Results

All authentication endpoints have been tested and verified:

✅ **Server health**: OK
✅ **Database health**: OK
✅ **User registration**: Working
✅ **JWT token generation**: Working
✅ **Protected endpoint `/api/auth/me`**: Working
✅ **Login with correct password**: Working
✅ **Login with wrong password**: Returns error
✅ **Missing token protection**: Working

### Example Test Commands

**Register a new user:**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "testpassword123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "testpassword123"
  }'
```

**Get current user (protected):**
```bash
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer <your_token_here>"
```

**Test invalid login:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "wrongpassword"
  }'
# Expected: "Invalid credentials"
```

**Test missing token:**
```bash
curl -X GET http://localhost:8080/api/auth/me
# Expected: "Authorization header required"
```

## Database Migrations

The authentication schema is created automatically on server startup via migration file:
- `server/migrations/001_initial_auth.sql`

The migration creates:
- `users` table with email/telegram support
- `wallets` table with encrypted key storage
- `sessions` table for JWT token tracking
- `telegram_auth` table for Telegram verification data
- Appropriate indexes for performance

## Error Handling

### Common Error Responses

**Invalid Credentials:**
```json
{
  "error": "Invalid credentials"
}
```
Status: `401 Unauthorized`

**User Already Exists:**
```json
{
  "error": "user with this email already exists"
}
```
Status: `400 Bad Request`

**Missing Authorization:**
```json
{
  "error": "Authorization header required"
}
```
Status: `401 Unauthorized`

**Invalid Token:**
```json
{
  "error": "Invalid or expired token"
}
```
Status: `401 Unauthorized`

## Development Notes

### Testing with Air

The server can be started with authentication keys using:

```bash
export AUTH_MASTER_KEY="your-key-here"
export JWT_SECRET="your-secret-here"
air
```

Or use the provided script:

```bash
./server/scripts/start_with_auth.sh
```

### Logging

All authentication events are logged via the server's logging system. In development mode, logs include:
- User registration events
- Login attempts
- Token validation
- Wallet creation

## Future Enhancements

Potential improvements:
- Refresh token implementation
- Two-factor authentication (2FA)
- Password reset functionality
- Email verification
- Rate limiting for login attempts
- Session management dashboard
- Telegram hash verification (currently skipped in development)

## Security Best Practices

1. **Never commit** `AUTH_MASTER_KEY` or `JWT_SECRET` to version control
2. **Use strong keys**: Generate using `openssl rand -base64 32`
3. **Rotate keys**: Change keys periodically in production
4. **Use HTTPS**: Always use TLS/SSL in production
5. **Monitor logs**: Watch for suspicious authentication patterns
6. **Implement rate limiting**: Prevent brute force attacks
7. **Validate input**: All user input is validated and sanitized

## Support

For issues or questions, refer to the main server README or project documentation.

