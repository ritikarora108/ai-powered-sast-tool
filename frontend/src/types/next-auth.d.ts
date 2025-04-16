import 'next-auth';
import 'next-auth/jwt';

// Extend the built-in session type
declare module 'next-auth' {
    /**
     * Extend the built-in session types
     */
    interface Session {
        accessToken?: string;
        backendToken?: string;
        user: {
            id?: string;
            name?: string;
            email?: string;
            image?: string;
        };
    }
}

// Extend the built-in JWT type
declare module 'next-auth/jwt' {
    /**
     * Extend the built-in JWT type
     */
    interface JWT {
        accessToken?: string;
        backendToken?: string;
        refreshToken?: string;
        accessTokenExpires?: number;
        error?: string;
    }
}
