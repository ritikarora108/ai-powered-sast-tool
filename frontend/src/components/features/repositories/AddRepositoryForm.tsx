'use client';

import { useState } from 'react';
import { PlusIcon } from '@heroicons/react/24/outline';

interface AddRepositoryFormProps {
    onSubmit: (url: string) => Promise<void>;
    isSubmitting?: boolean;
}

export default function AddRepositoryForm({
    onSubmit,
    isSubmitting = false,
}: AddRepositoryFormProps) {
    const [repoUrl, setRepoUrl] = useState('');
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        // Basic validation
        if (!repoUrl.trim()) {
            setError('Repository URL is required');
            return;
        }

        // Validate GitHub URL format
        if (!isValidGitHubUrl(repoUrl)) {
            setError('Please enter a valid GitHub repository URL');
            return;
        }

        setError(null);

        try {
            await onSubmit(repoUrl);
            setRepoUrl(''); // Clear form after successful submission
        } catch (err: any) {
            setError(err.message || 'Failed to add repository');
        }
    };

    // Function to validate GitHub URL
    const isValidGitHubUrl = (url: string): boolean => {
        try {
            // Add https:// prefix if missing
            let urlToCheck = url;
            if (!urlToCheck.startsWith('http://') && !urlToCheck.startsWith('https://')) {
                urlToCheck = 'https://' + urlToCheck;
            }

            const parsed = new URL(urlToCheck);
            return (
                parsed.hostname === 'github.com' &&
                parsed.pathname.split('/').filter(Boolean).length >= 2
            );
        } catch (e) {
            return false;
        }
    };

    return (
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                Add Repository for Analysis
            </h2>

            {error && (
                <div className="mb-4 rounded-md bg-red-50 p-4 text-sm text-red-700">{error}</div>
            )}

            <form onSubmit={handleSubmit}>
                <div className="flex flex-col md:flex-row gap-4">
                    <div className="flex-grow">
                        <label htmlFor="repoUrl" className="sr-only">
                            GitHub Repository URL
                        </label>
                        <input
                            type="text"
                            id="repoUrl"
                            className="block w-full rounded-md border-gray-300 dark:border-gray-600 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-gray-700 dark:text-white text-sm"
                            placeholder="https://github.com/username/repository"
                            value={repoUrl}
                            onChange={e => setRepoUrl(e.target.value)}
                            disabled={isSubmitting}
                            required
                        />
                    </div>
                    <button
                        type="submit"
                        disabled={isSubmitting}
                        className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {isSubmitting ? (
                            <>
                                <svg
                                    className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
                                    xmlns="http://www.w3.org/2000/svg"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                >
                                    <circle
                                        className="opacity-25"
                                        cx="12"
                                        cy="12"
                                        r="10"
                                        stroke="currentColor"
                                        strokeWidth="4"
                                    ></circle>
                                    <path
                                        className="opacity-75"
                                        fill="currentColor"
                                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                    ></path>
                                </svg>
                                Submitting...
                            </>
                        ) : (
                            <>
                                <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                                Add Repository
                            </>
                        )}
                    </button>
                </div>
                <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
                    Enter the URL of a public GitHub repository to scan for OWASP Top 10
                    vulnerabilities. You will be notified by email when the scan is complete.
                </p>
            </form>
        </div>
    );
}
