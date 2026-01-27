# Claude Code Guidelines for Warframe Wishlist

## Project Overview

This is a Go backend API for a Warframe wishlist application using:
- Chi router for HTTP routing
- MongoDB for data persistence
- Supabase JWT authentication
- Clean architecture with services, repositories, and handlers

## Testing Policy

**CRITICAL: Tests are protected and must not be deleted.**

### Rules for Claude instances:

1. **NEVER delete any test files or test functions.** Tests exist to prevent regressions and ensure code quality.

2. **NEVER modify tests to make them pass by weakening assertions.** If a test is failing, the implementation code should be fixed, not the test.

3. **If you cannot make a test pass:**
   - Stop and inform the user about the failing test
   - Explain what the test is checking
   - Describe why the current implementation doesn't satisfy the test
   - Ask the user to manually fix the issue or provide guidance

4. **When adding new features:**
   - Add corresponding unit tests
   - Ensure all existing tests continue to pass
   - Run `go test ./...` before considering the task complete

5. **Test coverage expectations:**
   - All services must have unit tests
   - All handlers must have unit tests
   - Middleware must have unit tests
   - Edge cases and error conditions should be tested

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/services/...

# Run tests with coverage
go test -cover ./...
```

## Project Structure

```
cmd/server/main.go           # Entry point
internal/
  config/                    # Environment configuration
  database/                  # MongoDB connection
  middleware/                # JWT authentication
  models/                    # Data models
  repository/                # Data access layer
  services/                  # Business logic
  handlers/                  # HTTP handlers
  mocks/                     # Test mocks
pkg/response/                # API response helpers
```

## Key Interfaces

The codebase uses interfaces for dependency injection and testability:

- `repository.ItemRepositoryInterface`
- `repository.WishlistRepositoryInterface`
- `services.ItemServiceInterface`
- `services.WishlistServiceInterface`
- `services.MaterialResolverInterface`

When modifying services or handlers, ensure interface compliance is maintained.

## API Endpoints

### Public
- `GET /health` - Health check
- `GET /api/v1/items/search` - Search items
- `GET /api/v1/items/{uniqueName}` - Get item details

### Protected (requires JWT)
- `GET /api/v1/wishlist` - Get user's wishlist
- `POST /api/v1/wishlist` - Add item to wishlist
- `DELETE /api/v1/wishlist/{uniqueName}` - Remove item
- `PATCH /api/v1/wishlist/{uniqueName}` - Update quantity
- `GET /api/v1/wishlist/materials` - Get aggregated materials

## Environment Variables

```
SERVER_PORT=8080
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=warframe
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_JWT_SECRET=your-jwt-secret
ALLOWED_ORIGINS=http://localhost:3000
```
