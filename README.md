# go-authkit

`go-authkit` is a Go SDK for simplifying authentication in your web applications. It provides foundational components for token management, password handling, and common authentication flows, designed to be integrated into your existing projects.

**Status:** Alpha - API is subject to change.

## Core Idea

* **SDK provides the tools:** Secure token (Paseto) and password (bcrypt) handling, and logic for core auth operations.
* **You bring your storage & email:** Implement simple interfaces (`UserStorer`, `EmailSender`) to connect `go-authkit` to your database and email service.
* **Gin support (initially):** Includes Gin middleware and optional pre-built handlers.

## Features (Current & Planned)

* [X]  PASETO Token Generation & Verification (`token/`)
* [X]  Bcrypt Password Hashing & Checking (`hash/`)
* [X]  Core User Struct & Statuses (`core/user.go`)
* [X]  Interfaces for Application Integration:
  * [X]  `core.UserStorer` (for user data persistence)
  * [X]  `core.EmailSender` (for sending emails)
* [X]  Configurable Settings (`config/`)
* [X]  Gin Framework Support (`ginhandler/`):
  * [X]  Authentication Middleware
  * [X]  Role-Based Access Control Middleware
  * [X]  (Optional) Pre-built handlers for Register, Login, Verify Email, Password Reset, User Info, Logout.
* [ ]  (Planned) OAuth2 Provider Integration (e.g., Google)
* [ ]  (Planned) Default `UserStorer` implementation for PostgreSQL (sqlc)
* [ ]  (Planned) More comprehensive examples and documentation

## Structure Overview

* `config/`: SDK configuration.
* `core/`: Core interfaces (`UserStorer`, `EmailSender`), user model, error types.
* `token/`: PASETO token logic.
* `hash/`: Password hashing.
* `ginhandler/`: Gin-specific middleware, handlers, request/response DTOs, utils.
* `example/`: (Coming Soon) A simple runnable example using Gin.

## Getting Started (Conceptual - Gin)

1. **Get the SDK:**

   ```bash
   go get github.com/shawgichan/go-authkit 
   ```
2. **Implement Interfaces:**

   * Create your implementation of `core.UserStorer` to interact with your database.
   * Create your implementation of `core.EmailSender` to use your email service.
3. **Initialize in Your Application:**

   ```go
   // In your main.go or setup function:
   import (
       "github.com/shawgichan/go-authkit/config"
       "github.com/shawgichan/go-authkit/token"
       "github.com/shawgichan/go-authkit/hash"
       "github.com/shawgichan/go-authkit/ginhandler"
       // Your implementations:
       // "yourapp/internal/mystore" 
       // "yourapp/internal/mymailer"
   )

   // Load SDK config
   sdkConfig := config.DefaultAuthConfig()
   sdkConfig.TokenSymmetricKey = "your_32_byte_secret_key" // From env
   sdkConfig.AppBaseURL = "https://yourapp.com"

   // Init SDK components
   tokenMaker, _ := token.NewPasetoMaker(sdkConfig.TokenSymmetricKey)
   passwordHasher := hash.NewBcryptHasher(0)

   // Init your implementations
   // myAppUserStore := mystore.New(yourDBConnection)
   // myAppEmailSender := mymailer.New(yourEmailClientConfig)

   // Init SDK's Gin Handler (optional - if you use its pre-built handlers)
   // authAPI := ginhandler.NewAuthGinHandler(myAppUserStore, tokenMaker, passwordHasher, myAppEmailSender, sdkConfig)

   // Setup Gin router
   router := gin.Default()

   // Use SDK middleware
   // authGroup := router.Group("/api")
   // authGroup.Use(ginhandler.AuthMiddleware(tokenMaker, myAppUserStore, sdkConfig))
   // {
   //    // Your protected routes
   //    // authGroup.GET("/profile", authAPI.UserInfoHandler) // Example using SDK handler
   // }

   // Or, adapt your existing application handlers to use sdkTokenMaker, sdkPasswordHasher, 
   // myAppUserStore (as core.UserStorer), and myAppEmailSender (as core.EmailSender).
   ```

## Contribution

This SDK is in its early stages. Feedback, ideas, and contributions are welcome! Please open an issue to discuss.

## License

This project is licensed under the **MIT License**. See the [LICENSE.md](LICENSE.md) file for details.
