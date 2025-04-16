# KeyGraph: AI-Powered Static Application Security Testing Tool

This repository contains an AI-powered Static Application Security Testing (SAST) tool that identifies security vulnerabilities in your code by analyzing GitHub repositories.

## Overview

KeyGraph scans repositories for OWASP Top 10 vulnerabilities using AI-powered analysis, providing detailed reports and recommendations to improve your code's security posture.

## Key Features

- Clone and analyze public GitHub repositories
- Detect and categorize vulnerabilities by OWASP Top 10
- User authentication via Google Sign-In
- Scalable architecture with Temporal for workflow management
- AI-powered code analysis using OpenAI models

## Project Structure

The project is organized as follows:

```
ai-powered-sast-tool/
├── frontend/                   # Next.js frontend application
│   ├── public/                 # Static assets
│   ├── src/                    # Source code
│   │   ├── app/                # Next.js App Router pages
│   │   ├── components/         # React components
│   │   ├── context/            # React contexts
│   │   ├── hooks/              # Custom React hooks
│   │   ├── services/           # API services
│   │   ├── utils/              # Utility functions
│   │   ├── constants/          # Constants and configuration
│   │   └── types/              # TypeScript type definitions
│   ├── .env.example            # Example environment variables
│   └── README.md               # Frontend-specific documentation
├── backend/                    # Go backend application
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
│   ├── internal/               # Internal packages
│   │   └── logger/             # Logging utilities
│   ├── services/               # Business logic
│   ├── temporal/               # Temporal workflows
│   │   ├── activities.go       # Workflow activities
│   │   └── workflows.go        # Workflow definitions
│   ├── scripts/                # Backend utility scripts
│   ├── .env.example            # Example environment variables
│   └── README.md               # Backend-specific documentation
├── scripts/                    # Utility scripts
├── deploy/                     # Deployment configuration
├── docs/                       # Documentation
└── .vscode/                    # VS Code configuration
```

## Architecture

This application consists of:

- **Frontend**: Next.js with TypeScript, Tailwind CSS, and NextAuth.js
- **Backend**: Golang API server with BAML for AI orchestration
- **Database**: PostgreSQL with sqlc for type-safe DB access
- **Workflow Engine**: Temporal for reliable execution of long-running tasks
- **AI Integration**: OpenAI for code analysis

## Getting Started

### Prerequisites

- Node.js 18+ (for frontend)
- Go 1.18+ (for backend)
- PostgreSQL
- Temporal server
- Google OAuth credentials
- OpenAI API key

### Quick Start

1. Clone this repository:

   ```bash
   git clone https://github.com/yourusername/ai-powered-sast-tool.git
   cd ai-powered-sast-tool
   ```

2. Set up the backend:

   ```bash
   cd backend
   cp .env.example .env
   # Edit .env with your configuration
   go mod download
   ```

3. Set up the frontend:

   ```bash
   cd frontend
   cp .env.example .env.local
   # Edit .env.local with your configuration
   npm install
   ```

4. Start the backend (using PowerShell):

   ```powershell
   ./run-backend.ps1
   ```

5. Start the frontend (using PowerShell):

   ```powershell
   ./run-frontend.ps1
   ```

6. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

### Setting up the database

Ensure PostgreSQL is installed and running, then:

```bash
cd backend/scripts
# Run the database setup script for your platform
```

## User Flow

1. Sign in with your Google account
2. Add a GitHub repository URL to scan
3. View scan results categorized by OWASP Top 10 vulnerabilities
4. Review detailed findings and recommendations

## Development

For detailed development instructions, see:

- [Frontend README](./frontend/README.md)
- [Backend README](./backend/README.md)

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
