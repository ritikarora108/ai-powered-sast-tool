'use client';

import { useEffect } from 'react';
import { ExclamationCircleIcon } from '@heroicons/react/24/outline';

export default function Error({
    error,
    reset,
}: {
    error: Error & { digest?: string };
    reset: () => void;
}) {
    useEffect(() => {
        // Log the error to an error reporting service
        console.error('Application error:', error);
    }, [error]);

    return (
        <div className="min-h-screen flex items-center justify-center p-4">
            <div className="w-full max-w-md">
                <div className="rounded-lg bg-white dark:bg-gray-800 p-6 shadow-lg">
                    <div className="flex items-center">
                        <div className="flex-shrink-0">
                            <ExclamationCircleIcon
                                className="h-6 w-6 text-red-500"
                                aria-hidden="true"
                            />
                        </div>
                        <h2 className="ml-3 text-lg font-medium text-gray-900 dark:text-gray-100">
                            Something went wrong
                        </h2>
                    </div>
                    <div className="mt-4">
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                            An error occurred while loading this page. You can try again by
                            refreshing.
                        </p>
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
    );
}
