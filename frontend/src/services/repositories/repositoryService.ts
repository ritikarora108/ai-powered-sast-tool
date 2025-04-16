import { apiClient } from '../api/apiClient';
import { DEFAULT_PAGE_NUMBER, DEFAULT_PAGE_SIZE } from '@/constants';

export interface Repository {
    id: string;
    name: string;
    owner: string;
    url: string;
    description: string;
    stars: number;
    language: string;
    createdAt: string;
    updatedAt: string;
}

export interface RepositoryListResponse {
    repositories: Repository[];
    totalCount: number;
    page: number;
    pageSize: number;
}

export interface RepositoryScanResult {
    id: string;
    repositoryId: string;
    scanDate: string;
    status: 'pending' | 'in_progress' | 'completed' | 'failed';
    vulnerabilities: {
        critical: number;
        high: number;
        medium: number;
        low: number;
        info: number;
    };
    findings: {
        id: string;
        category: string;
        severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
        description: string;
        location: string;
        recommendation: string;
    }[];
}

export const getRepositories = async (
    page = DEFAULT_PAGE_NUMBER,
    pageSize = DEFAULT_PAGE_SIZE
): Promise<RepositoryListResponse> => {
    const response = await apiClient.get<RepositoryListResponse>('/repositories', {
        params: { page, pageSize },
    });
    return response.data;
};

export const getRepository = async (id: string): Promise<Repository> => {
    const response = await apiClient.get<Repository>(`/repositories/${id}`);
    return response.data;
};

export const scanRepository = async (id: string): Promise<{ scanId: string }> => {
    const response = await apiClient.post<{ scanId: string }>(`/repositories/${id}/scan`);
    return response.data;
};

export const getScanResult = async (
    repositoryId: string,
    scanId: string
): Promise<RepositoryScanResult> => {
    const response = await apiClient.get<RepositoryScanResult>(
        `/repositories/${repositoryId}/scans/${scanId}`
    );
    return response.data;
};

export const addRepository = async (githubUrl: string): Promise<Repository> => {
    const response = await apiClient.post<Repository>('/repositories', { githubUrl });
    return response.data;
};
