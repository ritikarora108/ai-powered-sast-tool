'use client';

import { useState } from 'react';
import {
    ShieldExclamationIcon,
    ShieldCheckIcon,
    CodeBracketIcon,
} from '@heroicons/react/24/outline';

// Define types for vulnerabilities and scan results
export interface Vulnerability {
    id: string;
    category: string;
    categoryId: string;
    title: string;
    description: string;
    severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
    file_path?: string;
    line_number?: number;
    code_snippet?: string;
    recommendation?: string;
}

export interface ScanResult {
    id: string;
    repositoryId: string;
    status: 'completed' | 'failed' | 'in_progress';
    startedAt: string;
    completedAt?: string;
    vulnerabilities: Vulnerability[];
    error?: string;
}

interface ScanResultsProps {
    scanResult: ScanResult;
}

export default function ScanResults({ scanResult }: ScanResultsProps) {
    const [activeCategory, setActiveCategory] = useState<string | null>(null);

    // Add debug log for development
    console.log('ScanResults render:', {
        status: scanResult.status,
        vulnerabilitiesLength: scanResult.vulnerabilities?.length || 0,
        scanResult,
    });

    // Ensure vulnerabilities is always an array
    const vulnerabilities = scanResult.vulnerabilities || [];

    // Filter out non-OWASP vulnerabilities while keeping all OWASP Top 10 (A01-A10)
    const filteredVulnerabilities = vulnerabilities.filter(vuln =>
        vuln?.categoryId?.match(/^A(0[1-9]|10):2021$/)
    );

    // Group vulnerabilities by OWASP category
    const vulnerabilitiesByCategory = filteredVulnerabilities.reduce<
        Record<string, Vulnerability[]>
    >((acc, vulnerability) => {
        const { categoryId } = vulnerability;
        if (!acc[categoryId]) {
            acc[categoryId] = [];
        }
        acc[categoryId].push(vulnerability);
        return acc;
    }, {});

    // Get OWASP categories with vulnerabilities
    const categoriesWithVulnerabilities = Object.keys(vulnerabilitiesByCategory);

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

    // Get severity color
    const getSeverityColor = (severity: Vulnerability['severity']) => {
        switch (severity) {
            case 'critical':
                return 'text-red-700 bg-red-100';
            case 'high':
                return 'text-orange-700 bg-orange-100';
            case 'medium':
                return 'text-yellow-700 bg-yellow-100';
            case 'low':
                return 'text-blue-700 bg-blue-100';
            case 'info':
            default:
                return 'text-gray-700 bg-gray-100';
        }
    };

    // Show success icon if no vulnerabilities found or error icon if vulnerabilities exist
    const showSuccessIcon =
        scanResult.status === 'completed' &&
        (vulnerabilities.length === 0 || filteredVulnerabilities.length === 0);

    return (
        <div className="bg-white dark:bg-gray-800 shadow overflow-hidden rounded-lg">
            <div className="px-4 py-5 sm:px-6 border-b border-gray-200 dark:border-gray-700">
                <div className="flex items-center">
                    {showSuccessIcon ? (
                        <ShieldCheckIcon className="h-6 w-6 text-green-500 mr-2" />
                    ) : (
                        <ShieldExclamationIcon className="h-6 w-6 text-red-500 mr-2" />
                    )}
                    <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                        Scan Results -{' '}
                        {scanResult.status === 'completed'
                            ? 'Complete'
                            : scanResult.status === 'in_progress'
                              ? 'In Progress'
                              : 'Failed'}
                    </h3>
                </div>
                <p className="mt-1 max-w-2xl text-sm text-gray-500 dark:text-gray-400">
                    {filteredVulnerabilities.length} vulnerabilities found across{' '}
                    {categoriesWithVulnerabilities.length} OWASP categories
                </p>
            </div>

            {scanResult.status === 'failed' && scanResult.error && (
                <div className="bg-red-50 p-4">
                    <div className="flex">
                        <div className="ml-3">
                            <h3 className="text-sm font-medium text-red-800">Scan failed</h3>
                            <div className="mt-2 text-sm text-red-700">
                                <p>{scanResult.error}</p>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {scanResult.status === 'in_progress' && (
                <div className="p-6 flex justify-center">
                    <div className="flex flex-col items-center">
                        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-600"></div>
                        <p className="mt-4 text-sm text-gray-500 dark:text-gray-400">
                            Scanning repository for vulnerabilities...
                        </p>
                    </div>
                </div>
            )}

            {scanResult.status === 'completed' && (
                <>
                    {filteredVulnerabilities.length === 0 || vulnerabilities.length === 0 ? (
                        <div className="p-6 text-center">
                            <ShieldCheckIcon className="h-12 w-12 text-green-500 mx-auto mb-4" />
                            <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                                No vulnerabilities found
                            </h3>
                            <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
                                Great job! Your code is free from known OWASP Top 10
                                vulnerabilities.
                            </p>
                        </div>
                    ) : (
                        <div className="flex flex-col md:flex-row">
                            {/* Categories sidebar */}
                            <div className="w-full md:w-1/4 bg-gray-50 dark:bg-gray-700 overflow-y-auto border-r border-gray-200 dark:border-gray-600">
                                <ul>
                                    {categoriesWithVulnerabilities.map(categoryId => (
                                        <li key={categoryId}>
                                            <button
                                                className={`w-full text-left px-4 py-3 flex items-center justify-between hover:bg-gray-100 dark:hover:bg-gray-600 ${
                                                    activeCategory === categoryId
                                                        ? 'bg-gray-100 dark:bg-gray-600'
                                                        : ''
                                                }`}
                                                onClick={() => setActiveCategory(categoryId)}
                                            >
                                                <span className="font-medium text-sm text-gray-800 dark:text-white">
                                                    {categoryId} - {owaspCategoryNames[categoryId]}
                                                </span>
                                                <span className="bg-indigo-100 text-indigo-800 dark:bg-indigo-800 dark:text-indigo-100 text-xs font-medium rounded-full px-2.5 py-0.5">
                                                    {vulnerabilitiesByCategory[categoryId].length}
                                                </span>
                                            </button>
                                        </li>
                                    ))}
                                </ul>
                            </div>

                            {/* Vulnerabilities detail */}
                            <div className="w-full md:w-3/4 overflow-y-auto p-4">
                                {activeCategory ? (
                                    <div>
                                        <h4 className="text-lg font-medium mb-4 text-gray-900 dark:text-white">
                                            {activeCategory} - {owaspCategoryNames[activeCategory]}
                                        </h4>
                                        <div className="space-y-4">
                                            {vulnerabilitiesByCategory[activeCategory].map(vuln => (
                                                <div
                                                    key={vuln.id}
                                                    className="border border-gray-200 dark:border-gray-700 rounded-md overflow-hidden"
                                                >
                                                    <div className="px-4 py-3 bg-gray-50 dark:bg-gray-700 flex items-center justify-between">
                                                        <div className="flex items-center">
                                                            <CodeBracketIcon className="h-5 w-5 text-gray-500 dark:text-gray-400 mr-2" />
                                                            <h5 className="font-medium text-gray-900 dark:text-white">
                                                                {vuln.title || 'Vulnerability'}
                                                            </h5>
                                                        </div>
                                                        <span
                                                            className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${getSeverityColor(vuln.severity)}`}
                                                        >
                                                            {vuln.severity.charAt(0).toUpperCase() +
                                                                vuln.severity.slice(1)}
                                                        </span>
                                                    </div>
                                                    <div className="px-4 py-3">
                                                        <p className="text-sm text-gray-600 dark:text-gray-300 mb-4">
                                                            {vuln.description ||
                                                                'No description available'}
                                                        </p>

                                                        {(vuln.file_path || vuln.line_number) && (
                                                            <div className="flex items-center text-sm text-gray-500 dark:text-gray-400 mb-2">
                                                                <span className="font-medium">
                                                                    Location:
                                                                </span>
                                                                <span className="ml-2">
                                                                    {vuln.file_path ||
                                                                        'Unknown file'}{' '}
                                                                    {vuln.line_number
                                                                        ? `(Line ${vuln.line_number})`
                                                                        : ''}
                                                                </span>
                                                            </div>
                                                        )}

                                                        {vuln.code_snippet && (
                                                            <div className="mb-4">
                                                                <div className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
                                                                    Code snippet:
                                                                </div>
                                                                <pre className="bg-gray-800 text-gray-200 p-3 rounded-md overflow-x-auto text-xs">
                                                                    <code>{vuln.code_snippet}</code>
                                                                </pre>
                                                            </div>
                                                        )}

                                                        {vuln.recommendation && (
                                                            <div>
                                                                <div className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
                                                                    Recommendation:
                                                                </div>
                                                                <p className="text-sm text-gray-600 dark:text-gray-300">
                                                                    {vuln.recommendation}
                                                                </p>
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                ) : (
                                    <div className="p-6 text-center">
                                        <p className="text-gray-500 dark:text-gray-400">
                                            Select a category to view vulnerabilities
                                        </p>
                                    </div>
                                )}
                            </div>
                        </div>
                    )}
                </>
            )}
        </div>
    );
}
