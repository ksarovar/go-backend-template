
# TODO: Golang Backend Structure with User Register and Login APIs (MongoDB + Encryption + Swagger)

## Steps to Complete
- [x] Initialize Go module with go.mod
- [ ] Initialize Go module with go.mod
- [x] Create config/config.go for configuration management
- [x] Create database/db.go for database connection and migration
- [x] Create models/user.go for User model
- [x] Create middleware/auth.go for JWT authentication middleware
- [x] Create handlers/auth.go for register and login handlers
- [x] Create handlers/admin.go for admin operations (list users, delete user, update role)
- [x] Create utils/encryption.go with HashEmail function
- [x] Create main.go as the entry point with routing and server setup
- [x] Run go mod tidy to download dependencies
- [x] Test the server startup (go run main.go)
- [x] Update README.md with new API endpoints and usage examples
- [x] Generate Swagger documentation for all APIs
- [x] Update dependencies: Remove GORM, add MongoDB driver
- [x] Update config/config.go for MongoDB URL
- [x] Update models/user.go to use bson tags
- [x] Update database/db.go to connect to MongoDB
- [x] Update handlers/auth.go for MongoDB queries
- [x] Run go mod tidy after changes
- [x] Test server startup with MongoDB
- [x] Create utils/encryption.go for AES encryption/decryption
- [x] Update config/config.go to include encryption key
- [x] Update handlers/auth.go to encrypt/decrypt user data
- [x] Run go mod tidy after encryption changes
- [x] Test server startup with encryption
- [x] Add swaggo/swag dependency for Swagger
- [x] Add Swagger annotations to handlers/auth.go
- [x] Add Swagger route to main.go
- [x] Generate Swagger docs with swag init
- [x] Run go mod tidy after Swagger changes
- [x] Test server startup with Swagger
