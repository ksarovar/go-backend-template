# Golang Backend with User Authentication

This is a complete Golang backend application featuring user registration and login APIs, built with MongoDB for data storage, JWT for authentication, and AES encryption for sensitive data. The project includes Swagger documentation for API testing.

## Features

- User registration with email and password
- User login with JWT token generation
- AES-GCM encryption for user emails in database
- JWT-based authentication middleware
- Swagger UI for API documentation
- MongoDB integration
- Password hashing with bcrypt

## Prerequisites

Before setting up the project, ensure you have the following installed:

1. **Go** (version 1.19 or later)
   - Download from: https://golang.org/dl/
   - Verify installation: `go version`

2. **MongoDB** (version 4.4 or later)
   - Download from: https://www.mongodb.com/try/download/community
   - Start MongoDB service (default port: 27017)
   - For macOS with Homebrew: `brew services start mongodb/brew/mongodb-community`

3. **Git** (for cloning repositories)
   - Usually pre-installed on macOS

## Project Setup

### Step 1: Clone or Create Project Directory

If you're starting fresh, create a new directory:

```bash
mkdir golang-backend
cd golang-backend
```

### Step 2: Initialize Go Module

Initialize the Go module in your project directory:

```bash
go mod init golang-backend
```

### Step 3: Install Dependencies

Install all required Go packages:

```bash
go get github.com/gorilla/mux
go get go.mongodb.org/mongo-driver/mongo
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v4
go get github.com/swaggo/swag/cmd/swag
go get github.com/swaggo/http-swagger
```

### Step 4: Set Up Project Structure

Create the following directory structure:

```
golang-backend/
├── config/
│   └── config.go
├── database/
│   └── db.go
├── handlers/
│   └── auth.go
├── middleware/
│   └── auth.go
├── models/
│   └── user.go
├── utils/
│   └── encryption.go
├── docs/
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── main.go
├── go.mod
├── go.sum
└── README.md
```

### Step 5: Copy Code Files

Copy the code from the provided files into their respective locations:

- `config/config.go`
- `database/db.go`
- `handlers/auth.go`
- `middleware/auth.go`
- `models/user.go`
- `utils/encryption.go`
- `main.go`

### Step 6: Generate Swagger Documentation

Generate the Swagger documentation:

```bash
swag init
```

This will create/update the `docs/` directory with API documentation.

## Running the Project

### Step 1: Start MongoDB

Ensure MongoDB is running:

```bash
# On macOS with Homebrew
brew services start mongodb/brew/mongodb-community

# Or start manually
mongod
```

### Step 2: Run the Application

Start the Go server:

```bash
go run main.go
```

You should see output like:
```
MongoDB connected successfully
Server starting on :8080
```

### Step 3: Access the Application

- **API Base URL**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/

## API Endpoints

### Authentication
- `POST /register` - Register a new user
- `POST /login` - Login user

### User Routes (Protected)
- `GET /user/profile` - Get current user profile
- `PUT /user/profile` - Update current user profile

### Admin Routes (Protected - Admin Only)
- `GET /admin/users` - List all users with pagination
- `POST /admin/users/delete` - Delete a user by ID
- `PUT /admin/users/role` - Update user role (user/admin)

### Register User
- **URL**: `POST /register`
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- **Response**: `{"message": "User registered successfully"}`

### Login User
- **URL**: `POST /login`
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- **Response**: `{"token": "jwt_token_here"}`

## Testing the APIs

### Using cURL

Register a user:
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

Login:
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

Access protected user profile (replace TOKEN with actual JWT):
```bash
curl -X GET http://localhost:8080/user/profile \
  -H "Authorization: Bearer TOKEN"
```

Admin list users (replace TOKEN with admin JWT):
```bash
curl -X GET "http://localhost:8080/admin/users?page=1&limit=10" \
  -H "Authorization: Bearer TOKEN"
```

Admin delete user (replace TOKEN with admin JWT):
```bash
curl -X POST http://localhost:8080/admin/users/delete \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "USER_ID_HERE"}'
```

### Using Swagger UI

1. Open http://localhost:8080/swagger/ in your browser
2. Click on the endpoint you want to test
3. Click "Try it out"
4. Fill in the request parameters
5. Click "Execute"

## Configuration

The application uses a `.env` file for configuration. Copy the provided `.env` file and modify the values as needed:

```bash
# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017/golang_backend

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Encryption Configuration (must be 32 bytes for AES-256)
ENCRYPTION_KEY=12345678901234567890123456789012
```

**Important**: Change the `JWT_SECRET` and `ENCRYPTION_KEY` values in production for security.

Default values are provided in the code if environment variables are not set.

## Troubleshooting

### Port Already in Use

If you get "address already in use" error:

```bash
# Find process using port 8080
lsof -ti:8080

# Kill the process (replace PID with actual process ID)
kill -9 <PID>

# Or kill all processes on port 8080
lsof -ti:8080 | xargs kill -9
```

### MongoDB Connection Issues

- Ensure MongoDB is running: `brew services list | grep mongodb`
- Check MongoDB logs: `tail -f /usr/local/var/log/mongodb/mongo.log`
- Verify connection string in config

### Swagger Not Loading

- Ensure `swag init` was run successfully
- Check that `_ "golang-backend/docs"` import is present in `main.go`
- Restart the server after generating docs

### Build Issues

Clean and rebuild:
```bash
go clean -modcache
go mod tidy
go build
```

## Security Notes

- User emails are encrypted in the database using AES-GCM
- Passwords are hashed using bcrypt
- JWT tokens expire after 24 hours
- Role-based access control (user/admin roles)
- Admin-only endpoints for user management
- In production, use proper email hashing for lookups instead of plain text

## Development

To modify the code:

1. Make changes to the relevant files
2. Run `go mod tidy` to update dependencies
3. Regenerate Swagger docs: `swag init`
4. Restart the server: `go run main.go`

## License

This project is for educational purposes. Use appropriate licensing for production use.
