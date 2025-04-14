# Keygraph SAST Tool Frontend

This is the frontend component of the Keygraph AI-powered Static Application Security Testing (SAST) tool. It provides a user interface for scanning GitHub repositories for OWASP Top 10 vulnerabilities.

## Features

- Google Sign-In authentication
- Repository submission and scanning
- Viewing scan results categorized by OWASP Top 10
- Responsive design for desktop and mobile

## Technologies Used

- Next.js 14 with App Router
- TypeScript
- Tailwind CSS
- Next Auth for authentication
- Axios for API requests
- Headless UI for components

## Setup and Installation

1. Clone the repository
2. Install dependencies:
   ```
   npm install
   ```
3. Create a `.env.local` file with the following variables:
   ```
   NEXT_PUBLIC_API_URL=http://localhost:8080
   NEXTAUTH_URL=http://localhost:3000
   NEXTAUTH_SECRET=your-secret-key
   GOOGLE_CLIENT_ID=your-google-client-id
   GOOGLE_CLIENT_SECRET=your-google-client-secret
   ```
4. Start the development server:
   ```
   npm run dev
   ```

## Usage

1. Navigate to `http://localhost:3000`
2. Sign in with your Google account
3. Add a GitHub repository URL to scan
4. View scan results and OWASP vulnerability details

## Integration with Backend

This frontend is designed to work with the Keygraph SAST Tool backend. The backend handles:

- Repository cloning
- Code scanning using OpenAI
- OWASP Top 10 vulnerability detection
- Temporal workflow orchestration

Make sure the backend is running on the URL specified in your `.env.local` file.

## License

Copyright © 2023 Keygraph. All rights reserved.
