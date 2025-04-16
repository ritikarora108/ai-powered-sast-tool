import './globals.css';
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import { AuthProvider } from '@/components/auth';
import { ErrorBoundary } from '@/components/utility';

const inter = Inter({
    subsets: ['latin'],
    variable: '--font-inter',
    display: 'swap',
});

export const metadata: Metadata = {
    title: 'KeyGraph SAST - AI-powered Security Scanner',
    description:
        'AI-powered Static Application Security Testing tool that detects OWASP Top 10 vulnerabilities in your codebase',
    keywords: 'security, SAST, OWASP, vulnerabilities, AI, static analysis',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
    return (
        <html lang="en" className={`${inter.variable} scroll-smooth antialiased`}>
            <body className="min-h-screen bg-background text-foreground" suppressHydrationWarning>
                <ErrorBoundary>
                    <AuthProvider>{children}</AuthProvider>
                </ErrorBoundary>
            </body>
        </html>
    );
}
