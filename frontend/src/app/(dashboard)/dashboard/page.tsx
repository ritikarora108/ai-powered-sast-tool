'use client';

import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import {
    ShieldExclamationIcon,
    ShieldCheckIcon,
    ClockIcon,
    ArrowPathIcon,
    PlusIcon,
    ChartBarIcon,
    ExclamationCircleIcon,
} from '@heroicons/react/24/outline';
import { OwaspTop10, AddRepositoryForm } from '@/components';
import { repositoryApi } from '@/services/api';

// Define types for dashboard data
interface ScanData {
    id: string;
    repositoryId: string;
    repositoryName: string;
    repositoryOwner: string;
    scanDate: string;
    vulnerabilityCount?: number;
    status: 'completed' | 'in_progress' | 'failed' | 'pending';
}

interface DashboardData {
    totalRepositories: number;
    scannedRepositories: number;
    totalVulnerabilities: number;
    recentScans: ScanData[];
    vulnerabilitiesByCategory: Record<string, number>;
}

// Internal type to handle the API response structure
interface RepoData {
    id: string;
    name: string;
    owner: string;
    status: string;
    last_scan_at?: string;
    vulnerability_count?: number;
    [key: string]: any; // Allow for other properties
}

export default function DashboardPage() {
    const { data: session, status } = useSession();
    const router = useRouter();
    const [isLoading, setIsLoading] = useState(true);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);
    const [dashboardData, setDashboardData] = useState<DashboardData>({
        totalRepositories: 0,
        scannedRepositories: 0,
        totalVulnerabilities: 0,
        recentScans: [],
        vulnerabilitiesByCategory: {},
    });

    useEffect(() => {
        if (status === 'unauthenticated') {
            router.push('/auth/signin');
        }
    }, [status, router]);

    useEffect(() => {
        if (status === 'authenticated') {
            loadDashboardData();
        }
    }, [status]);

    const loadDashboardData = async () => {
        setIsLoading(true);

        try {
            // Fetch actual repositories from the API
            const reposResponse = await repositoryApi.getRepositories();

            if (!reposResponse.data) {
                throw new Error('Failed to fetch repositories');
            }

            const repositories = reposResponse.data as RepoData[];
            const scannedRepos = repositories.filter(
                repo => repo.status === 'completed' || repo.status === 'failed'
            );

            // Calculate total vulnerabilities
            let totalVulns = 0;
            const vulnerabilitiesByCategory: Record<string, number> = {};

            // Prepare recent scans data
            const recentScans: ScanData[] = repositories
                .filter(repo => repo.last_scan_at)
                .map(repo => ({
                    id: repo.id,
                    repositoryId: repo.id,
                    repositoryName: repo.name,
                    repositoryOwner: repo.owner,
                    scanDate: repo.last_scan_at || '',
                    vulnerabilityCount: repo.vulnerability_count || 0,
                    status: repo.status as 'completed' | 'in_progress' | 'failed' | 'pending',
                }))
                .sort((a, b) => new Date(b.scanDate).getTime() - new Date(a.scanDate).getTime())
                .slice(0, 5); // Get only the 5 most recent scans

            // Sum up vulnerabilities
            recentScans.forEach(scan => {
                if (scan.vulnerabilityCount) {
                    totalVulns += scan.vulnerabilityCount;
                }
            });

            // Update dashboard data with real values
            setDashboardData({
                totalRepositories: repositories.length,
                scannedRepositories: scannedRepos.length,
                totalVulnerabilities: totalVulns,
                recentScans,
                vulnerabilitiesByCategory,
            });
        } catch (err) {
            console.error('Failed to load dashboard data:', err);
        } finally {
            setIsLoading(false);
        }
    };

    const handleSubmitRepository = async (repoUrl: string) => {
        setIsSubmitting(true);
        setError(null);
        setSuccessMessage(null);

        try {
            // Get user email for notifications
            const userEmail = session?.user?.email || '';

            // Submit the repository for scanning
            const response = await repositoryApi.submitRepository(repoUrl, userEmail);

            if (response.data) {
                // Show success message with email notification info
                setSuccessMessage(
                    `Repository scan initiated successfully! The scan is now in progress. You will be notified by email (${userEmail}) when the scan is complete.`
                );

                // Refresh dashboard data to include the new repository
                await loadDashboardData();
                return;
            }

            if (response.error) {
                // Extract error message from response.error which could be a string or an object
                const errorMessage =
                    typeof response.error === 'string'
                        ? response.error
                        : response.error.message || 'Unknown error occurred';
                setError(errorMessage);
            }
        } catch (err: any) {
            setError(err.message || 'Failed to submit repository');
            console.error('Error submitting repository:', err);
        } finally {
            setIsSubmitting(false);
        }
    };

    if (status === 'loading' || isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-600"></div>
            </div>
        );
    }

    // Map of OWASP category IDs to their full names
    const owaspCategoryNames: Record<string, string> = {
        'A01:2021': 'Broken Access Control',
        'A02:2021': 'Cryptographic Failures',
        'A03:2021': 'Injection',
        'A04:2021': 'Insecure Design',
        'A05:2021': 'Security Misconfiguration',
        'A06:2021': 'Vulnerable and Outdated Components',
        'A07:2021': 'Identification and Authentication Failures',
        'A08:2021': 'Software and Data Integrity Failures',
        'A09:2021': 'Security Logging and Monitoring Failures',
        'A10:2021': 'Server-Side Request Forgery',
    };

    return (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="mb-8">
                <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">Dashboard</h1>
                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    Welcome to the Keygraph SAST Tool Dashboard
                </p>
            </div>

            {/* Repository Submission Form */}
            <div className="mb-8">
                <AddRepositoryForm onSubmit={handleSubmitRepository} isSubmitting={isSubmitting} />

                {successMessage && (
                    <div className="mt-4 p-4 bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300 rounded-md">
                        <div className="flex">
                            <ShieldCheckIcon className="h-5 w-5 text-green-400 mr-2" />
                            <span>{successMessage}</span>
                        </div>
                    </div>
                )}

                {error && (
                    <div className="mt-4 p-4 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 rounded-md">
                        <div className="flex">
                            <ExclamationCircleIcon className="h-5 w-5 text-red-400 mr-2" />
                            <span>{error}</span>
                        </div>
                    </div>
                )}
            </div>

            {/* Quick stats */}
            <div className="grid grid-cols-1 gap-5 sm:grid-cols-3 mb-8">
                <div className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
                    <div className="px-4 py-5 sm:p-6">
                        <div className="flex items-center">
                            <div className="flex-shrink-0 bg-indigo-500 rounded-md p-3">
                                <ShieldCheckIcon
                                    className="h-6 w-6 text-white"
                                    aria-hidden="true"
                                />
                            </div>
                            <div className="ml-5 w-0 flex-1">
                                <dl>
                                    <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                                        Total Repositories
                                    </dt>
                                    <dd>
                                        <div className="text-lg font-medium text-gray-900 dark:text-white">
                                            {dashboardData.totalRepositories}
                                        </div>
                                    </dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                    <div className="bg-gray-50 dark:bg-gray-700 px-4 py-4 sm:px-6">
                        <div className="text-sm">
                            <Link
                                href="/repositories"
                                className="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
                            >
                                View all repositories
                            </Link>
                        </div>
                    </div>
                </div>

                <div className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
                    <div className="px-4 py-5 sm:p-6">
                        <div className="flex items-center">
                            <div className="flex-shrink-0 bg-green-500 rounded-md p-3">
                                <ChartBarIcon className="h-6 w-6 text-white" aria-hidden="true" />
                            </div>
                            <div className="ml-5 w-0 flex-1">
                                <dl>
                                    <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                                        Scanned Repositories
                                    </dt>
                                    <dd>
                                        <div className="text-lg font-medium text-gray-900 dark:text-white">
                                            {dashboardData.scannedRepositories} /{' '}
                                            {dashboardData.totalRepositories}
                                        </div>
                                    </dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                    <div className="bg-gray-50 dark:bg-gray-700 px-4 py-4 sm:px-6">
                        <div className="text-sm">
                            <Link
                                href="/repositories"
                                className="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
                            >
                                View scan results
                            </Link>
                        </div>
                    </div>
                </div>

                <div className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
                    <div className="px-4 py-5 sm:p-6">
                        <div className="flex items-center">
                            <div className="flex-shrink-0 bg-red-500 rounded-md p-3">
                                <ShieldExclamationIcon
                                    className="h-6 w-6 text-white"
                                    aria-hidden="true"
                                />
                            </div>
                            <div className="ml-5 w-0 flex-1">
                                <dl>
                                    <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                                        Total Vulnerabilities
                                    </dt>
                                    <dd>
                                        <div className="text-lg font-medium text-gray-900 dark:text-white">
                                            {dashboardData.totalVulnerabilities}
                                        </div>
                                    </dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                    <div className="bg-gray-50 dark:bg-gray-700 px-4 py-4 sm:px-6">
                        <div className="text-sm">
                            <button
                                onClick={loadDashboardData}
                                className="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
                            >
                                Refresh data
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Recent scans */}
                <div className="lg:col-span-2">
                    <div className="bg-white dark:bg-gray-800 shadow overflow-hidden rounded-lg">
                        <div className="px-4 py-5 sm:px-6 border-b border-gray-200 dark:border-gray-700">
                            <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white">
                                Recent Scans
                            </h3>
                        </div>
                        <div className="overflow-x-auto">
                            {dashboardData.recentScans.length === 0 ? (
                                <div className="p-6 text-center text-gray-500 dark:text-gray-400">
                                    No scans have been performed yet.
                                </div>
                            ) : (
                                <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                                    <thead className="bg-gray-50 dark:bg-gray-700">
                                        <tr>
                                            <th
                                                scope="col"
                                                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider"
                                            >
                                                Repository
                                            </th>
                                            <th
                                                scope="col"
                                                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider"
                                            >
                                                Scan Date
                                            </th>
                                            <th
                                                scope="col"
                                                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider"
                                            >
                                                Status
                                            </th>
                                            <th
                                                scope="col"
                                                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider"
                                            >
                                                Results
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                                        {dashboardData.recentScans.map((scan: any) => (
                                            <tr
                                                key={scan.id}
                                                className="hover:bg-gray-50 dark:hover:bg-gray-700"
                                            >
                                                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                                                    <Link
                                                        href={`/repositories/${scan.repositoryId}`}
                                                        className="hover:text-indigo-600 dark:hover:text-indigo-400"
                                                    >
                                                        {scan.repositoryOwner}/{scan.repositoryName}
                                                    </Link>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                                                    {new Date(scan.scanDate).toLocaleString()}
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    {scan.status === 'completed' ? (
                                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                                            <ShieldCheckIcon className="mr-1.5 h-3 w-3 text-green-500" />
                                                            Completed
                                                        </span>
                                                    ) : scan.status === 'in_progress' ? (
                                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                                            <ClockIcon className="mr-1.5 h-3 w-3 text-blue-500" />
                                                            In Progress
                                                        </span>
                                                    ) : (
                                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                                                            <ExclamationCircleIcon className="mr-1.5 h-3 w-3 text-red-500" />
                                                            Failed
                                                        </span>
                                                    )}
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                                                    {scan.status === 'completed' ? (
                                                        <span
                                                            className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                                scan.vulnerabilityCount > 0
                                                                    ? 'bg-red-100 text-red-800'
                                                                    : 'bg-green-100 text-green-800'
                                                            }`}
                                                        >
                                                            {scan.vulnerabilityCount}{' '}
                                                            vulnerabilities
                                                        </span>
                                                    ) : (
                                                        'N/A'
                                                    )}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            )}
                        </div>
                        <div className="bg-gray-50 dark:bg-gray-700 px-4 py-4 sm:px-6 border-t border-gray-200 dark:border-gray-600">
                            <div className="flex justify-between">
                                <Link
                                    href="/repositories"
                                    className="text-sm font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
                                >
                                    View all repositories
                                </Link>
                                <Link
                                    href="/repositories"
                                    className="inline-flex items-center text-sm font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
                                >
                                    <PlusIcon className="-ml-1 mr-1 h-5 w-5" aria-hidden="true" />
                                    Add repository
                                </Link>
                            </div>
                        </div>
                    </div>
                </div>

                {/* OWASP Top 10 */}
                <div>
                    <OwaspTop10 />
                </div>
            </div>
        </div>
    );
}
