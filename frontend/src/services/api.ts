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
  getRepositories: async (retryCount = 0): Promise<ApiResponse<Repository[]>> => {
    try {
      console.log(`Fetching repositories from ${API_URL}/repositories`);
      const response = await api.get('/repositories');
      console.log('Repositories API response:', response);
      
      // Validate the response data
      if (!response.data) {
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
    } catch (error: any) {
      console.error('Error fetching repositories:', error);
      if (error.response) {
        console.error('Error response status:', error.response.status);
        console.error('Error response data:', error.response.data);
      } else if (error.request) {
        console.error('No response received:', error.request);
      } else {
        console.error('Error message:', error.message);
      }
      
      // Implement retry logic (max 3 attempts)
      if (retryCount < 3) {
        console.log(`Retrying repository fetch (attempt ${retryCount + 1}/3)`);
        // Wait for a short time before retrying
        await new Promise(resolve => setTimeout(resolve, 1000));
        return repositoryApi.getRepositories(retryCount + 1);
      }
      
      return {
        error: error.response?.data?.message || error.message || 'An error occurred',
        status: error.response?.status || 500,
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

// Helper function to normalize repository data from backend format to frontend format
function normalizeRepositoryData(repositories: any[]): Repository[] {
  return repositories.map(repo => ({
    id: repo.ID || repo.id,
    name: repo.Name || repo.name,
    owner: repo.Owner || repo.owner,
    url: repo.URL || repo.url,
    clone_url: repo.CloneURL || repo.clone_url,
    created_at: repo.CreatedAt || repo.created_at,
    updated_at: repo.UpdatedAt || repo.updated_at,
    last_scan_at: repo.LastScanAt || repo.last_scan_at,
    status: repo.Status || repo.status
  }));
} 