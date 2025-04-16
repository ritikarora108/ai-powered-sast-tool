'use client';

import { useSession } from 'next-auth/react';

// This component provides a simple way to protect routes that require authentication
export default function AuthGuard({ children }: { children: React.ReactNode }) {
    // Get session status to check if user is authenticated
    const { status } = useSession();

    // Show loading state while checking authentication
    if (status === 'loading') {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-600"></div>
            </div>
        );
    }

    // If authenticated, render the protected children components
    if (status === 'authenticated') {
        return <>{children}</>;
    }

    // For unauthenticated users, show loading - middleware will handle the redirect
    return (
        <div className="flex items-center justify-center min-h-screen">
            <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-600"></div>
        </div>
    );
}
