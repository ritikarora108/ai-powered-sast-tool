# AI-Powered SAST Tool Backend

This is the backend for an AI-powered Static Application Security Testing (SAST) tool that analyzes GitHub repositories for OWASP Top 10 vulnerabilities.

## Features

- Authenticate users with Google Sign-In
- Clone and analyze GitHub repositories
- Detect OWASP Top 10 vulnerabilities using AI
- Store results in PostgreSQL database
- Use Temporal for workflow orchestration
- Provide API endpoints for frontend integration

## Tech Stack

- Golang
- Temporal
- PostgreSQL
- OpenAI
- BAML (AI LLM orchestration)
- Google OAuth2 API

## Prerequisites

- Go 1.18 or later
- PostgreSQL
- Temporal server
- Google OAuth credentials
- OpenAI API key

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
# Temporal Configuration
TEMPORAL_HOST=localhost:7233

# GitHub Configuration (optional for testing)
GITHUB_TOKEN=your_github_token

# PostgreSQL Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=sast
DB_URL="postgres://postgres:your_password@localhost:5432/sast?sslmode=disable"

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

# JWT Configuration
JWT_SECRET=your_jwt_secret

# OpenAI Configuration
OPENAI_API_KEY=your_openai_api_key

# Logging Configuration
LOG_LEVEL=debug

# Environment
APP_ENV=development

# Frontend URL (for CORS)
FRONTEND_URL=http://localhost:3000
```

## Setup Database

1. Ensure PostgreSQL is installed and running.

2. Use the provided setup scripts to initialize the database:

On Linux/macOS:

```bash
cd scripts
chmod +x setup_db.sh
./setup_db.sh
```

On Windows:

```powershell
cd scripts
.\setup_db.ps1
```

The scripts will:

- Create the `sast` database
- Apply the database schema
- Set up necessary extensions

Alternatively, you can manually create the database:

```sql
CREATE DATABASE sast;
```

Then run the schema file:

```bash
psql -U postgres -d sast -f db/schema/schema.sql
```

## Build and Run

```bash
# Build the application
go build

# Run the application
./backend
```

The server will start on port 8080 by default. You can change this by setting the `PORT` environment variable.

## API Endpoints

### Authentication

- `GET /auth/google` - Redirects to Google Sign-In
- `GET /auth/google/callback` - Callback URL for Google Sign-In

### Public Endpoints

- `GET /health` - Health check endpoint
- `POST /scan` - Scan a public GitHub repository
- `GET /scan/{id}/status` - Get scan status
- `GET /scan/{id}/results` - Get scan results
- `GET /scan/{id}/debug` - Debug a scan workflow

### Protected Endpoints (require authentication)

- `POST /api/repositories` - Create a new repository
- `GET /api/repositories` - List repositories
- `GET /api/repositories/{id}` - Get repository details
- `POST /api/repositories/{id}/scan` - Scan a repository
- `GET /api/repositories/{id}/vulnerabilities` - Get vulnerabilities for a repository
- `GET /api/users/me` - Get authenticated user profile

## Frontend Integration

The backend provides all necessary API endpoints for frontend integration. The frontend can authenticate users via Google Sign-In and then use the protected API endpoints to interact with the application.

## Development

To add a new feature or fix a bug:

1. Create a new branch
2. Make your changes
3. Test your changes
4. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
