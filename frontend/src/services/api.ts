import axios from 'axios';
import { ApiResponse, Repository, ScanResult, ScanStatus } from '@/types';
import { getSession } from 'next-auth/react';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Store JWT token
let backendJwtToken: string | null = null;

// Exchange Google token for backend JWT token
const exchangeGoogleToken = async (googleToken: string): Promise<string | null> => {
  try {
    console.log('Attempting to exchange Google token for backend JWT');
    const response = await axios.post(`${API_URL}/auth/token`, { token: googleToken }, {
      headers: {
        'Content-Type': 'application/json',
      }
    });
    
    if (response.data && response.data.token) {
      console.log('Successfully received backend JWT token');
      return response.data.token;
    }
    console.error('Backend response missing token:', response.data);
    return null;
  } catch (error: any) {
    console.error('Failed to exchange Google token:', error);
    if (error.response) {
      console.error('Error response data:', error.response.data);
      console.error('Error response status:', error.response.status);
    }
    return null;
  }
};

// Add auth token to requests if user is logged in
api.interceptors.request.use(async (config) => {
  try {
    const session = await getSession();
    
    // If we have a backend token in the session, use it (highest priority)
    if (session?.backendToken) {
      console.log('Using backend JWT token from session');
      config.headers.Authorization = `Bearer ${session.backendToken}`;
      return config;
    }
    
    // If we have the backend JWT token in memory, use it
    if (backendJwtToken) {
      console.log('Using cached backend JWT token');
      config.headers.Authorization = `Bearer ${backendJwtToken}`;
      return config;
    }
    
    // If we have a Google token but no backend token, try to exchange it
    if (session?.accessToken && !backendJwtToken) {
      console.log('Attempting to exchange Google token');
      const token = await exchangeGoogleToken(session.accessToken);
      if (token) {
        backendJwtToken = token;
        console.log('Setting Authorization header with backend JWT');
        config.headers.Authorization = `Bearer ${token}`;
        return config;
      }
    }
    
    // Fall back to using the Google token directly (not ideal but may work if backend accepts it)
    if (session?.accessToken) {
      console.log('Falling back to Google token as authorization');
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

export const repositoryApi = {
  // Submit a repository URL for scanning
  submitRepository: async (repoUrl: string): Promise<ApiResponse<ScanStatus>> => {
    try {
      const response = await api.post('/scan', { repo_url: repoUrl });
      return {
        data: response.data,
        status: response.status,
      };
    } catch (error: any) {
      return {
        error: error.response?.data?.message || 'An error occurred',
        status: error.response?.status || 500,
      };
    }
  },

  // Get scan status
  getScanStatus: async (scanId: string): Promise<ApiResponse<ScanStatus>> => {
    try {
      const response = await api.get(`/scan/${scanId}/status`);
      return {
        data: response.data,
        status: response.status,
      };
    } catch (error: any) {
      return {
        error: error.response?.data?.message || 'An error occurred',
        status: error.response?.status || 500,
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
    } catch (error: any) {
      return {
        error: error.response?.data?.message || 'An error occurred',
        status: error.response?.status || 500,
      };
    }
  },

  // Get user repositories
  getRepositories: async (): Promise<ApiResponse<Repository[]>> => {
    try {
      const response = await api.get('/repositories');
      return {
        data: response.data,
        status: response.status,
      };
    } catch (error: any) {
      return {
        error: error.response?.data?.message || 'An error occurred',
        status: error.response?.status || 500,
      };
    }
  },

  // Get a single repository
  getRepository: async (id: string): Promise<ApiResponse<Repository>> => {
    try {
      const response = await api.get(`/repositories/${id}`);
      return {
        data: response.data,
        status: response.status,
      };
    } catch (error: any) {
      return {
        error: error.response?.data?.message || 'An error occurred',
        status: error.response?.status || 500,
      };
    }
  },

  // Get repository vulnerabilities
  getVulnerabilities: async (id: string): Promise<ApiResponse<ScanResult>> => {
    try {
      const response = await api.get(`/repositories/${id}/vulnerabilities`);
      return {
        data: response.data,
        status: response.status,
      };
    } catch (error: any) {
      return {
        error: error.response?.data?.message || 'An error occurred',
        status: error.response?.status || 500,
      };
    }
  },
}; 