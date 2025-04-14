## 📌 Overview

**🛠 Product**: AI‑Powered SAST Tool for OWASP Top 10 Detection

**🎯 Purpose**: Allow users to submit public GitHub repo URLs and receive categorized AI-powered static analysis.

**🏁 Goal**: Build an end-to-end SAST solution that integrates into dev workflows and improves vulnerability detection + remediation.

---

## 🏗 Architectural Overview

**System Design**: Modular, service-oriented architecture with separation between frontend, backend, and database. Fully containerized via Docker.

### 🔹 Core Components:

### 🔸 Frontend (React/Next.js)

- Google Sign‑In for authentication
- Views: repo submission, status list, vulnerability results

### 🔸 Backend (Golang)

- RESTful APIs for repo submission and results
- Uses **Temporal** for scanning orchestration
- AI vulnerability detection via **BAML + OpenAI**
- GitHub repo cloning and code extraction modules

### 🔸 Database (PostgreSQL + sqlc)

- User data
- Scan metadata
- Vulnerability reports

### 🔸 DevOps

- Docker & Docker Compose for local + production deployment

---

## ⚙️ Technology Stack & Rationale

| Component | Tech | Why |
| --- | --- | --- |
| Frontend | React/Next.js | SEO, fast SPAs |
| Auth | Google Sign-In | Simple OAuth2 |
| Backend | Golang | Performance, concurrency |
| Workflow Engine | Temporal | Fault-tolerance, retries |
| AI Orchestration | BAML + OpenAI | Accurate static analysis |
| DB | PostgreSQL + sqlc | Type-safe, scalable |
| DevOps | Docker | Environment consistency |

---

## 🔍 Component Details

### A. 🔐 User Auth + Repo Submission (Frontend)

- **Google OAuth2** protects routes.
- **Submission form** accepts public GitHub repo URLs → POST to API.

---

### B. 🧠 Repository Cloning & Analysis (Backend)

- **Clone Module**: Uses `go-git`
- **File Filter**: Scans only relevant code files (`.js`, `.py`, `.go`, etc.)
- **Temporal Workflow**:
    - `CloneRepo → ExtractFiles → AIAnalysis → SaveResults`

---

### C. 📊 Results Dashboard (Frontend)

- Table view of submitted repos with scan status.
- Detail view with OWASP Top 10 grouping, file path, severity, line number, and remediation hints.

---

### D. 🧾 Database Schema (Simplified ERD)

| Table | Fields |
| --- | --- |
| `users` | id, email, auth provider |
| `repositories` | id, user_id (FK), URL, status, timestamp |
| `vulnerabilities` | id, repo_id (FK), OWASP category, severity, file path, snippet, fix |

---

## 🧭 User Workflow Diagram

![User_Flow.png](attachment:7503ada6-bbfe-47c1-9102-f99919ac993d:User_Flow.png)

---

## 🚀 DevOps & Deployment

### 🔧 Docker Setup

```yaml

version: '3.8'
services:
  frontend:
    build: ./frontend
    ports: ["3000:3000"]
    environment: [NODE_ENV=production]
    depends_on: [backend]

  backend:
    build: ./backend
    ports: ["8080:8080"]
    environment:
      - DB_HOST=postgres
      - DB_USER=youruser
      - DB_PASSWORD=yourpassword
      - OPENAI_API_KEY=yourkey
    depends_on: [postgres]

  postgres:
    image: postgres:14
    environment:
      - POSTGRES_USER=youruser
      - POSTGRES_PASSWORD=yourpassword
      - POSTGRES_DB=yourdb
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:

```

### ⚙️ CI/CD

- GitHub Actions to test, build Docker images
- Dockerized environment = smooth local ↔ prod flow

---

## 🧰 Tool Usage Summary

| Tool | Why |
| --- | --- |
| Temporal | Async orchestration, fault-tolerant |
| BAML + OpenAI | Precise AI-driven code analysis |
| sqlc | Type-safe DB layer |
| Docker | Environment consistency |

---

## 📈 Considerations & Next Steps

- **Prompt Engineering**: Iterate BAML prompts for better AI detection.
- **Scalability**: Add autoscaling for backend.
- **Error Handling**: Improve Temporal & API error fallback logic.
- **User Metrics**: Track false positives, time to scan, UX flow.

---

## ✅ Conclusion

This spec outlines a scalable, production-grade blueprint that:

- Uses AI to detect vulnerabilities automatically.
- Leverages Temporal + OpenAI + BAML for async workflows + AI insights.
- Containerized and type-safe across the stack.
- Provides a clean UI for end-users.