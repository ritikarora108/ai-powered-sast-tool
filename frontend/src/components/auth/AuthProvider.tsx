'use client';

import { SessionProvider } from 'next-auth/react';
import { useEffect } from 'react';

export default function AuthProvider({ children }: { children: React.ReactNode }) {
    // Check for token in localStorage on mount to help with persistence
    useEffect(() => {
        // This helps ensure sessions are properly restored on page reload
        const checkStoredSession = async () => {
            // Trigger a session check
            try {
                const response = await fetch('/api/auth/session');
                if (response.ok) {
                    console.log('Session verified on application load');
                }
            } catch (error) {
                console.error('Error verifying session:', error);
            }
        };

        checkStoredSession();
    }, []);

    return (
        <SessionProvider
            refetchInterval={0}
            refetchOnWindowFocus={false}
            refetchWhenOffline={false}
        >
            {children}
        </SessionProvider>
    );
}
