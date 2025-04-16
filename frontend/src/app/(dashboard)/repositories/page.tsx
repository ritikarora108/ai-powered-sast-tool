'use client';

import React, { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import { repositoryApi } from '@/services/api';
import { Repository } from '@/types';
import Link from 'next/link';
import { DashboardLayout } from '@/components';
import {
    CheckCircleIcon,
    ExclamationCircleIcon,
    ClockIcon,
    ArrowPathIcon,
    FolderIcon,
    MagnifyingGlassIcon,
    ChevronUpDownIcon,
    ChevronDownIcon,
    ChevronUpIcon,
    EyeIcon,
    ArrowPathRoundedSquareIcon,
    ShieldCheckIcon,
    CodeBracketIcon,
} from '@heroicons/react/24/outline';

type SortKey = 'name' | 'status' | 'last_scan_at';
type SortDirection = 'asc' | 'desc';

export default function RepositoriesPage() {
    const { data: session, status } = useSession();
    const router = useRouter();
    const [repositories, setRepositories] = useState<Repository[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [refreshing, setRefreshing] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [sortKey, setSortKey] = useState<SortKey>('last_scan_at');
    const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
    const [filterStatus, setFilterStatus] = useState<string | null>(null);

    useEffect(() => {
        if (status === 'unauthenticated') {
            router.push('/auth/signin');
        }
    }, [status, router]);

    useEffect(() => {
        if (status === 'authenticated') {
            loadRepositories();
        }
    }, [status]);

    const loadRepositories = async () => {
        setIsLoading(true);
        setError(null);

        try {
            console.log('Attempting to load repositories...');
            const response = await repositoryApi.getRepositories();
            console.log('Repository API call response:', response);

            if (response.data) {
                // Make sure response.data is an array
                if (Array.isArray(response.data)) {
                    setRepositories(response.data);
                    console.log(`Loaded ${response.data.length} repositories`, response.data);
                } else {
                    console.error('Repository data is not an array:', response.data);
                    setError('Invalid data format received from server');
                    setRepositories([]);
                }
            } else if (response.error) {
                console.error('API returned error:', response.error);
                setError(
                    typeof response.error === 'string'
                        ? response.error
                        : response.error.message || 'Unknown error'
                );
                setRepositories([]);
            } else {
                console.log('Empty response received, setting empty repositories array');
                setRepositories([]);
            }
        } catch (err) {
            console.error('Failed to load repositories:', err);
            setError('Failed to load repositories');
            setRepositories([]);
        } finally {
            setIsLoading(false);
        }
    };

    const handleRefresh = async () => {
        setRefreshing(true);
        await loadRepositories();
        setRefreshing(false);
    };

    const handleSortChange = (key: SortKey) => {
        if (sortKey === key) {
            // Toggle direction if the same key is clicked
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            // Set new key and default to ascending
            setSortKey(key);
            setSortDirection('asc');
        }
    };

    const getSortIcon = (key: SortKey) => {
        if (sortKey !== key) {
            return <ChevronUpDownIcon className="h-4 w-4 text-gray-400" />;
        }

        return sortDirection === 'asc' ? (
            <ChevronUpIcon className="h-4 w-4 text-indigo-600" />
        ) : (
            <ChevronDownIcon className="h-4 w-4 text-indigo-600" />
        );
    };

    // Filter and sort repositories
    const filteredAndSortedRepositories = repositories
        .filter(repo => {
            // Apply search filter
            if (searchQuery) {
                const searchLower = searchQuery.toLowerCase();
                return (
                    repo.name.toLowerCase().includes(searchLower) ||
                    repo.owner.toLowerCase().includes(searchLower) ||
                    `${repo.owner}/${repo.name}`.toLowerCase().includes(searchLower)
                );
            }

            // Apply status filter
            if (filterStatus) {
                // Special handling for "in_progress" to include "pending" status
                if (filterStatus === 'in_progress') {
                    return repo.status === 'in_progress' || repo.status === 'pending';
                }
                return repo.status === filterStatus;
            }

            return true;
        })
        .sort((a, b) => {
            // Handle sorting
            if (sortKey === 'name') {
                const nameA = `${a.owner}/${a.name}`.toLowerCase();
                const nameB = `${b.owner}/${b.name}`.toLowerCase();
                return sortDirection === 'asc'
                    ? nameA.localeCompare(nameB)
                    : nameB.localeCompare(nameA);
            } else if (sortKey === 'status') {
                const statusA = a.status || '';
                const statusB = b.status || '';
                return sortDirection === 'asc'
                    ? statusA.localeCompare(statusB)
                    : statusB.localeCompare(statusA);
            } else if (sortKey === 'last_scan_at') {
                const dateA = a.last_scan_at ? new Date(a.last_scan_at).getTime() : 0;
                const dateB = b.last_scan_at ? new Date(b.last_scan_at).getTime() : 0;
                return sortDirection === 'asc' ? dateA - dateB : dateB - dateA;
            }

            return 0;
        });

    if (status === 'loading' || isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-600"></div>
            </div>
        );
    }

    return (
        <div>
            <div className="sm:flex sm:items-center sm:justify-between">
                <div className="flex items-center">
                    <FolderIcon className="h-8 w-8 text-indigo-600 dark:text-indigo-400 mr-3" />
                    <h1 className="text-xl font-semibold text-gray-900 dark:text-white">
                        Your Repositories
                    </h1>
                </div>
                <div className="mt-4 sm:mt-0 sm:flex sm:space-x-3">
                    <div className="flex space-x-3">
                        <button
                            onClick={handleRefresh}
                            className="inline-flex items-center rounded-md bg-white dark:bg-gray-800 px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700"
                            disabled={refreshing}
                        >
                            {refreshing ? (
                                <>
                                    <ArrowPathIcon className="animate-spin -ml-0.5 mr-2 h-5 w-5 text-gray-400" />
                                    Refreshing...
                                </>
                            ) : (
                                <>
                                    <ArrowPathIcon className="-ml-0.5 mr-2 h-5 w-5 text-gray-400" />
                                    Refresh
                                </>
                            )}
                        </button>
                        <Link
                            href="/dashboard"
                            className="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        >
                            <CodeBracketIcon className="-ml-0.5 mr-2 h-5 w-5" />
                            Add Repository
                        </Link>
                    </div>
                </div>
            </div>

            <div className="mt-6 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm">
                <div className="border-b border-gray-200 dark:border-gray-700 px-4 py-3 sm:flex sm:items-center sm:justify-between">
                    <div className="relative max-w-sm">
                        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
                            <MagnifyingGlassIcon className="h-4 w-4 text-gray-400" />
                        </div>
                        <input
                            type="text"
                            className="block w-full rounded-md border-0 py-1.5 pl-9 text-gray-900 dark:text-white dark:bg-gray-700 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            placeholder="Search repositories..."
                            value={searchQuery}
                            onChange={e => setSearchQuery(e.target.value)}
                        />
                    </div>
                    <div className="mt-3 sm:ml-4 sm:mt-0">
                        <div className="flex space-x-3">
                            <select
                                id="status-filter"
                                className="block w-full rounded-md border-0 py-1.5 pl-3 pr-10 text-gray-900 dark:text-white dark:bg-gray-700 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                value={filterStatus || ''}
                                onChange={e => setFilterStatus(e.target.value || null)}
                            >
                                <option value="">All statuses</option>
                                <option value="completed">Completed</option>
                                <option value="in_progress">In progress</option>
                                <option value="failed">Failed</option>
                            </select>
                        </div>
                    </div>
                </div>

                {error && (
                    <div className="m-4 p-4 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 rounded-md">
                        <div className="flex">
                            <ExclamationCircleIcon className="h-5 w-5 text-red-400 mr-2" />
                            <span>{error}</span>
                        </div>
                    </div>
                )}

                {filteredAndSortedRepositories.length === 0 && !isLoading ? (
                    <div className="text-center py-16">
                        <FolderIcon className="mx-auto h-12 w-12 text-gray-400" />
                        <h3 className="mt-2 text-sm font-semibold text-gray-900 dark:text-white">
                            No repositories
                        </h3>
                        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                            {searchQuery
                                ? 'No repositories match your search criteria.'
                                : filterStatus
                                  ? `No repositories with status: ${filterStatus}`
                                  : 'Add a repository to get started.'}
                        </p>
                        <div className="mt-6">
                            <Link
                                href="/dashboard"
                                className="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                            >
                                <CodeBracketIcon className="-ml-0.5 mr-2 h-5 w-5" />
                                Add Repository
                            </Link>
                        </div>
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                            <thead className="bg-gray-50 dark:bg-gray-700/50">
                                <tr>
                                    <th
                                        scope="col"
                                        className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer"
                                        onClick={() => handleSortChange('name')}
                                    >
                                        <div className="flex items-center space-x-1">
                                            <span>Repository</span>
                                            {getSortIcon('name')}
                                        </div>
                                    </th>
                                    <th
                                        scope="col"
                                        className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer"
                                        onClick={() => handleSortChange('status')}
                                    >
                                        <div className="flex items-center space-x-1">
                                            <span>Status</span>
                                            {getSortIcon('status')}
                                        </div>
                                    </th>
                                    <th
                                        scope="col"
                                        className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer"
                                        onClick={() => handleSortChange('last_scan_at')}
                                    >
                                        <div className="flex items-center space-x-1">
                                            <span>Last Scan</span>
                                            {getSortIcon('last_scan_at')}
                                        </div>
                                    </th>
                                    <th
                                        scope="col"
                                        className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
                                    >
                                        Actions
                                    </th>
                                </tr>
                            </thead>
                            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                                {filteredAndSortedRepositories.map((repo, index) => (
                                    <tr
                                        key={repo.id || `repo-${index}-${Date.now()}`}
                                        className="hover:bg-gray-50 dark:hover:bg-gray-700/50"
                                    >
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="flex items-center">
                                                <div className="flex-shrink-0">
                                                    <span
                                                        className={`inline-flex h-10 w-10 items-center justify-center rounded-lg ${
                                                            repo.status === 'completed'
                                                                ? 'bg-green-100 dark:bg-green-900/20'
                                                                : repo.status === 'in_progress' ||
                                                                    repo.status === 'pending'
                                                                  ? 'bg-yellow-100 dark:bg-yellow-900/20'
                                                                  : 'bg-gray-100 dark:bg-gray-700'
                                                        }`}
                                                    >
                                                        {repo.status === 'completed' ? (
                                                            <ShieldCheckIcon className="h-6 w-6 text-green-600 dark:text-green-400" />
                                                        ) : repo.status === 'in_progress' ||
                                                          repo.status === 'pending' ? (
                                                            <ClockIcon className="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
                                                        ) : (
                                                            <CodeBracketIcon className="h-6 w-6 text-gray-600 dark:text-gray-400" />
                                                        )}
                                                    </span>
                                                </div>
                                                <div className="ml-4">
                                                    <div className="text-sm font-medium text-gray-900 dark:text-white">
                                                        {repo.owner}/{repo.name}
                                                    </div>
                                                    <div className="text-sm text-gray-500 dark:text-gray-400 truncate max-w-xs">
                                                        {repo.url}
                                                    </div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                                            <span
                                                className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                    repo.status === 'completed'
                                                        ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
                                                        : repo.status === 'in_progress' ||
                                                            repo.status === 'pending'
                                                          ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
                                                          : repo.status === 'failed'
                                                            ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
                                                            : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
                                                }`}
                                            >
                                                {repo.status === 'completed' && (
                                                    <CheckCircleIcon className="mr-1 h-3.5 w-3.5 text-green-500 dark:text-green-400" />
                                                )}
                                                {(repo.status === 'in_progress' ||
                                                    repo.status === 'pending') && (
                                                    <ClockIcon className="mr-1 h-3.5 w-3.5 text-yellow-500 dark:text-yellow-400" />
                                                )}
                                                {repo.status === 'failed' && (
                                                    <ExclamationCircleIcon className="mr-1 h-3.5 w-3.5 text-red-500 dark:text-red-400" />
                                                )}
                                                {repo.status === 'completed'
                                                    ? 'Completed'
                                                    : repo.status === 'in_progress' ||
                                                        repo.status === 'pending'
                                                      ? 'In Progress'
                                                      : repo.status === 'failed'
                                                        ? 'Failed'
                                                        : repo.status || 'Unknown'}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                                            {repo.last_scan_at
                                                ? new Date(repo.last_scan_at).toLocaleString()
                                                : 'Never'}
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                            <div className="flex justify-end space-x-2">
                                                <Link
                                                    href={`/repositories/${repo.id}`}
                                                    className="text-indigo-600 dark:text-indigo-400 hover:text-indigo-900 dark:hover:text-indigo-300 px-2 py-1 rounded-md hover:bg-indigo-50 dark:hover:bg-indigo-900/20 flex items-center"
                                                >
                                                    <EyeIcon className="h-4 w-4 mr-1" />
                                                    View
                                                </Link>
                                                <button
                                                    onClick={async () => {
                                                        setIsLoading(true);
                                                        try {
                                                            // Make sure we have a valid URL to submit
                                                            if (!repo.url) {
                                                                throw new Error(
                                                                    'Repository URL is missing'
                                                                );
                                                            }
                                                            // Use the user's session email for notifications
                                                            const userEmail =
                                                                session?.user?.email || '';
                                                            const response =
                                                                await repositoryApi.submitRepository(
                                                                    repo.url,
                                                                    userEmail
                                                                );
                                                            if (response.data) {
                                                                console.log(
                                                                    'Scan initiated:',
                                                                    response.data
                                                                );
                                                            }
                                                            await loadRepositories();
                                                        } catch (error) {
                                                            console.error(
                                                                'Failed to rescan:',
                                                                error
                                                            );
                                                            setError(
                                                                'Failed to initiate scan. Please try again.'
                                                            );
                                                        } finally {
                                                            setIsLoading(false);
                                                        }
                                                    }}
                                                    disabled={
                                                        isLoading || repo.status === 'in_progress'
                                                    }
                                                    className={`px-2 py-1 rounded-md flex items-center ${
                                                        isLoading || repo.status === 'in_progress'
                                                            ? 'text-gray-400 dark:text-gray-600 cursor-not-allowed'
                                                            : 'text-indigo-600 dark:text-indigo-400 hover:text-indigo-900 dark:hover:text-indigo-300 hover:bg-indigo-50 dark:hover:bg-indigo-900/20'
                                                    }`}
                                                >
                                                    <ArrowPathRoundedSquareIcon className="h-4 w-4 mr-1" />
                                                    Rescan
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    );
}
