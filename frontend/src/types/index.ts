export interface Repository {
    id: string;
    owner: string;
    name: string;
    url: string;
    clone_url: string;
    created_at: string;
    updated_at: string;
    last_scan_at?: string;
    status?: string;
}

export interface ScanStatus {
    id?: string;
    scan_id?: string;
    status: string;
    run_id: string;
    repository: string;
    repository_id?: string;
    message?: string;
    results_available?: boolean;
}

export interface Vulnerability {
    id: string;
    repository_id: string;
    scan_id: string;
    category: string;
    severity: string;
    file_path: string;
    line_number: number;
    description: string;
    recommendation: string;
    created_at: string;
}

export interface OWASPCategory {
    id: string;
    name: string;
    vulnerabilities: Vulnerability[];
}

export interface ScanResult {
    scan_id: string;
    repository_id: string;
    repository_name: string;
    repository_url: string;
    scan_started_at: string;
    scan_completed_at: string;
    categories?: OWASPCategory[];
    vulnerabilities_count: number;
    status: string;
    message?: string;
    results_available?: boolean;
    vulnerabilities_by_category?: Record<string, any[]>;
    severity_counts?: {
        high: number;
        medium: number;
        low: number;
    };
}

// Common types used throughout the application

// Pagination types
export interface PaginationParams {
    page: number;
    pageSize: number;
}

export interface PaginatedResponse<T> {
    data: T[];
    totalCount: number;
    page: number;
    pageSize: number;
    totalPages: number;
}

// API Response types
export interface ApiResponse<T> {
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

// UI types
export type ButtonVariant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive';
export type ButtonSize = 'sm' | 'md' | 'lg';

export type AlertType = 'info' | 'success' | 'warning' | 'error';
