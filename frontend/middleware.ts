import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { getToken } from 'next-auth/jwt';

export async function middleware(request: NextRequest) {
    const path = request.nextUrl.pathname;

    // Define public paths that don't require authentication
    const isPublicPath = path === '/auth/signin' || path === '/';

    // Get the session token
    const token = await getToken({
        req: request,
        secret: process.env.NEXTAUTH_SECRET,
    });

    // Redirect unauthenticated users to signin page ONLY for protected routes
    if (!isPublicPath && !token) {
        const url = new URL('/auth/signin', request.url);
        url.searchParams.set('callbackUrl', encodeURI(request.url));
        return NextResponse.redirect(url);
    }

    // Redirect authenticated users away from signin page to dashboard
    // But don't redirect from home page as we handle that client-side
    if (path === '/auth/signin' && token) {
        return NextResponse.redirect(new URL('/dashboard', request.url));
    }

    return NextResponse.next();
}

// Specify paths that trigger the middleware
export const config = {
    matcher: ['/', '/dashboard/:path*', '/repositories/:path*', '/auth/signin'],
};
