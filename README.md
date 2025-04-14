# AI-Powered SAST Tool

A lightweight Static Application Security Testing (SAST) tool that uses AI to detect OWASP Top 10 vulnerabilities in codebases from public GitHub repositories.

## Key Features

- Clone and analyze public GitHub repositories
- Detect and categorize vulnerabilities by OWASP Top 10
- User authentication via Google Sign-In
- Scalable architecture with Temporal for workflow management
- AI-powered code analysis using OpenAI models

## Architecture

This application consists of:

- **Frontend**: Next.js with TypeScript, Tailwind CSS, and Next Auth
- **Backend**: Golang API server with BAML for AI orchestration
- **Database**: PostgreSQL with sqlc for type-safe DB access
- **Workflow Engine**: Temporal for reliable execution of long-running tasks
- **AI Integration**: OpenAI for code analysis

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Node.js 18+ (for local development)
- Go 1.20+ (for local development)
- PostgreSQL 15 (for local development)
- OpenAI API key
- Google OAuth credentials

### Environment Setup

1. Clone this repository:

   ```bash
   git clone https://github.com/yourusername/ai-powered-sast-tool.git
   cd ai-powered-sast-tool
   ```

2. Create a `.env` file in the root directory:

   ```
   GOOGLE_CLIENT_ID=your-google-client-id
   GOOGLE_CLIENT_SECRET=your-google-client-secret
   OPENAI_API_KEY=your-openai-api-key
   ```

3. Run with Docker Compose:

   ```bash
   docker-compose up
   ```

4. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

### Local Development

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

#### Backend

```bash
cd backend
go mod download
go run main.go
```

## Usage

1. Sign in with your Google account
2. Add a GitHub repository URL to scan
3. View scan results categorized by OWASP Top 10 vulnerabilities
4. Review detailed findings and recommendations

## User Flows

### Authentication

- Google Sign-In to protect all routes
- No user management or admin panel needed

### Scanning Workflow

1. User submits a GitHub repo URL
2. Backend clones the repo and analyzes the codebase for OWASP Top 10 vulnerabilities using AI
3. Results are persisted to PostgreSQL
4. Results are displayed categorized by OWASP Top 10 on the frontend

### UI Views

- Home Page
- "Add GitHub Repo" form
- Table of scanned repos with status
- Repo Detail View with findings organized by OWASP Top 10 category

## Project Structure

```
ai-powered-sast-tool/
├── backend/                    # Go backend
│   ├── api/                    # API routes and handlers
│   │   └── middleware/         # API middleware
│   ├── baml/                   # BAML AI configuration
│   │   └── code_scanner.baml   # BAML prompts
│   ├── db/                     # Database access
│   │   ├── migrations/         # SQL migrations
│   │   ├── query/              # SQL queries
│   │   ├── schema/             # Database schema
│   │   └── sqlc/               # Generated database code
│   ├── handlers/               # API handlers
│   │   ├── repository.go       # Repository handlers
│   │   ├── auth.go             # Authentication handlers
│   │   └── user.go             # User handlers
│   ├── internal/               # Internal packages
│   │   └── logger/             # Logging utilities
│   ├── services/               # Business logic
│   │   ├── github.go           # GitHub API integration
│   │   ├── auth.go             # Authentication service
│   │   ├── scanner.go          # Vulnerability scanner
│   │   └── openai.go           # OpenAI integration
│   ├── temporal/               # Temporal workflows
│   │   ├── activities.go       # Workflow activities
│   │   └── workflows.go        # Workflow definitions
│   ├── scripts/                # Backend utility scripts
│   ├── .env                    # Environment variables
│   ├── go.mod                  # Go dependencies
│   ├── go.sum                  # Go dependency checksums
│   └── main.go                 # Entry point
├── frontend/                   # Next.js frontend
│   ├── public/                 # Static assets
│   ├── src/                    # Source code
│   │   ├── app/                # Next.js app router
│   │   ├── components/         # React components
│   │   ├── services/           # API services
│   │   └── types/              # TypeScript types
│   ├── .env.local              # Environment variables
│   ├── Dockerfile              # Frontend Docker config
│   ├── next.config.ts          # Next.js configuration
│   ├── package.json            # NPM dependencies
│   └── tsconfig.json           # TypeScript configuration
├── deploy/                     # Deployment configuration
│   ├── docker-compose.yml      # Docker Compose config
│   ├── Dockerfile.backend      # Backend Docker config
│   └── Dockerfile.frontend     # Frontend Docker config
├── docs/                       # Documentation
│   ├── PRD.md                  # Product Requirements Document
│   └── Technical_Spec.md       # Technical Specification
├── scripts/                    # Utility scripts
│   ├── test_scan.sh            # Scan testing script
│   ├── setup_db.sh             # Database setup script
│   └── generate_sqlc.sh        # SQL code generation
├── .env.sample                 # Sample environment variables
├── .gitignore                  # Git ignore file
└── LICENSE                     # License information
```

## License

Copyright © 2025 Ritik Arora. All rights reserved.
