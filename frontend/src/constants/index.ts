// API URLs
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

// Authentication
export const AUTH_TOKEN_KEY = 'auth_token';
export const USER_KEY = 'user_data';

// Pagination
export const DEFAULT_PAGE_SIZE = 10;
export const DEFAULT_PAGE_NUMBER = 1;

// Routes
export const ROUTES = {
    HOME: '/',
    LOGIN: '/auth/login',
    REGISTER: '/auth/register',
    FORGOT_PASSWORD: '/auth/forgot-password',
    DASHBOARD: '/dashboard',
    REPOSITORIES: '/repositories',
    SETTINGS: '/settings',
};

// OWASP Top 10 Categories
export const OWASP_CATEGORIES = [
    {
        id: 'A01:2021',
        name: 'Broken Access Control',
        description: 'Restrictions on authenticated users are not properly enforced.',
    },
    {
        id: 'A02:2021',
        name: 'Cryptographic Failures',
        description:
            'Failures related to cryptography which often lead to sensitive data exposure.',
    },
    {
        id: 'A03:2021',
        name: 'Injection',
        description:
            'User-supplied data is not validated, filtered, or sanitized by the application.',
    },
    {
        id: 'A04:2021',
        name: 'Insecure Design',
        description: 'Flaws in the design that cannot be fixed by proper implementation.',
    },
    {
        id: 'A05:2021',
        name: 'Security Misconfiguration',
        description: 'Improper implementation of controls intended to keep application data safe.',
    },
    {
        id: 'A06:2021',
        name: 'Vulnerable and Outdated Components',
        description: 'Using components with known vulnerabilities.',
    },
    {
        id: 'A07:2021',
        name: 'Identification and Authentication Failures',
        description:
            'Incorrectly implemented authentication that allows attackers to compromise passwords.',
    },
    {
        id: 'A08:2021',
        name: 'Software and Data Integrity Failures',
        description:
            'Software and data integrity failures relate to code and infrastructure that does not protect against integrity violations.',
    },
    {
        id: 'A09:2021',
        name: 'Security Logging and Monitoring Failures',
        description: 'This category helps detect, escalate, and respond to active breaches.',
    },
    {
        id: 'A10:2021',
        name: 'Server-Side Request Forgery',
        description:
            'Occurs when a web application fetches a remote resource without validating the user-supplied URL.',
    },
];
