import React from 'react';
import {
    ShieldCheckIcon,
    CodeBracketIcon,
    ArrowTopRightOnSquareIcon,
} from '@heroicons/react/24/outline';

interface OwaspVulnerability {
    id: string;
    name: string;
    description: string;
}

export default function OwaspTop10() {
    const owaspVulnerabilities: OwaspVulnerability[] = [
        {
            id: 'A01:2021',
            name: 'Broken Access Control',
            description: 'Permission and authorization issues',
        },
        {
            id: 'A02:2021',
            name: 'Cryptographic Failures',
            description: 'Data encryption problems',
        },
        {
            id: 'A03:2021',
            name: 'Injection',
            description: 'SQL, NoSQL, OS, and LDAP injection',
        },
        {
            id: 'A04:2021',
            name: 'Insecure Design',
            description: 'Flaws in design and architecture',
        },
        {
            id: 'A05:2021',
            name: 'Security Misconfiguration',
            description: 'Missing security hardening',
        },
        {
            id: 'A06:2021',
            name: 'Vulnerable and Outdated Components',
            description: 'Using components with known vulnerabilities',
        },
        {
            id: 'A07:2021',
            name: 'Identification and Authentication Failures',
            description: 'Issues with identity confirmation',
        },
        {
            id: 'A08:2021',
            name: 'Software and Data Integrity Failures',
            description: 'Code and infrastructure integrity issues',
        },
        {
            id: 'A09:2021',
            name: 'Security Logging and Monitoring Failures',
            description: 'Insufficient logging and monitoring',
        },
        {
            id: 'A10:2021',
            name: 'Server-Side Request Forgery',
            description: 'Fetching remote resources without validation',
        },
    ];

    return (
        <div className="overflow-hidden rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-800 shadow-sm">
            <div className="p-6 border-b border-gray-200 dark:border-gray-700">
                <div className="flex items-center justify-between">
                    <div className="flex items-center">
                        <ShieldCheckIcon className="h-6 w-6 text-indigo-600 dark:text-indigo-400 mr-2" />
                        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                            OWASP Top 10 (2021)
                        </h3>
                    </div>
                    <a
                        href="https://owasp.org/Top10/"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-indigo-600 dark:text-indigo-400 text-sm font-medium flex items-center hover:text-indigo-500"
                    >
                        Learn more
                        <ArrowTopRightOnSquareIcon className="h-4 w-4 ml-1" />
                    </a>
                </div>
                <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
                    Our scanner detects the following OWASP Top 10 vulnerabilities
                </p>
            </div>
            <ul className="divide-y divide-gray-200 dark:divide-gray-700">
                {owaspVulnerabilities.map(vuln => (
                    <li key={vuln.id} className="px-6 py-4 hover:bg-gray-50 dark:hover:bg-gray-700">
                        <div className="flex items-start">
                            <div className="flex-shrink-0 pt-0.5">
                                <CodeBracketIcon className="h-5 w-5 text-indigo-500" />
                            </div>
                            <div className="ml-3">
                                <p className="text-sm font-medium text-gray-900 dark:text-white">
                                    {vuln.id} {vuln.name}
                                </p>
                                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                                    {vuln.description}
                                </p>
                            </div>
                        </div>
                    </li>
                ))}
            </ul>
            <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700">
                <a
                    href="https://owasp.org/Top10/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="w-full inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                >
                    View OWASP Top 10 Website
                    <ArrowTopRightOnSquareIcon className="ml-2 -mr-0.5 h-4 w-4" />
                </a>
            </div>
        </div>
    );
}
