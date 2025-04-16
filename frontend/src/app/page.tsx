'use client'; // This directive marks this component as a client component, enabling client-side interactivity

import Link from 'next/link';
import { useSession } from 'next-auth/react'; // For authentication state management
import { useRouter } from 'next/navigation'; // For programmatic navigation
import {
    ShieldCheckIcon,
    ServerIcon,
    CodeBracketIcon,
    LockClosedIcon,
} from '@heroicons/react/24/outline'; // SVG icons for UI elements

// Array of features displayed in the features section
// Each feature has a name, description, and associated icon component
const features = [
    {
        name: 'GitHub Integration',
        description: 'Seamlessly analyze public repositories with a simple URL.',
        icon: CodeBracketIcon,
    },
    {
        name: 'OWASP Top 10 Coverage',
        description: 'Detect vulnerabilities categorized by OWASP Top 10 security risks.',
        icon: ShieldCheckIcon,
    },
    {
        name: 'AI-Powered Analysis',
        description: 'Leverage OpenAI models for intelligent vulnerability detection.',
        icon: ServerIcon,
    },
    {
        name: 'Secure Authentication',
        description: 'Protected by Google Sign-In for enterprise-grade security.',
        icon: LockClosedIcon,
    },
];

// Home component - the landing page that shows to unauthenticated users
export default function HomePage() {
    const router = useRouter();
    const { status } = useSession();

    // Redirect to dashboard if already authenticated
    if (status === 'authenticated') {
        router.push('/dashboard');
        return (
            <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
                <div className="w-full max-w-md">
                    <div className="mx-auto flex justify-center">
                        <div className="h-12 w-12 animate-spin rounded-full border-t-2 border-b-2 border-indigo-600"></div>
                    </div>
                    <h2 className="mt-6 text-center text-2xl font-bold tracking-tight text-gray-900">
                        Redirecting to Dashboard...
                    </h2>
                </div>
            </div>
        );
    }

    // Show loading screen while checking authentication status
    if (status === 'loading') {
        return (
            <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
                <div className="w-full max-w-md">
                    <div className="mx-auto flex justify-center">
                        <div className="h-12 w-12 animate-spin rounded-full border-t-2 border-b-2 border-indigo-600"></div>
                    </div>
                    <h2 className="mt-6 text-center text-2xl font-bold tracking-tight text-gray-900">
                        Loading Keygraph SAST Tool...
                    </h2>
                </div>
            </div>
        );
    }

    // Main content for unauthenticated users
    return (
        <div className="bg-white">
            {/* Hero section */}
            <div className="relative isolate px-6 pt-14 lg:px-8">
                <div className="mx-auto max-w-2xl py-32 sm:py-48 lg:py-56">
                    <div className="text-center">
                        <h1 className="text-4xl font-bold tracking-tight text-gray-900 sm:text-6xl">
                            AI-Powered Static Application Security Testing
                        </h1>
                        <p className="mt-6 text-lg leading-8 text-gray-600">
                            Scan your GitHub repositories for OWASP Top 10 vulnerabilities using
                            advanced AI models. Get detailed analysis and remediation advice to
                            secure your code.
                        </p>
                        <div className="mt-10 flex items-center justify-center gap-x-6">
                            <Link
                                href="/auth/signin"
                                className="rounded-md bg-indigo-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                            >
                                Sign in with Google
                            </Link>
                        </div>
                    </div>
                </div>
            </div>

            {/* Features section */}
            <div className="py-24 sm:py-32">
                <div className="mx-auto max-w-7xl px-6 lg:px-8">
                    <div className="mx-auto max-w-2xl lg:text-center">
                        <h2 className="text-base font-semibold leading-7 text-indigo-600">
                            Advanced Security
                        </h2>
                        <p className="mt-2 text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
                            Everything you need to secure your code
                        </p>
                        <p className="mt-6 text-lg leading-8 text-gray-600">
                            Our AI-powered SAST tool helps you identify security vulnerabilities in
                            your code before they become problems.
                        </p>
                    </div>
                    <div className="mx-auto mt-16 max-w-2xl sm:mt-20 lg:mt-24 lg:max-w-4xl">
                        <dl className="grid max-w-xl grid-cols-1 gap-x-8 gap-y-10 lg:max-w-none lg:grid-cols-2 lg:gap-y-16">
                            {features.map(feature => (
                                <div key={feature.name} className="relative pl-16">
                                    <dt className="text-base font-semibold leading-7 text-gray-900">
                                        <div className="absolute left-0 top-0 flex h-10 w-10 items-center justify-center rounded-lg bg-indigo-600">
                                            <feature.icon
                                                className="h-6 w-6 text-white"
                                                aria-hidden="true"
                                            />
                                        </div>
                                        {feature.name}
                                    </dt>
                                    <dd className="mt-2 text-base leading-7 text-gray-600">
                                        {feature.description}
                                    </dd>
                                </div>
                            ))}
                        </dl>
                    </div>
                </div>
            </div>
        </div>
    );
}
