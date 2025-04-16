'use client';

import { Inter } from 'next/font/google';
import './globals.css';

const inter = Inter({
    subsets: ['latin'],
    variable: '--font-inter',
    display: 'swap',
});

export default function GlobalError({
    error,
    reset,
}: {
    error: Error & { digest?: string };
    reset: () => void;
}) {
    return (
        <html lang="en" className={`${inter.variable} scroll-smooth antialiased`}>
            <body className="min-h-screen bg-background text-foreground">
                <div className="min-h-screen flex items-center justify-center p-4">
                    <div className="w-full max-w-md">
                        <div className="rounded-lg bg-white dark:bg-gray-800 p-6 shadow-lg">
                            <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                                Something went wrong!
                            </h2>
                            <div className="mt-4">
                                <p className="text-sm text-gray-600 dark:text-gray-400">
                                    The application encountered a critical error.
                                </p>
                                {process.env.NODE_ENV === 'development' && (
                                    <div className="mt-2 p-2 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 text-xs rounded overflow-auto max-h-40">
                                        {error.message}
                                    </div>
                                )}
                            </div>
                            <div className="mt-6">
                                <button
                                    onClick={() => reset()}
                                    className="inline-flex items-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                                >
                                    Try again
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </body>
        </html>
    );
}
