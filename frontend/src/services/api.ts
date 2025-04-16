import axios, { AxiosError } from 'axios';
import { Repository, ScanResult, ScanStatus } from '../types';
import { getSession } from 'next-auth/react';
import { Session } from 'next-auth';

// Base URL for the backend API
// Falls back to localhost for development if the environment variable is not set
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Extends the NextAuth Session type to include our custom properties
// We need to store both the Google OAuth token and our backend JWT token
interface ExtendedSession extends Session {
    accessToken?: string; // Google OAuth access token
    idToken?: string; // Google OAuth ID token
    backendToken?: string; // Our backend JWT token
    accessTokenExpires?: number; // When the access token expires
}

// Define our API response interface
interface ApiResponse<T> {
    success?: boolean;
    data?: T;
    status?: number;
    error?:
        | {
              message: string;
              code?: string;
          }
        | string;
}

// Create a configured axios instance for making API requests
// This provides consistent configuration for all API calls
const api = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

// In-memory storage for the JWT token from our backend
// This avoids having to exchange tokens on every request
let backendJwtToken: string | null = null;
let tokenExchangeInProgress = false;
let lastTokenRefresh = 0;
let lastTokenError: Error | null = null;
let consecutiveFailures = 0;

// Define an error interface
interface ApiError extends Error {
    response?: {
        data?: {
            message?: string;
        };
        status?: number;
    };
    request?: unknown;
}

/**
 * Exchanges a Google OAuth token for a JWT token from our backend
 * This is necessary because our backend uses its own authentication system
 *
 * @param googleToken - The Google OAuth access token to exchange
 * @param idToken - Optional Google ID token which may be used as fallback
 * @returns A Promise that resolves to the backend JWT token or null if the exchange failed
 */
const exchangeGoogleToken = async (
    googleToken: string,
    idToken?: string
): Promise<string | null> => {
    // Prevent multiple simultaneous token exchanges
    if (tokenExchangeInProgress) {
        console.log('Token exchange already in progress, waiting...');
        // Wait for current exchange to finish
        await new Promise(resolve => setTimeout(resolve, 1000));
        return backendJwtToken;
    }

    // If we've had a token error recently and have reached max failures, return null
    // This prevents hammering the server with requests that are likely to fail
    if (lastTokenError && Date.now() - lastTokenRefresh < 60000 && consecutiveFailures > 3) {
        console.log('Too many recent token failures, cooling down');
        return null;
    }

    // Throttle token refresh requests to avoid hammering the server
    const now = Date.now();
    if (now - lastTokenRefresh < 5000) {
        // 5 seconds cooldown
        console.log('Token refresh throttled, using existing token');
        return backendJwtToken;
    }

    try {
        tokenExchangeInProgress = true;
        console.log('Attempting to exchange Google token for backend JWT');

        // First try with access token
        const response = await axios.post(
            `${API_URL}/auth/token`,
            { token: googleToken },
            {
                headers: {
                    'Content-Type': 'application/json',
                },
            }
        );

        if (response.data && response.data.token) {
            console.log('Successfully received backend JWT token');
            lastTokenRefresh = Date.now();
            backendJwtToken = response.data.token;
            lastTokenError = null;
            consecutiveFailures = 0;
            return response.data.token;
        }
        console.error('Backend response missing token:', response.data);

        // If access token exchange failed and we have an ID token, try that as fallback
        if (idToken) {
            console.log('Attempting to exchange Google ID token as fallback');
            const idTokenResponse = await axios.post(
                `${API_URL}/auth/token`,
                { token: idToken, token_type: 'id_token' },
                {
                    headers: {
                        'Content-Type': 'application/json',
                    },
                }
            );

            if (idTokenResponse.data && idTokenResponse.data.token) {
                console.log('Successfully received backend JWT token using ID token');
                lastTokenRefresh = Date.now();
                backendJwtToken = idTokenResponse.data.token;
                lastTokenError = null;
                consecutiveFailures = 0;
                return idTokenResponse.data.token;
            }
        }

        consecutiveFailures++;
        return null;
    } catch (error) {
        console.error('Failed to exchange Google token:', error);
        lastTokenError = error as Error;
        consecutiveFailures++;

        const axiosError = error as AxiosError;
        if (axiosError.response) {
            console.error('Error response data:', axiosError.response.data);
            console.error('Error response status:', axiosError.response.status);
        }
        return null;
    } finally {
        tokenExchangeInProgress = false;
    }
};

// Request interceptor to automatically add authentication tokens to all requests
// This implements a token management strategy with the following priorities:
// 1. Use backend token from session if available
// 2. Use cached backend token if available
// 3. Exchange Google token for backend token if Google token is available
// 4. Fall back to Google token directly (may not work with all endpoints)
api.interceptors.request.use(async config => {
    try {
        // Get session with authentication data
        const session = (await getSession()) as ExtendedSession | null;

        // If we have a backend token in the session, use it (highest priority)
        if (session?.backendToken) {
            console.log('Using backend JWT token from session');
            backendJwtToken = session.backendToken; // Cache the token for future use
            config.headers = { ...config.headers } as any;
            config.headers.Authorization = `Bearer ${session.backendToken}`;
            return config;
        }

        // If we have the backend JWT token in memory, use it
        if (backendJwtToken) {
            console.log('Using cached backend JWT token');
            config.headers = { ...config.headers } as any;
            config.headers.Authorization = `Bearer ${backendJwtToken}`;
            return config;
        }

        // If we have a Google token but no backend token, try to exchange it
        if (session?.accessToken && !backendJwtToken) {
            console.log('Attempting to exchange Google token');
            const token = await exchangeGoogleToken(session.accessToken, session.idToken);
            if (token) {
                backendJwtToken = token;
                console.log('Setting Authorization header with backend JWT');
                config.headers = { ...config.headers } as any;
                config.headers.Authorization = `Bearer ${token}`;
                return config;
            }
        }

        // Fall back to using the Google token directly (not ideal but may work if backend accepts it)
        if (session?.accessToken) {
            console.log('Falling back to Google token as authorization');
            config.headers = { ...config.headers } as any;
            config.headers.Authorization = `Bearer ${session.accessToken}`;
        } else {
            console.log('No authentication token available');
        }

        return config;
    } catch (error) {
        console.error('Error in request interceptor:', error);
        return config;
    }
});

// Response interceptor to handle authentication errors
// This automatically handles 401 Unauthorized errors by trying to refresh the token
api.interceptors.response.use(
    response => response,
    async error => {
        // If we get a 401 Unauthorized error, it might be due to an expired token
        if (error.response && error.response.status === 401) {
            console.log('Received 401 error, attempting to refresh token');

            // Clear the cached token
            backendJwtToken = null;

            try {
                // Get a fresh session
                const session = (await getSession()) as ExtendedSession | null;

                // If we have a Google token, try to get a new backend token
                if (session?.accessToken) {
                    // Force a token refresh by calling the token endpoint directly
                    const newToken = await exchangeGoogleToken(
                        session.accessToken,
                        session.idToken
                    );
                    if (newToken) {
                        // Update cached token
                        backendJwtToken = newToken;

                        // Retry the request with the new token
                        const config = error.config;
                        if (config) {
                            // Make sure headers object exists
                            config.headers = config.headers || {};
                            config.headers.Authorization = `Bearer ${newToken}`;

                            // Create a new axios instance for the retry to avoid interceptor loops
                            return await axios(config);
                        }
                    }
                }
            } catch (refreshError) {
                console.error('Error during token refresh:', refreshError);
                // Continue with original error if refresh fails
            }
        }

        return Promise.reject(error);
    }
);

// Repository API functions for interacting with the backend repository endpoints
export const repositoryApi = {
    /**
     * Submits a repository URL for scanning
     * This can be used anonymously or with authentication
     *
     * @param repoUrl - GitHub repository URL to scan
     * @param email - Optional email for notifications (for anonymous users)
     * @returns Promise with API response containing scan status
     */
    submitRepository: async (repoUrl: string, email?: string): Promise<ApiResponse<ScanStatus>> => {
        try {
            console.log(`Submitting repository URL: ${repoUrl} to ${API_URL}/scan`);

            // Ensure the URL is properly formatted
            if (!repoUrl.startsWith('http')) {
                repoUrl = 'https://' + repoUrl;
            }

            // Prepare request payload, always include email when provided
            const payload: { repo_url: string; email?: string } = { repo_url: repoUrl };
            if (email && email.trim()) {
                payload.email = email.trim();
                console.log(`Using email for notifications: ${email}`);
            } else {
                console.log(
                    'No email provided for notifications - using logged-in user email from backend'
                );
            }

            const response = await api.post('/scan', payload);
            console.log('Repository submission response:', response);

            return {
                data: response.data,
                status: response.status,
            };
        } catch (error: unknown) {
            console.error('Error submitting repository:', error);

            const apiError = error as ApiError;
            // Provide more detailed error information
            let errorMessage = 'An error occurred';

            if (apiError.response) {
                console.error('Error response data:', apiError.response.data);
                console.error('Error response status:', apiError.response.status);

                if (apiError.response.data?.message) {
                    errorMessage = apiError.response.data.message;
                } else if (typeof apiError.response.data === 'string') {
                    errorMessage = apiError.response.data as string;
                }
            } else if (apiError.message) {
                errorMessage = apiError.message;
            }

            return {
                error: errorMessage,
                status: apiError.response?.status || 500,
            };
        }
    },

    /**
     * Gets the current status of a scan
     *
     * @param scanId - ID of the scan to check
     * @returns Promise with API response containing scan status
     */
    getScanStatus: async (scanId: string): Promise<ApiResponse<ScanStatus>> => {
        try {
            const response = await api.get(`/scan/${scanId}/status`);
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error: unknown) {
            const apiError = error as ApiError;
            return {
                error: apiError.response?.data?.message || 'An error occurred',
                status: apiError.response?.status || 500,
            };
        }
    },

    // Get scan results
    getScanResults: async (scanId: string): Promise<ApiResponse<ScanResult>> => {
        try {
            const response = await api.get(`/scan/${scanId}/results`);
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error: unknown) {
            const apiError = error as ApiError;
            return {
                error: apiError.response?.data?.message || 'An error occurred',
                status: apiError.response?.status || 500,
            };
        }
    },

    // Get user repositories
    getRepositories: async (retryCount = 0): Promise<ApiResponse<Repository[]>> => {
        try {
            console.log(`Fetching repositories from ${API_URL}/repositories`);
            const response = await api.get('/repositories');
            console.log('Repositories API response:', response);

            // Validate the response data
            if (response.data === undefined || response.data === null) {
                throw new Error('No data received from server');
            }

            // Handle case where response might not be an array
            if (!Array.isArray(response.data)) {
                console.warn('Response is not an array:', response.data);
                // If data.repositories exists and is an array, use it
                if (response.data.repositories && Array.isArray(response.data.repositories)) {
                    console.log('Using repositories field from response');
                    return {
                        data: normalizeRepositoryData(response.data.repositories),
                        status: response.status,
                    };
                } else {
                    throw new Error('Invalid response format: expected array of repositories');
                }
            }

            // Normalize the repository data (convert property names from backend to frontend format)
            const normalizedData = normalizeRepositoryData(response.data);

            return {
                data: normalizedData,
                status: response.status,
            };
        } catch (error: unknown) {
            console.error('Error fetching repositories:', error);
            const apiError = error as ApiError;

            if (apiError.response) {
                console.error('Error response status:', apiError.response.status);
                console.error('Error response data:', apiError.response.data);
            } else if (apiError.request) {
                console.error('No response received:', apiError.request);
            } else {
                console.error('Error message:', apiError.message);
            }

            // Implement retry logic (max 3 attempts)
            if (retryCount < 3) {
                console.log(`Retrying repository fetch (attempt ${retryCount + 1}/3)`);
                // Wait for a short time before retrying
                await new Promise(resolve => setTimeout(resolve, 1000));
                return repositoryApi.getRepositories(retryCount + 1);
            }

            return {
                error: apiError.response?.data?.message || apiError.message || 'An error occurred',
                status: apiError.response?.status || 500,
            };
        }
    },

    // Get a single repository
    getRepository: async (id: string): Promise<ApiResponse<Repository>> => {
        try {
            const response = await api.get(`/repositories/${id}`);

            // Normalize the repository data from backend format to frontend format
            if (response.data) {
                const normalizedRepo = normalizeRepositoryData([response.data])[0];
                return {
                    data: normalizedRepo,
                    status: response.status,
                };
            }

            return {
                data: response.data,
                status: response.status,
            };
        } catch (error: unknown) {
            const apiError = error as ApiError;
            return {
                error: apiError.response?.data?.message || 'An error occurred',
                status: apiError.response?.status || 500,
            };
        }
    },

    // Get repository vulnerabilities
    getVulnerabilities: async (id: string): Promise<ApiResponse<ScanResult>> => {
        try {
            console.log(`Fetching vulnerabilities for repository ${id}`);
            const response = await api.get(`/repositories/${id}/vulnerabilities`);

            console.log('Vulnerability API response status:', response.status);
            console.log('Vulnerability API response type:', typeof response.data);

            if (Array.isArray(response.data)) {
                console.log(`Received array of vulnerabilities with ${response.data.length} items`);
                // If we get an array directly, wrap it in a ScanResult object
                return {
                    data: {
                        scan_id: id,
                        repository_id: id,
                        repository_name: '',
                        repository_url: '',
                        scan_started_at: new Date().toISOString(),
                        scan_completed_at: new Date().toISOString(),
                        vulnerabilities_count: response.data.length,
                        status: 'completed',
                        results_available: true,
                        vulnerabilities_by_category: { Unknown: response.data },
                    },
                    status: response.status,
                };
            } else if (response.data && typeof response.data === 'object') {
                console.log('Received object with properties:', Object.keys(response.data));

                // If we have a scan_id property, assume it's already a ScanResult
                if ('scan_id' in response.data) {
                    return {
                        data: response.data,
                        status: response.status,
                    };
                }

                // If we have neither scan_id nor vulnerabilities array, wrap in a ScanResult
                return {
                    data: {
                        scan_id: id,
                        repository_id: id,
                        repository_name: '',
                        repository_url: '',
                        scan_started_at: new Date().toISOString(),
                        scan_completed_at: new Date().toISOString(),
                        vulnerabilities_count: 0,
                        status: 'completed',
                        results_available: true,
                        ...response.data,
                    },
                    status: response.status,
                };
            }

            // Default return
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error: unknown) {
            const apiError = error as ApiError;
            console.error('Error fetching vulnerabilities:', apiError);
            return {
                error: apiError.response?.data?.message || 'An error occurred',
                status: apiError.response?.status || 500,
            };
        }
    },

    // Trigger a new scan for a repository
    scanRepository: async (id: string): Promise<ApiResponse<any>> => {
        try {
            console.log(`Initiating new scan for repository ${id}`);
            const response = await api.post(`/repositories/${id}/scan`);

            console.log('Scan initiation response:', response.data);
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error: unknown) {
            const apiError = error as ApiError;
            console.error('Error initiating scan:', apiError);

            // Provide detailed error information
            let errorMessage = 'An error occurred';
            if (apiError.response) {
                console.error('Error response data:', apiError.response.data);
                console.error('Error response status:', apiError.response.status);

                if (apiError.response.data?.message) {
                    errorMessage = apiError.response.data.message;
                } else if (typeof apiError.response.data === 'string') {
                    errorMessage = apiError.response.data as string;
                }
            } else if (apiError.message) {
                errorMessage = apiError.message;
            }

            return {
                error: errorMessage,
                status: apiError.response?.status || 500,
            };
        }
    },
};

// Helper function to normalize repository data from backend format to frontend format
interface RepositoryData {
    ID?: string;
    id?: string;
    Name?: string;
    name?: string;
    Owner?: string;
    owner?: string;
    URL?: string;
    url?: string;
    CloneURL?: string;
    clone_url?: string;
    CreatedAt?: string;
    created_at?: string;
    UpdatedAt?: string;
    updated_at?: string;
    LastScanAt?: string;
    last_scan_at?: string;
    Status?: string;
    status?: string;
}

function normalizeRepositoryData(repositories: RepositoryData[]): Repository[] {
    return repositories.map(repo => ({
        id: repo.ID || repo.id || '',
        name: repo.Name || repo.name || '',
        owner: repo.Owner || repo.owner || '',
        url: repo.URL || repo.url || '',
        clone_url: repo.CloneURL || repo.clone_url || '',
        created_at: repo.CreatedAt || repo.created_at || '',
        updated_at: repo.UpdatedAt || repo.updated_at || '',
        last_scan_at: repo.LastScanAt || repo.last_scan_at || '',
        status: repo.Status || repo.status || '',
    }));
}
