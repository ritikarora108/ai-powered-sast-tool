import NextAuth from 'next-auth';
import GoogleProvider from 'next-auth/providers/google';
import axios from 'axios';
import { JWT } from 'next-auth/jwt';
import { Session } from 'next-auth';
import type { NextAuthOptions } from 'next-auth';
import type { Account } from 'next-auth';
import type { SessionStrategy } from 'next-auth';

// Extend the JWT and Session types
declare module 'next-auth' {
    interface Session {
        accessToken?: string;
        backendToken?: string;
        idToken?: string;
        accessTokenExpires?: number;
    }
}

declare module 'next-auth/jwt' {
    interface JWT {
        accessToken?: string;
        refreshToken?: string;
        backendToken?: string;
        idToken?: string;
        accessTokenExpires?: number;
    }
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Function to exchange Google token for backend JWT
const exchangeGoogleToken = async (googleToken: string): Promise<string | null> => {
    try {
        console.log('Exchanging Google token for backend JWT');
        const response = await axios.post(
            `${API_URL}/auth/token`,
            { token: googleToken },
            {
                headers: {
                    'Content-Type': 'application/json',
                },
                timeout: 10000, // Add a timeout to avoid hanging requests
            }
        );

        if (response.data && response.data.token) {
            console.log('Successfully received backend JWT token');
            return response.data.token;
        }
        console.error('Backend response missing token:', response.data);
        return null;
    } catch (error: unknown) {
        const apiError = error as {
            response?: {
                data?: unknown;
                status?: number;
            };
            message?: string;
        };

        console.error('Failed to exchange Google token:', apiError);
        if (apiError.response) {
            console.error('Error response data:', apiError.response.data);
            console.error('Error response status:', apiError.response.status);
        }
        return null;
    }
};

// Define authentication options
export const authOptions: NextAuthOptions = {
    providers: [
        GoogleProvider({
            clientId: process.env.GOOGLE_CLIENT_ID!,
            clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
            authorization: {
                params: {
                    prompt: 'consent',
                    access_type: 'offline',
                    response_type: 'code',
                    scope: 'openid email profile https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile',
                },
            },
        }),
    ],
    session: {
        strategy: 'jwt' as SessionStrategy,
        maxAge: 30 * 24 * 60 * 60, // 30 days
        updateAge: 24 * 60 * 60, // 24 hours
    },
    callbacks: {
        async jwt({ token, account, user }: { token: JWT; account: Account | null; user?: any }) {
            // Initial sign in
            if (account && user) {
                console.log('Initial sign in, storing tokens in JWT');
                token.accessToken = account.access_token;
                token.refreshToken = account.refresh_token;
                token.accessTokenExpires = account.expires_at ? account.expires_at * 1000 : 0;
                token.idToken = account.id_token; // Store ID token as well

                // Try to exchange Google token for backend JWT
                if (account.access_token) {
                    try {
                        const backendToken = await exchangeGoogleToken(account.access_token);
                        if (backendToken) {
                            token.backendToken = backendToken;
                            console.log('Added backend JWT to session token');
                        } else {
                            console.error('Failed to get backend token during initial signin');
                        }
                    } catch (error) {
                        console.error('Error exchanging token in jwt callback:', error);
                    }
                }
                return token;
            }

            // Return previous token if the access token has not expired yet
            const accessTokenExpires = token.accessTokenExpires as number;
            if (accessTokenExpires && Date.now() < accessTokenExpires) {
                return token;
            }

            // Access token has expired, try to refresh it
            console.log('Token has expired, attempting to refresh');
            // For now, we don't have refresh token logic implemented
            // In a production app, you would use the refresh token to get a new access token

            // Just return the token even if expired, the client will handle refreshing if needed
            return token;
        },
        async session({ session, token }: { session: Session; token: JWT }) {
            // Add custom properties to the session object
            session.accessToken = token.accessToken as string;
            session.backendToken = token.backendToken as string;
            session.idToken = token.idToken as string; // Add ID token to session

            // Add token expiry info for client-side expiry checking
            session.accessTokenExpires = token.accessTokenExpires as number;
            return session;
        },
    },
    pages: {
        signIn: '/auth/signin',
    },
    debug: process.env.NODE_ENV === 'development',
};

// Create and export the handler
const handler = NextAuth(authOptions);
export { handler as GET, handler as POST };
