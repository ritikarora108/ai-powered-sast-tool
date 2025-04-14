import Link from 'next/link';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'AI-powered SAST Tool',
  description: 'Detect OWASP Top 10 vulnerabilities in your codebase with AI',
};

export default function Home() {
  return (
    <div className="flex flex-col min-h-screen">
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex justify-between items-center">
          <h1 className="text-3xl font-bold text-gray-900">AI-powered SAST Tool</h1>
          <div>
            <Link 
              href="/auth/signin"
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Sign In with Google
            </Link>
          </div>
        </div>
      </header>

      <main className="flex-grow">
        <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h2 className="text-base font-semibold text-indigo-600 tracking-wide uppercase">Company_Name</h2>
            <p className="mt-1 text-4xl font-extrabold text-gray-900 sm:text-5xl sm:tracking-tight lg:text-6xl">
              AI-powered SAST Tool
            </p>
            <p className="max-w-xl mt-5 mx-auto text-xl text-gray-500">
              Detect OWASP Top 10 vulnerabilities in your codebase with artificial intelligence.
            </p>
          </div>

          <div className="mt-10">
            <div className="max-w-xl mx-auto">
              <div className="grid grid-cols-1 gap-8 sm:grid-cols-2">
                <div className="bg-white overflow-hidden shadow rounded-lg">
                  <div className="px-4 py-5 sm:p-6">
                    <h3 className="text-lg font-medium text-gray-900">GitHub Integration</h3>
                    <p className="mt-2 text-sm text-gray-500">
                      Analyze public repositories directly from GitHub with a simple URL.
                    </p>
                  </div>
                </div>
                <div className="bg-white overflow-hidden shadow rounded-lg">
                  <div className="px-4 py-5 sm:p-6">
                    <h3 className="text-lg font-medium text-gray-900">OWASP Top 10</h3>
                    <p className="mt-2 text-sm text-gray-500">
                      Detect vulnerabilities categorized by OWASP Top 10 security risks.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      <footer className="bg-white">
        <div className="max-w-7xl mx-auto py-12 px-4 overflow-hidden sm:px-6 lg:px-8">
          <p className="text-center text-base text-gray-500">
            &copy; 2025 Ritik Arora. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}
