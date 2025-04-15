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
  scan_id: string;
  status: string;
  run_id: string;
  repository: string;
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
  name: string;
  description: string;
  vulnerabilities: Vulnerability[];
}

export interface ScanResult {
  scan_id: string;
  repository_id: string;
  repository_name: string;
  repository_url: string;
  scan_started_at: string;
  scan_completed_at: string;
  categories: OWASPCategory[];
  vulnerabilities_count: number;
  status: string;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  status: number;
} 