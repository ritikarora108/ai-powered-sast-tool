# AI-Powered SAST Tool

An AI-powered Static Application Security Testing (SAST) tool designed to scan public GitHub repositories for OWASP Top 10 vulnerabilities.

## Features

- Scan public GitHub repositories for security vulnerabilities
- AI-powered vulnerability detection using OpenAI via BAML
- Clean UI to visualize and categorize security findings
- Integration with Google Sign-In for authentication
- Detailed vulnerability reports with remediation suggestions

## Tech Stack

### Frontend

- Next.js (React)
- Tailwind CSS/shadcn-ui
- Google Sign-In for authentication

### Backend

- Golang
- Chi for HTTP routing
- Temporal for workflow orchestration
- BAML for AI orchestration
- OpenAI for vulnerability detection
- PostgreSQL with sqlc for type-safe database access

### DevOps

- Docker and Docker Compose for containerization
- GitHub Actions for CI/CD (coming soon)

## Getting Started

### Prerequisites

- Docker and Docker Compose
- OpenAI API key
- Go 1.22+
- Node.js 20+
- PostgreSQL (if running locally)

### Setup

1. Clone the repository

```bash
git clone https://github.com/your-username/ai-powered-sast-tool.git
cd ai-powered-sast-tool
```

2. Create a `.env` file in the root directory with the following variables:

```
OPENAI_API_KEY=your_openai_api_key
```

3. Start the application with Docker Compose:

```bash
docker-compose -f deploy/docker-compose.yml up -d
```

4. Access the application:

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Temporal UI: http://localhost:8088

### Local Development

#### Backend

1. Change to the backend directory:

```bash
cd backend
```

2. Install dependencies:

```bash
go mod download
```

3. Run the application:

```bash
go run main.go
```

#### Frontend

1. Change to the frontend directory:

```bash
cd frontend
```

2. Install dependencies:

```bash
npm install
```

3. Run the development server:

```bash
npm run dev
```

## Project Structure

```
ai-powered-sast-tool/
├── docs/                 # Technical documentation, PRD, specs, diagrams, AI prompts
├── deploy/               # Docker, docker-compose, and deployment configs
├── frontend/             # Next.js React frontend
├── backend/              # Golang backend with Temporal, sqlc, and OpenAI
│   ├── api/              # API definitions and handlers
│   ├── internal/         # Internal packages
│   ├── db/               # Database access layer (sqlc)
│   ├── temporal/         # Temporal workflows and activities
│   ├── services/         # Business logic and services
│   ├── handlers/         # HTTP handlers
│   └── baml/             # BAML configuration and client
├── scripts/              # Utility shell/python scripts
└── .gitignore
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
