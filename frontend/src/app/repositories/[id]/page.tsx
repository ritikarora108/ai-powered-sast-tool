'use client';

import React, { useEffect, useState } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import { repositoryApi } from '@/services/api';
import { Repository, ScanResult, OWASPCategory } from '@/types';
import DashboardLayout from '@/components/DashboardLayout';

export default function RepositoryDetailPage({ params }: { params: { id: string } }) {
  const { status } = useSession();
  const router = useRouter();
  const [repository, setRepository] = useState<Repository | null>(null);
  const [scanResults, setScanResults] = useState<ScanResult | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<string>('all');
  const [mounted, setMounted] = useState(false);
  
  // Direct access to params.id is supported in current Next.js version
  // TODO: Update to use React.use() when required in future Next.js versions
  const id = params.id;
  
  // Add a mounted state to prevent hydration mismatches
  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (status === 'unauthenticated') {
      router.push('/auth/signin');
    }
  }, [status, router]);

  useEffect(() => {
    // Only execute this effect when the component is mounted client-side
    if (!mounted) return;
    
    // Redirect to repositories page if ID is undefined or invalid
    if (!id || id === 'undefined') {
      console.log('Repository ID is invalid, redirecting to repositories page');
      router.push('/repositories');
      return;
    }
    
    if (status === 'authenticated') {
      console.log(`Loading repository with ID: ${id}`);
      loadRepositoryData(id);
    }
  }, [status, id, router, mounted]);

  const loadRepositoryData = async (repoId: string) => {
    setIsLoading(true);
    setError(null);

    try {
      console.log(`Loading repository data for ID: ${repoId}`);
      // Get repository details
      const repoResponse = await repositoryApi.getRepository(repoId);
      
      if (repoResponse.data) {
        console.log("Repository data loaded:", repoResponse.data);
        setRepository(repoResponse.data);
      } else if (repoResponse.error) {
        console.error("Error loading repository:", repoResponse.error);
        setError(repoResponse.error);
      }

      // Get vulnerabilities
      const vulnResponse = await repositoryApi.getVulnerabilities(repoId);
      if (vulnResponse.data) {
        console.log("Vulnerabilities loaded:", vulnResponse.data);
        
        // Transform vulnerabilities data into required format if needed
        if (Array.isArray(vulnResponse.data) && !vulnResponse.data.categories) {
          // Group vulnerabilities by type (OWASP category)
          const groupedByType = vulnResponse.data.reduce((acc, vuln) => {
            const type = vuln.Type || 'Uncategorized';
            if (!acc[type]) {
              acc[type] = {
                name: type,
                description: `${type} vulnerabilities`,
                vulnerabilities: []
              };
            }
            acc[type].vulnerabilities.push({
              id: vuln.ID,
              description: vuln.Description,
              file_path: vuln.FilePath,
              line_number: vuln.LineStart,
              severity: vuln.Severity?.toLowerCase() || 'medium',
              recommendation: vuln.Remediation
            });
            return acc;
          }, {});
          
          // Convert to array structure expected by the UI
          const formattedData = {
            vulnerabilities_count: vulnResponse.data.length,
            categories: Object.values(groupedByType),
            scan_id: `scan-${repoId}`,
            repository_id: repoId,
            repository_name: repository?.name || 'unknown',
            repository_url: repository?.url || 'unknown',
            status: 'completed',
            created_at: new Date().toISOString(),
            scan_started_at: repository?.last_scan_at || new Date().toISOString(),
            scan_completed_at: new Date().toISOString(),
            severity_counts: {
              high: vulnResponse.data.filter(v => v.Severity?.toLowerCase() === 'high').length || 0,
              medium: vulnResponse.data.filter(v => v.Severity?.toLowerCase() === 'medium').length || 0,
              low: vulnResponse.data.filter(v => v.Severity?.toLowerCase() === 'low').length || 0
            }
          } as ScanResult;
          
          setScanResults(formattedData);
        } else {
          setScanResults(vulnResponse.data);
        }
      } else if (vulnResponse.error && !repoResponse.error) {
        console.error("Error loading vulnerabilities:", vulnResponse.error);
        setError(vulnResponse.error);
      }
    } catch (err) {
      console.error("Failed to load repository data:", err);
      setError('Failed to load repository data');
    } finally {
      setIsLoading(false);
    }
  };

  // Calculate total vulnerabilities
  const totalVulnerabilities = (scanResults?.vulnerabilities_count || 
    (scanResults?.categories?.reduce(
      (total, category) => total + category.vulnerabilities.length,
      0
    ) || 0));

  // Get categories from backend API response instead of hardcoding them
  const filteredCategories = activeTab === 'all' 
    ? scanResults?.categories || [] 
    : scanResults?.categories?.filter(category => category.name === activeTab) || [];

  // Format date safely to avoid hydration issues
  const formatDate = (dateString: string | undefined | null) => {
    if (!mounted || !dateString) return 'Never';
    try {
      return new Date(dateString).toLocaleString();
    } catch (e) {
      return 'Invalid date';
    }
  };

  if (status === 'loading' || isLoading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-screen">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="py-6">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {repository ? (
            <div>
              <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-900">
                  {repository.owner}/{repository.name}
                </h1>
                <div className="flex space-x-3">
                  <button
                    onClick={() => {
                      console.log(`Refreshing repository with ID: ${id}`);
                      if (id) loadRepositoryData(id);
                    }}
                    className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                    disabled={isLoading}
                  >
                    {isLoading ? (
                      <>
                        <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Refreshing...
                      </>
                    ) : (
                      'Refresh Results'
                    )}
                  </button>
                  <a
                    href={repository.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                  >
                    View on GitHub
                  </a>
                </div>
              </div>
              
              <div className="mt-2 flex items-center">
                <p className="text-sm text-gray-500 mr-4">
                  Last scan: {formatDate(repository.last_scan_at)}
                </p>
                <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${
                  repository.status === 'completed' ? 'bg-green-100 text-green-800' : 
                  repository.status === 'in_progress' ? 'bg-yellow-100 text-yellow-800' : 
                  'bg-gray-100 text-gray-800'
                }`}>
                  {repository.status || 'Unknown'}
                </span>
              </div>
              
              {error && (
                <div className="mt-4 p-4 bg-red-50 text-red-700 rounded-md">
                  {error}
                </div>
              )}
              
              {scanResults ? (
                <div className="mt-6">
                  <div className="bg-white shadow overflow-hidden sm:rounded-lg">
                    <div className="px-4 py-5 sm:px-6">
                      <h2 className="text-lg leading-6 font-medium text-gray-900">
                        Scan Results
                      </h2>
                      <p className="mt-1 max-w-2xl text-sm text-gray-500">
                        {totalVulnerabilities} vulnerabilities found across {scanResults?.categories?.length || 0} OWASP categories
                      </p>
                    </div>
                    
                    {/* Category Tabs */}
                    <div className="border-b border-gray-200">
                      <nav className="flex overflow-x-auto" aria-label="Tabs">
                        <button
                          onClick={() => setActiveTab('all')}
                          className={`${
                            activeTab === 'all'
                              ? 'border-indigo-500 text-indigo-600'
                              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                          } whitespace-nowrap py-4 px-4 border-b-2 font-medium text-sm`}
                        >
                          All ({totalVulnerabilities})
                        </button>
                        
                        {scanResults?.categories?.map((category) => (
                          <button
                            key={category.name}
                            onClick={() => setActiveTab(category.name)}
                            className={`${
                              activeTab === category.name
                                ? 'border-indigo-500 text-indigo-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                            } whitespace-nowrap py-4 px-4 border-b-2 font-medium text-sm`}
                          >
                            {category.name} ({category.vulnerabilities.length})
                          </button>
                        ))}
                      </nav>
                    </div>
                    
                    {/* Vulnerabilities List */}
                    <div className="divide-y divide-gray-200">
                      {filteredCategories.length === 0 ? (
                        <div className="px-4 py-5 sm:px-6 text-gray-500">
                          No vulnerabilities found in this category
                        </div>
                      ) : (
                        filteredCategories.map((category) => (
                          <div key={category.name} className="px-4 py-5 sm:px-6">
                            <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                              {category.name}
                            </h3>
                            <p className="mb-4 text-sm text-gray-500">
                              {category.description}
                            </p>
                            
                            <div className="mt-2 divide-y divide-gray-100">
                              {category.vulnerabilities.map((vuln) => (
                                <div key={vuln.id} className="py-4">
                                  <div className="flex items-start">
                                    <div className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                      vuln.severity === 'high' ? 'bg-red-100 text-red-800' :
                                      vuln.severity === 'medium' ? 'bg-yellow-100 text-yellow-800' :
                                      'bg-blue-100 text-blue-800'
                                    }`}>
                                      {vuln.severity}
                                    </div>
                                    <div className="ml-3 flex-1">
                                      <h4 className="text-sm font-medium text-gray-900">
                                        {vuln.file_path}:{vuln.line_number}
                                      </h4>
                                      <p className="mt-1 text-sm text-gray-600">
                                        {vuln.description}
                                      </p>
                                      <p className="mt-2 text-sm text-indigo-600">
                                        <strong>Recommendation:</strong> {vuln.recommendation}
                                      </p>
                                    </div>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>
                        ))
                      )}
                    </div>
                  </div>
                </div>
              ) : !isLoading && !error ? (
                <div className="mt-6 bg-white shadow overflow-hidden sm:rounded-lg">
                  <div className="px-4 py-5 sm:px-6 text-gray-500">
                    No scan results available. Run a scan to find vulnerabilities.
                  </div>
                </div>
              ) : null}
            </div>
          ) : !isLoading && !error ? (
            <div className="text-center py-12">
              <p className="text-lg text-gray-500">Repository not found</p>
            </div>
          ) : null}
        </div>
      </div>
    </DashboardLayout>
  );
} 