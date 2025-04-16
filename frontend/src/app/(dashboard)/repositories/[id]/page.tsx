'use client';

import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeftIcon, ArrowPathIcon } from '@heroicons/react/24/outline';
import { repositoryApi } from '@/services/api';
import ScanResults, {
    ScanResult as ComponentScanResult,
    Vulnerability as ComponentVulnerability,
} from '@/components/features/security/ScanResults';
import { ScanResult as ApiScanResult, Vulnerability as ApiVulnerability } from '@/types';

interface RepositoryDetailsPageProps {
    params: {
        id: string;
    };
}

// Convert API ScanResult to Component ScanResult format
const convertToComponentScanResult = (apiScanResult: ApiScanResult): ComponentScanResult => {
    console.log('Starting conversion from API to Component ScanResult');
    const vulnerabilities: ComponentVulnerability[] = [];

    // Process categories and extract vulnerabilities
    if (apiScanResult.categories) {
        console.log('Using categories format', apiScanResult.categories);
        apiScanResult.categories.forEach(category => {
            // Skip if not an OWASP Top 10 category (A01-A10)
            if (!category.id || !category.id.match(/^A(0[1-9]|10):2021$/)) {
                return;
            }

            category.vulnerabilities.forEach(vuln => {
                vulnerabilities.push({
                    id: vuln.id,
                    category: category.name,
                    categoryId: category.id || `A0${(vulnerabilities.length % 9) + 1}:2021`, // Use category id if available
                    title: vuln.description.split('.')[0] || 'Vulnerability',
                    description: vuln.description,
                    severity:
                        (vuln.severity.toLowerCase() as ComponentVulnerability['severity']) ||
                        'medium',
                    file_path: vuln.file_path,
                    line_number: vuln.line_number,
                    recommendation: vuln.recommendation,
                });
            });
        });
    } else if (apiScanResult.vulnerabilities_by_category) {
        console.log(
            'Using vulnerabilities_by_category format',
            apiScanResult.vulnerabilities_by_category
        );
        Object.entries(apiScanResult.vulnerabilities_by_category).forEach(
            ([categoryName, vulns]) => {
                // Check if vulns is an array
                if (!Array.isArray(vulns)) {
                    console.log(`Category ${categoryName} does not contain an array, skipping`);
                    return;
                }

                // Skip if not a valid OWASP Top 10 category
                if (!categoryName.match(/^A(0[1-9]|10):2021$/)) {
                    console.log(`Skipping non-OWASP Top 10 category: ${categoryName}`);
                    return;
                }

                console.log(
                    `Processing category: ${categoryName} with ${vulns.length} vulnerabilities`
                );
                vulns.forEach((vuln, index) => {
                    console.log(`Processing vulnerability: ${JSON.stringify(vuln)}`);
                    vulnerabilities.push({
                        id: vuln.id || `vuln-${index}`,
                        category: categoryName,
                        categoryId: categoryName.match(/^A(0[1-9]|10):2021$/)
                            ? categoryName
                            : getOwaspCategoryId(index),
                        title: vuln.title || vuln.description?.split('.')[0] || 'Vulnerability',
                        description: vuln.description || '',
                        severity:
                            (vuln.severity?.toLowerCase() as ComponentVulnerability['severity']) ||
                            'medium',
                        file_path: vuln.file_path,
                        line_number: vuln.line_number,
                        recommendation: vuln.recommendation,
                    });
                });
            }
        );
    } else if (Array.isArray(apiScanResult) && apiScanResult.length > 0) {
        // Handle case where direct array of vulnerabilities is returned
        console.log('Using direct vulnerability array format', apiScanResult);

        apiScanResult.forEach((vuln: any, index) => {
            // Try to determine category name from vulnerability type or a default
            const categoryName = vuln.vulnerability_type || vuln.type || 'Unknown';

            // Skip unknown vulnerabilities
            if (categoryName === 'Unknown') {
                console.log(`Skipping unknown vulnerability type: ${JSON.stringify(vuln)}`);
                return;
            }

            console.log(`Processing direct vulnerability ${index}: ${JSON.stringify(vuln)}`);
            vulnerabilities.push({
                id: vuln.id || `vuln-${index}`,
                category: categoryName,
                categoryId: getOwaspCategoryId(index),
                title: vuln.title || vuln.description?.split('.')[0] || 'Vulnerability',
                description: vuln.description || '',
                severity:
                    (vuln.severity?.toLowerCase() as ComponentVulnerability['severity']) ||
                    'medium',
                file_path: vuln.file_path || vuln.filepath,
                line_number: vuln.line_number || vuln.line_start,
                recommendation: vuln.recommendation || vuln.remediation,
            });
        });
    } else {
        console.log('Neither categories nor vulnerabilities_by_category found in API response');

        // Try to find any array property that might contain vulnerabilities
        const possibleArrayProps = Object.entries(apiScanResult).filter(
            ([_, value]) => Array.isArray(value) && value.length > 0
        );

        console.log(
            'Checking for possible array properties:',
            possibleArrayProps.map(([key]) => key)
        );

        for (const [key, value] of possibleArrayProps) {
            // Skip known non-vulnerability arrays
            if (['categories'].includes(key)) continue;

            console.log(
                `Found array property ${key} with ${value.length} items, attempting to process`
            );

            // Check if first item has properties that look like vulnerabilities
            const firstItem = value[0];
            if (
                firstItem &&
                typeof firstItem === 'object' &&
                ('severity' in firstItem ||
                    'type' in firstItem ||
                    'vulnerability_type' in firstItem ||
                    'description' in firstItem)
            ) {
                console.log(`Property ${key} looks like it contains vulnerabilities, processing`);

                value.forEach((vuln: any, index: number) => {
                    // Try to determine category name from vulnerability type or a default
                    const categoryName = vuln.vulnerability_type || vuln.type || 'Unknown';

                    // Skip unknown vulnerabilities
                    if (categoryName === 'Unknown') {
                        console.log(`Skipping unknown vulnerability type: ${JSON.stringify(vuln)}`);
                        return;
                    }

                    vulnerabilities.push({
                        id: vuln.id || `vuln-${index}`,
                        category: categoryName,
                        categoryId: getOwaspCategoryId(index),
                        title: vuln.title || vuln.description?.split('.')[0] || 'Vulnerability',
                        description: vuln.description || '',
                        severity:
                            (vuln.severity?.toLowerCase() as ComponentVulnerability['severity']) ||
                            'medium',
                        file_path: vuln.file_path || vuln.filepath,
                        line_number: vuln.line_number || vuln.line_start,
                        recommendation: vuln.recommendation || vuln.remediation,
                    });
                });

                break; // Stop after finding the first valid array
            }
        }
    }

    console.log(`Processed ${vulnerabilities.length} vulnerabilities total`);

    // If the scan is reported as completed but results_available is false, show as in_progress
    let status: ComponentScanResult['status'] = 'in_progress';
    if (apiScanResult.status === 'completed') {
        status = apiScanResult.results_available === false ? 'in_progress' : 'completed';
        console.log(
            `Setting status to ${status} based on results_available=${apiScanResult.results_available}`
        );
    } else if (apiScanResult.status === 'failed') {
        status = 'failed';
        console.log('Setting status to failed');
    } else {
        console.log(`Unknown status: ${apiScanResult.status}, defaulting to in_progress`);
    }

    return {
        id: apiScanResult.scan_id,
        repositoryId: apiScanResult.repository_id,
        status: status,
        startedAt: apiScanResult.scan_started_at,
        completedAt: apiScanResult.scan_completed_at,
        vulnerabilities,
        error: apiScanResult.message,
    };
};

// Helper function to get a valid OWASP category ID (A01-A10)
const getOwaspCategoryId = (index: number): string => {
    // Calculate a number between 1 and 10
    const categoryNum = (index % 10) + 1;
    // Format as A01, A02, etc., but A10 for the 10th
    return categoryNum === 10 ? 'A10:2021' : `A0${categoryNum}:2021`;
};

export default function RepositoryDetailsPage({ params }: RepositoryDetailsPageProps) {
    const { id } = params;
    const { data: session, status } = useSession();
    const router = useRouter();

    const [repository, setRepository] = useState<any>(null);
    const [scanResult, setScanResult] = useState<ComponentScanResult | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isScanning, setIsScanning] = useState(false);

    useEffect(() => {
        if (status === 'unauthenticated') {
            router.push('/auth/signin');
        }
    }, [status, router]);

    useEffect(() => {
        if (status === 'authenticated' && id) {
            loadRepositoryData();
        }
    }, [id, status]);

    const loadRepositoryData = async () => {
        setIsLoading(true);
        setError(null);

        try {
            // Fetch repository data
            const repoResponse = await repositoryApi.getRepository(id);
            if (repoResponse.error) {
                // Handle both string and object error formats
                const errorMessage =
                    typeof repoResponse.error === 'string'
                        ? repoResponse.error
                        : repoResponse.error.message || 'Unknown error';
                throw new Error(errorMessage);
            }

            if (repoResponse.data) {
                setRepository(repoResponse.data);
            }

            // Fetch vulnerabilities data
            const vulnResponse = await repositoryApi.getVulnerabilities(id);
            if (vulnResponse.error) {
                // Handle both string and object error formats
                const errorMessage =
                    typeof vulnResponse.error === 'string'
                        ? vulnResponse.error
                        : vulnResponse.error.message || 'Unknown error';
                throw new Error(errorMessage);
            }

            if (vulnResponse.data) {
                // Debug incoming data
                console.log('Raw vulnerability data from API:', vulnResponse.data);
                console.log('Results available flag:', vulnResponse.data.results_available);
                console.log('Status:', vulnResponse.data.status);
                console.log('Vulnerabilities count:', vulnResponse.data.vulnerabilities_count);
                console.log(
                    'Vulnerabilities by category:',
                    vulnResponse.data.vulnerabilities_by_category
                );

                // Check if scan is actually complete and results are available
                if (
                    vulnResponse.data.status === 'completed' &&
                    vulnResponse.data.results_available === false
                ) {
                    console.log(
                        'Scan marked as completed but results not available yet, polling...'
                    );
                    // Poll again in 5 seconds
                    setTimeout(() => loadRepositoryData(), 5000);
                    return;
                }

                // Convert API response to component format
                const componentScanResult = convertToComponentScanResult(vulnResponse.data);
                console.log('Converted scan result:', componentScanResult);
                console.log(
                    'Vulnerabilities after conversion:',
                    componentScanResult.vulnerabilities
                );
                setScanResult(componentScanResult);
            }
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred';
            console.error('Failed to load repository data:', err);
            setError(`Failed to load repository data: ${errorMessage}`);
        } finally {
            setIsLoading(false);
        }
    };

    const handleRescan = async () => {
        setIsScanning(true);
        try {
            // Trigger a new scan using the API service
            const scanResponse = await repositoryApi.scanRepository(id);

            if (scanResponse.error) {
                // Handle both string and object error formats
                const errorMessage =
                    typeof scanResponse.error === 'string'
                        ? scanResponse.error
                        : scanResponse.error.message || 'Unknown error';
                throw new Error(errorMessage);
            }

            console.log('Scan initiated successfully:', scanResponse.data);

            // Wait a moment to allow the scan to start
            await new Promise(resolve => setTimeout(resolve, 2000));

            // Reload repository data
            await loadRepositoryData();
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred';
            console.error('Failed to initiate scan:', err);
            setError(`Failed to initiate scan: ${errorMessage}`);
        } finally {
            setIsScanning(false);
        }
    };

    if (status === 'loading' || isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-600"></div>
            </div>
        );
    }

    return (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="mb-6">
                <Link
                    href="/repositories"
                    className="flex items-center text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 dark:hover:text-indigo-300 font-medium"
                >
                    <ArrowLeftIcon className="h-4 w-4 mr-1" />
                    Back to Repositories
                </Link>
            </div>

            {error && (
                <div className="mb-6 rounded-md bg-red-50 p-4">
                    <div className="flex">
                        <div className="ml-3">
                            <h3 className="text-sm font-medium text-red-800">Error</h3>
                            <div className="mt-2 text-sm text-red-700">
                                <p>{error}</p>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {repository && (
                <>
                    <div className="bg-white dark:bg-gray-800 shadow overflow-hidden rounded-lg mb-8">
                        <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
                            <div>
                                <h1 className="text-xl font-semibold text-gray-900 dark:text-white">
                                    {repository.owner}/{repository.name}
                                </h1>
                                <p className="mt-1 max-w-2xl text-sm text-gray-500 dark:text-gray-400">
                                    <a
                                        href={repository.url}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="hover:text-indigo-600 dark:hover:text-indigo-400"
                                    >
                                        {repository.url}
                                    </a>
                                </p>
                            </div>
                            <button
                                onClick={handleRescan}
                                disabled={isScanning}
                                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
                            >
                                {isScanning ? (
                                    <>
                                        <ArrowPathIcon className="animate-spin -ml-1 mr-2 h-5 w-5" />
                                        Scanning...
                                    </>
                                ) : (
                                    <>
                                        <ArrowPathIcon className="-ml-1 mr-2 h-5 w-5" />
                                        Run New Scan
                                    </>
                                )}
                            </button>
                        </div>
                        <div className="border-t border-gray-200 dark:border-gray-700 px-4 py-5 sm:p-0">
                            <dl className="sm:divide-y sm:divide-gray-200 dark:sm:divide-gray-700">
                                <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                                    <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                                        Repository URL
                                    </dt>
                                    <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100 sm:mt-0 sm:col-span-2">
                                        <a
                                            href={repository.url}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            className="text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
                                        >
                                            {repository.url}
                                        </a>
                                    </dd>
                                </div>
                                <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                                    <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                                        Last Scan
                                    </dt>
                                    <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100 sm:mt-0 sm:col-span-2">
                                        {repository.last_scan_at
                                            ? new Date(repository.last_scan_at).toLocaleString()
                                            : 'Never'}
                                    </dd>
                                </div>
                                <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                                    <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                                        Added On
                                    </dt>
                                    <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100 sm:mt-0 sm:col-span-2">
                                        {new Date(repository.created_at).toLocaleDateString()}
                                    </dd>
                                </div>
                            </dl>
                        </div>
                    </div>

                    {scanResult && <ScanResults scanResult={scanResult} />}
                </>
            )}
        </div>
    );
}
