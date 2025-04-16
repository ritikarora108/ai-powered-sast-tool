import { apiClient } from '../api/apiClient';
import { AUTH_TOKEN_KEY, USER_KEY } from '@/constants';

export interface User {
    id: string;
    email: string;
    name: string;
    createdAt: string;
}

// Note: These interfaces are kept for type consistency,
// but we're only using Google Sign-in as per requirements
export interface LoginResponse {
    token: string;
    user: User;
}

export interface RegisterResponse {
    token: string;
    user: User;
}

// Removed email/password login function - only Google Sign-in is used

// Removed register function - only Google Sign-in is used

export const logout = (): void => {
    localStorage.removeItem(AUTH_TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
};

export const getCurrentUser = (): User | null => {
    const userData = localStorage.getItem(USER_KEY);
    return userData ? JSON.parse(userData) : null;
};

export const isAuthenticated = (): boolean => {
    return !!localStorage.getItem(AUTH_TOKEN_KEY);
};
