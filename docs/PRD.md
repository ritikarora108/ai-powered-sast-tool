# 📄 Product Requirements Document (PRD)

**Product Name:** AI-Powered SAST Tool for OWASP Top 10 Vulnerability Detection

**Document Version:** v1.0

**Author:** Ritik (Product + Full Stack Engineer Candidate)

**Date:** April 12, 2025

---

## 📚 Table of Contents

- [Overview](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Purpose and Vision](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Target Audience and User Personas](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Problem Statement](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Objectives and Success Metrics](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Scope](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Assumptions, Dependencies, and Constraints](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Feature Requirements & User Stories](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Technical Requirements](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Non-Functional Requirements](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [User Flow and Use Cases](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Risks and Mitigation Strategies](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Timeline and Milestones](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)
- [Appendix & Open Issues](https://www.notion.so/AI-Powered-SAST-Tool-for-OWASP-Top-10-Vulnerability-Detection-1d2529fcfc65806c9950f68de272430b?pvs=21)

---

## 1. 🧾 Overview

This PRD outlines the development of an AI-driven SAST tool that scans public GitHub repositories to identify vulnerabilities grouped by the OWASP Top 10. The tool combines static code analysis techniques with AI-assisted vulnerability detection using Temporal, BAML, and OpenAI’s LLMs.

---

## 2. 🎯 Purpose and Vision

**Purpose**

Develop a production-grade web app that enables users to submit GitHub URLs, scan code for vulnerabilities using AI, and categorize them as per OWASP Top 10.

**Vision**

- Empower developers and security engineers with an intuitive, accurate, and fast vulnerability scanning tool.
- Automate code analysis and security classification.
- Serve as a foundational security automation platform.

---

## 3. 👥 Target Audience and User Personas

**Primary Users**

- Developer Security Engineers
- DevOps and Security Teams
- Open-Source Contributors

**User Persona Example**

**Name:** Alex, Security Engineer

**Needs:** Rapid vulnerability insights; CI/CD integration

**Pain Points:** Manual static analysis is slow and error-prone

---

## 4. 🚨 Problem Statement

- Manual reviews are time-consuming.
- Traditional SAST tools lack accuracy in large codebases.
- UX of current tools doesn’t fit into developer workflows.

---

## 5. 📏 Objectives and Success Metrics

**Objectives**

- Automate scanning of GitHub repos
- AI-based OWASP classification
- Dashboard with remediation suggestions
- Secure, scalable architecture

**Success Metrics**

- 🧑‍💻 User Adoption: Weekly/monthly active scans
- 🎯 Detection Accuracy: Benchmark-matched rates
- ⚡ Performance: <60s average scan time
- 👍 User Satisfaction: Positive feedback
- ⏱️ Uptime: 99.9%

---

## 6. 📦 Scope

**In-Scope**

- Google Sign-In
- GitHub Repo Submission
- Cloning + Analysis via Temporal
- AI detection via BAML/OpenAI
- PostgreSQL dashboard
- Dockerized setup

**Out-of-Scope**

- Private repos
- Manual vulnerability classification
- Key revocation (future scope)

---

## 7. ⚙️ Assumptions, Dependencies, and Constraints

**Assumptions**

- Repositories are public and accessible.
- OpenAI API keys are functional.
- Modern browsers are used.

**Dependencies**

- Temporal
- BAML/OpenAI
- Google Auth
- PostgreSQL + sqlc
- Docker

**Constraints**

- Large repo limits
- External API latency
- GDPR + HTTPS compliance

---

## 8. 🧩 Feature Requirements & User Stories

### 8.1 Authentication

**Feature:** Google Sign-In

**User Story:**

> As a developer, I want to securely sign in so I can view past scans without compromising security.
> 

---

### 8.2 Repository Submission

**Feature:** GitHub Repo Form

**User Story:**

> As a user, I want to submit a GitHub URL so the system can scan it automatically.
> 

---

### 8.3 Repository Cloning & Analysis

**Feature:** Backend Cloning Service

**User Story:**

> As a system, I want to clone the repo and extract code files for AI scanning.
> 

---

### 8.4 AI-Powered Vulnerability Scanning

**Feature:** BAML/OpenAI Integration

**User Story:**

> As a security engineer, I want automatic detection of OWASP Top 10 vulnerabilities.
> 

---

### 8.5 Results Dashboard

**Feature:** UI with categorized results

**User Story:**

> As a user, I want to view vulnerabilities by severity and category to fix them efficiently.
> 

---

## 9. 🔧 Technical Requirements

**Frontend**

- React / Next.js
- Google Sign-In
- Responsive dashboard

**Backend**

- Golang
- Temporal workflows
- AI via BAML/OpenAI
- go-git for cloning
- PostgreSQL + sqlc
- REST APIs

**DevOps**

- Docker (Dev + Prod)
- Docker Compose/K8s
- HTTPS, secure APIs, rate limiting

---

## 10. 📐 Non-Functional Requirements

- **Scalability:** Handle concurrent scans
- **Performance:** <60s scan time
- **Reliability:** 99.9% uptime
- **Maintainability:** Modular, documented codebase
- **Security:** HTTPS, API key protection, input validation

---

## 11. 🧭 User Flow and Use Cases

**Flow**

1. User signs in via Google
2. Submits public GitHub URL
3. Backend clones repo and starts workflow
4. AI scans code via BAML/OpenAI
5. Results saved to DB
6. User views categorized vulnerabilities

🧠 *Use Case Diagram:* *(To be included in Notion with image block.)*

---

## 12. ⚠️ Risks and Mitigation Strategies

| Risk | Mitigation |
| --- | --- |
| Large Repos | Limit file size, paginate |
| API Rate Limits | Add retries, cache responses |
| Security Vulnerabilities | Secure input, API key management |
| Overwhelming UI | User testing, iteration |

---

## 13. ⏳ Timeline and Milestones

| Milestone | Duration | Description |
| --- | --- | --- |
| MVP | Days 1–2 | Auth + repo scan setup |
| AI Integration | Day 3 | BAML/OpenAI |
| Dashboard UI | Day 4 | Results view |
| Testing | Day 5 | QA + security |
| Final Review | Day 6 | Docs + polish |

---

## 14. 📎 Appendix & Open Issues

**Open Issues**

- Finalize AI prompt structure
- File type support scope
- Robust error handling
- Key revocation strategies (future)

**To Be Added**

- Architecture diagrams
- REST API specs
- UI wireframes

---

## ✅ Conclusion

This PRD outlines a full-stack AI-powered SAST tool to help users detect OWASP Top 10 vulnerabilities in public repositories. It emphasizes automation, UX, security, and scalability — serving as a launchpad for continuous security integration in developer workflows.