'use client';

import { useEffect, useState } from 'react';
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
  const [isPolling, setIsPolling] = useState(false);
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (status === 'unauthenticated') {
      router.push('/auth/signin');
    }
  }, [status, router]);

  useEffect(() => {
    // Redirect to repositories page if ID is undefined
    if (params.id === 'undefined') {
      router.push('/repositories');
      return;
    }
    
    if (status === 'authenticated' && params.id) {
      loadRepositoryData();
    }
  }, [status, params.id, router]);

  // Clean up any existing polling interval when component unmounts
  useEffect(() => {
    return () => {
      if (pollingInterval) {
        clearInterval(pollingInterval);
      }
    };
  }, [pollingInterval]);

  const loadRepositoryData = async () => {
    setIsLoading(true);
    setError(null);

    try {
      // Get repository details
      const repoResponse = await repositoryApi.getRepository(params.id);
      if (repoResponse.data) {
        setRepository(repoResponse.data);
        
        // If repository status is in_progress, start polling
        if (repoResponse.data.status === 'in_progress') {
          startPolling();
        }
      } else if (repoResponse.error) {
        setError(repoResponse.error);
      }

      // Get vulnerabilities
      const vulnResponse = await repositoryApi.getVulnerabilities(params.id);
      if (vulnResponse.data) {
        setScanResults(vulnResponse.data);
      } else if (vulnResponse.error && !repoResponse.error) {
        setError(vulnResponse.error);
      }
    } catch (err) {
      setError('Failed to load repository data');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  const startPolling = () => {
    // Don't start polling if already polling
    if (isPolling) return;
    
    console.log(`Starting polling for repository ${params.id}`);
    setIsPolling(true);
    
    // Poll status every 5 seconds
    const interval = setInterval(async () => {
      try {
        console.log(`Checking scan status for repository ${params.id}...`);
        
        // Get latest repository status
        const repoResponse = await repositoryApi.getRepository(params.id);
        
        if (repoResponse.data) {
          setRepository(repoResponse.data);
          
          // If status has changed from in_progress to something else
          if (repoResponse.data.status !== 'in_progress') {
            console.log(`Scan completed with status: ${repoResponse.data.status}. Stopping polling.`);
            
            // Stop polling
            if (pollingInterval) {
              clearInterval(pollingInterval);
              setPollingInterval(null);
            }
            setIsPolling(false);
            
            // Refresh vulnerabilities data
            const vulnResponse = await repositoryApi.getVulnerabilities(params.id);
            if (vulnResponse.data) {
              setScanResults(vulnResponse.data);
            }
          }
        }
      } catch (err) {
        console.error('Error while polling repository status:', err);
        
        // In case of error, stop polling
        clearInterval(pollingInterval!);
        setPollingInterval(null);
        setIsPolling(false);
      }
    }, 5000);
    
    setPollingInterval(interval);
  };

  // Calculate total vulnerabilities
  const totalVulnerabilities = (scanResults?.vulnerabilities_count || scanResults?.categories?.reduce(
    (total, category) => total + category.vulnerabilities.length,
    0
  ) || 0);

  // Get categories from backend API response instead of hardcoding them
  const filteredCategories = activeTab === 'all' 
    ? scanResults?.categories || [] 
    : scanResults?.categories?.filter(category => category.name === activeTab) || [];

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
                <h1 className="text-2xl font-semibold text-gray-900">
                  {repository.owner}/{repository.name}
                </h1>
                <div className="flex space-x-3">
                  {isPolling && (
                    <div className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-yellow-800 bg-yellow-100">
                      <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-yellow-800" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Scan in progress...
                    </div>
                  )}
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
                  Last scan: {repository.last_scan_at ? new Date(repository.last_scan_at).toLocaleString() : 'Never'}
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
                        {totalVulnerabilities} vulnerabilities found across {scanResults.categories.length} OWASP categories
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
                        
                        {scanResults.categories.map((category) => (
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