# KeyGraph AI-Powered SAST Tool Frontend

This is the frontend implementation for the KeyGraph AI-Powered Static Application Security Testing (SAST) tool. It's built with Next.js and designed to work with the Go backend.

## Features

- **Google Sign-In Authentication**: Secure all routes with Google authentication
- **Repository Management**: Add public GitHub repositories for scanning
- **Security Analysis**: View detailed security scan results organized by OWASP Top 10 categories
- **Modern UI**: Clean, responsive user interface with dark mode support

## Technology Stack

- **Framework**: [Next.js](https://nextjs.org/)
- **Authentication**: [NextAuth.js](https://next-auth.js.org/) with Google Provider
- **Styling**: [Tailwind CSS](https://tailwindcss.com/)
- **Icons**: [Heroicons](https://heroicons.com/)
- **HTTP Client**: [Axios](https://axios-http.com/)

## Getting Started

### Prerequisites

- Node.js 18+ and npm/yarn
- Google OAuth credentials (Client ID and Client Secret)
- Running backend server (see the backend README)

### Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/yourusername/ai-powered-sast-tool.git
    cd ai-powered-sast-tool/frontend
    ```

2. Install dependencies:

    ```bash
    npm install
    # or
    yarn install
    ```

3. Set up environment variables:

    ```bash
    cp .env.example .env.local
    ```

    Then edit `.env.local` to add your Google OAuth credentials and other required values.

4. Start the development server:

    ```bash
    npm run dev
    # or
    yarn dev
    ```

5. Open [http://localhost:3000](http://localhost:3000) to view the application.

## Deployment

### Build for Production

```bash
npm run build
# or
yarn build
```

### Start Production Server

```bash
npm run start
# or
yarn start
```

## Project Structure

```
frontend/
├── public/           # Static assets
├── src/
│   ├── app/          # Next.js App Router pages
│   │   ├── (auth)/   # Authentication pages
│   │   ├── (dashboard)/ # Protected dashboard pages
│   │   ├── api/      # API routes
│   │   ├── components/   # React components
│   │   │   ├── features/ # Feature components
│   │   │   ├── ui/       # UI components
│   │   ├── context/      # React contexts
│   │   ├── hooks/        # Custom React hooks
│   │   ├── services/     # API services
│   │   ├── types/        # TypeScript types
│   │   └── utils/        # Utility functions
│   ├── .env.example      # Example environment variables
│   ├── next.config.mjs   # Next.js configuration
│   └── tailwind.config.js # Tailwind CSS configuration
```

## Authentication Flow

1. Users visit the application and are redirected to the sign-in page
2. Users authenticate with Google
3. NextAuth.js handles the OAuth flow and creates a session
4. The session is used to authenticate API requests to the backend
5. Protected routes check for an active session before rendering

## Contributing

Please read [CONTRIBUTING.md](../CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.
