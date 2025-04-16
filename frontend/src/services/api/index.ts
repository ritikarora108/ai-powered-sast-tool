import axios from 'axios';

// Define the base API URL from environment variable
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Create axios instance with default config
const api = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Add request interceptor to attach authentication token
api.interceptors.request.use(
    config => {
        const token = localStorage.getItem('auth_token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    error => Promise.reject(error)
);

// Interface for repository data
export interface Repository {
    id: string;
    name: string;
    owner: string;
    url: string;
    status: 'completed' | 'in_progress' | 'pending' | 'failed';
    last_scan_at?: string;
    vulnerability_count?: number;
    created_at: string;
    updated_at: string;
}

// Repository API functions
export const repositoryApi = {
    // Get all repositories for authenticated user
    getRepositories: async () => {
        try {
            const response = await api.get('/repositories');
            return { data: response.data };
        } catch (error: any) {
            console.error('Failed to fetch repositories:', error);
            return { error: error.response?.data?.message || 'Failed to fetch repositories' };
        }
    },

    // Get a specific repository by ID
    getRepository: async (id: string) => {
        try {
            const response = await api.get(`/repositories/${id}`);
            return { data: response.data };
        } catch (error: any) {
            console.error(`Failed to fetch repository ${id}:`, error);
            return { error: error.response?.data?.message || 'Failed to fetch repository' };
        }
    },

    // Add a new repository for scanning
    addRepository: async (url: string) => {
        try {
            const response = await api.post('/repositories', { url });
            return { data: response.data };
        } catch (error: any) {
            console.error('Failed to add repository:', error);
            return { error: error.response?.data?.message || 'Failed to add repository' };
        }
    },

    // Get scan results for a repository
    getScanResults: async (repositoryId: string) => {
        try {
            const response = await api.get(`/repositories/${repositoryId}/scan`);
            return { data: response.data };
        } catch (error: any) {
            console.error(`Failed to fetch scan results for repository ${repositoryId}:`, error);
            return { error: error.response?.data?.message || 'Failed to fetch scan results' };
        }
    },

    // Trigger a new scan for a repository
    scanRepository: async (repositoryId: string) => {
        try {
            const response = await api.post(`/repositories/${repositoryId}/scan`);
            return { data: response.data };
        } catch (error: any) {
            console.error(`Failed to trigger scan for repository ${repositoryId}:`, error);
            return { error: error.response?.data?.message || 'Failed to trigger repository scan' };
        }
    },
};

export default api;
