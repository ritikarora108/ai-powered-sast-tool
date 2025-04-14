'use client';

import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import { repositoryApi } from '@/services/api';
import { Repository, ScanStatus } from '@/types';
import Link from 'next/link';
import DashboardLayout from '@/components/DashboardLayout';
import { CheckCircleIcon, ExclamationCircleIcon } from '@heroicons/react/24/outline';

export default function Dashboard() {
  const { data: session, status } = useSession();
  const router = useRouter();
  const [repoUrl, setRepoUrl] = useState('');
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [scanStatus, setScanStatus] = useState<ScanStatus | null>(null);
  const [pollingScanId, setPollingScandId] = useState<string | null>(null);
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(null);
  const [scannedRepoDetails, setScannedRepoDetails] = useState<{id: string, owner: string, name: string} | null>(null);
  const [githubUrl, setGithubUrl] = useState('');
  const [scanId, setScanId] = useState<string | null>(null);
  const [repoId, setRepoId] = useState<string | null>(null);
  const [repoDetails, setRepoDetails] = useState<{ owner: string; name: string } | null>(null);
  const [isPolling, setIsPolling] = useState(false);

  // Redirect if not authenticated
  useEffect(() => {
    if (status === 'unauthenticated') {
      router.push('/auth/signin');
    }
  }, [status, router]);

  // Load repositories
  useEffect(() => {
    if (status === 'authenticated') {
      // Add a small delay to ensure authentication is properly set up
      const timer = setTimeout(() => {
        loadRepositories();
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [status]);
  
  // Add a retry mechanism
  const [retryCount, setRetryCount] = useState(0);
  useEffect(() => {
    if (error && error.includes('401') && retryCount < 3) {
      const timer = setTimeout(() => {
        console.log(`Retrying after authentication error (attempt ${retryCount + 1})`);
        setRetryCount(prev => prev + 1);
        loadRepositories();
      }, 2000); // Wait 2 seconds between retries
      return () => clearTimeout(timer);
    }
  }, [error, retryCount]);

  // Clean up polling interval on unmount
  useEffect(() => {
    return () => {
      if (pollingInterval) {
        clearInterval(pollingInterval);
      }
    };
  }, [pollingInterval]);

  // Poll for scan status updates
  useEffect(() => {
    if (pollingScanId) {
      const pollForScanStatus = async () => {
        try {
          console.log(`Checking scan status for ${pollingScanId}...`);
          const response = await repositoryApi.getScanStatus(pollingScanId);
          
          if (response.data && response.data.status) {
            console.log(`Current scan status: ${response.data.status}`);
            
            // If scan is completed or failed, stop polling and refresh repos
            if (response.data.status === 'completed' || 
                response.data.status === 'failed' || 
                response.data.status === 'canceled' || 
                response.data.status === 'timed_out') {
              console.log(`Scan ${pollingScanId} ${response.data.status}. Stopping polling.`);
              if (pollingInterval) {
                clearInterval(pollingInterval);
                setPollingInterval(null);
              }
              setPollingScandId(null);
              
              // Reload repositories to see updated status
              loadRepositories();
              
              // Show notification if needed
              if (response.data.status === 'completed') {
                // Display completed notification
                setScanStatus(prev => ({
                  ...prev!,
                  status: 'completed',
                  message: `Scan completed successfully! View details to see results.`
                }));
                
                // Clear notification after 30 seconds
                setTimeout(() => {
                  setScanStatus(null);
                }, 30000);
              }
            }
          }
        } catch (err) {
          console.error('Error checking scan status:', err);
        }
      };
      
      // Initial check
      pollForScanStatus();
      
      // Set interval to check every 5 seconds
      const interval = setInterval(pollForScanStatus, 5000);
      setPollingInterval(interval);
      
      // Clean up on unmount
      return () => clearInterval(interval);
    }
  }, [pollingScanId]);

  const loadRepositories = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      console.log('Loading repositories...');
      const response = await repositoryApi.getRepositories();
      if (response.data) {
        setRepositories(response.data);
        console.log('Successfully loaded repositories');
      } else if (response.error) {
        setError(`${response.error} (Status: ${response.status})`);
        console.error('Repository load error:', response);
      }
    } catch (err) {
      setError('Failed to load repositories');
      console.error('Repository load exception:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!githubUrl.trim()) return;

    setIsLoading(true);
    setScanStatus(null);
    setScanId(null);
    setRepoId(null);
    setRepoDetails(null);

    try {
      // Parse owner and repo name from the GitHub URL
      const urlMatch = githubUrl.match(/github\.com\/([^\/]+)\/([^\/]+)/);
      const owner = urlMatch ? urlMatch[1] : '';
      const name = urlMatch ? urlMatch[2] : '';

      if (owner && name) {
        setRepoDetails({ owner, name });
      }

      const response = await repositoryApi.submitRepository(githubUrl);
      if (response.data && response.data.scan_id) {
        setScanId(response.data.scan_id);
        setScanStatus({
          scan_id: response.data.scan_id,
          status: response.data.status,
          run_id: response.data.run_id || '',
          repository: response.data.repository || '',
          message: 'Scan initiated successfully!'
        });
        
        // Extract repository ID from the response
        const repository = response.data.repository as string;
        if (repository) {
          // First check if repository is actually a database ID
          if (/^[a-f0-9-]+$/.test(repository)) {
            setRepoId(repository);
          } else {
            // Repository might be a URL - try to find matching repo by URL
            const matchingRepo = repositories.find(repo => 
              repo.url === repository
            );
            if (matchingRepo) {
              setRepoId(matchingRepo.id);
            } else {
              console.warn('Could not determine repository ID from scan response');
            }
          }
        }
        
        startPolling(response.data.scan_id);
      } else if (response.error) {
        setScanStatus({
          scan_id: '',
          status: 'error',
          run_id: '',
          repository: '',
          message: response.error
        });
      }
    } catch (err) {
      console.error('Failed to submit repository:', err);
      setScanStatus({
        scan_id: '',
        status: 'error',
        run_id: '',
        repository: '',
        message: 'Failed to submit repository. Please try again.'
      });
    } finally {
      setIsLoading(false);
    }
  };

  const startPolling = (id: string) => {
    setIsPolling(true);
    // Clear any existing interval
    if (pollingInterval) {
      clearInterval(pollingInterval);
    }

    const interval = setInterval(async () => {
      try {
        const response = await repositoryApi.getScanStatus(id);
        
        // Only process if we have data
        if (!response.data) return;
        
        // Update scan status with the response data
        setScanStatus({
          scan_id: response.data.scan_id,
          status: response.data.status,
          run_id: response.data.run_id || '',
          repository: response.data.repository || '',
          message: response.data.status === 'completed' 
            ? 'Scan completed successfully! View details to see results.'
            : response.data.status === 'error'
            ? 'Scan failed. Please try again.'
            : 'Scan in progress...'
        });

        // Check if scan has completed or failed
        const isComplete = ['completed', 'error'].includes(response.data.status);
        if (!isComplete) return;
        
        // Clear the polling interval
        clearInterval(interval);
        setIsPolling(false);
        setPollingInterval(null);
        
        // Handle repository ID if available
        if (response.data.repository) {
          const repoUrl = response.data.repository as string;
          // Check if it's a UUID (likely a database ID)
          if (/^[a-f0-9-]+$/.test(repoUrl)) {
            setRepoId(repoUrl);
          } else {
            // It's a URL - find matching repository
            const matchingRepo = repositories.find(repo => repo.url === repoUrl);
            if (matchingRepo) {
              setRepoId(matchingRepo.id);
            } else {
              // Try to reload repositories to find the new one
              const repoResponse = await repositoryApi.getRepositories();
              if (repoResponse.data) {
                const newMatchingRepo = repoResponse.data.find(repo => repo.url === repoUrl);
                if (newMatchingRepo) {
                  setRepoId(newMatchingRepo.id);
                }
              }
            }
          }
        }
        
        // Reload repositories
        loadRepositories();
        
        // Clear status after some time
        setTimeout(() => {
          setScanStatus(null);
          setScanId(null);
        }, 30000);
      } catch (err) {
        console.error('Failed to check scan status:', err);
        clearInterval(interval);
        setIsPolling(false);
        setPollingInterval(null);
        setScanStatus({
          scan_id: '',
          status: 'error',
          run_id: '',
          repository: '',
          message: 'Failed to check scan status. Please try again.'
        });
      }
    }, 5000);

    setPollingInterval(interval);
  };

  if (status === 'loading') {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  return (
    <DashboardLayout>
      <div className="py-6">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h1 className="text-2xl font-semibold text-gray-900">Dashboard</h1>
        </div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Form to add new repository */}
          <div className="bg-white shadow rounded-lg mt-5 p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Scan a GitHub Repository</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label htmlFor="githubUrl" className="block text-sm font-medium text-gray-700">
                  GitHub Repository URL
                </label>
                <div className="mt-1">
                  <input
                    type="text"
                    name="githubUrl"
                    id="githubUrl"
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                    placeholder="https://github.com/owner/repository"
                    value={githubUrl}
                    onChange={(e) => setGithubUrl(e.target.value)}
                    required
                  />
                </div>
              </div>
              <div>
                <button
                  type="submit"
                  disabled={isLoading}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  {isLoading ? (
                    <>
                      <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Scanning...
                    </>
                  ) : (
                    'Scan Repository'
                  )}
                </button>
              </div>
            </form>
            
            {scanStatus && (
              <div 
                className={`mt-4 p-4 rounded-md ${
                  scanStatus.status === 'completed' ? 'bg-green-50 text-green-700' :
                  scanStatus.status === 'error' ? 'bg-red-50 text-red-700' :
                  'bg-blue-50 text-blue-700'
                }`}
              >
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    {scanStatus.status === 'completed' && (
                      <CheckCircleIcon className="h-5 w-5 text-green-400" aria-hidden="true" />
                    )}
                    {scanStatus.status === 'error' && (
                      <ExclamationCircleIcon className="h-5 w-5 text-red-400" aria-hidden="true" />
                    )}
                    {scanStatus.status === 'in_progress' && (
                      <div className="animate-spin h-5 w-5 text-blue-400">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                      </div>
                    )}
                  </div>
                  <div className="ml-3">
                    <p className="text-sm font-medium">
                      {scanStatus.message}
                      {scanStatus.status === 'completed' && repoId && (
                        <Link href={`/repositories/${repoId}`} className="ml-2 text-green-800 underline font-medium">
                          View details
                        </Link>
                      )}
                      {scanStatus.status === 'completed' && !repoId && (
                        <span className="ml-2 text-yellow-800">
                          Repository details will be available soon
                        </span>
                      )}
                    </p>
                    {repoDetails && (
                      <p className="text-xs mt-1">
                        Repository: {repoDetails.owner}/{repoDetails.name}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
          
          {/* List of repositories */}
          <div className="mt-8">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Your Repositories</h2>
            
            {repositories.length === 0 && !isLoading ? (
              <div className="bg-white shadow rounded-lg p-6 text-gray-500">
                No repositories found. Scan a repository to get started.
              </div>
            ) : (
              <div className="bg-white shadow overflow-hidden rounded-lg">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Repository</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Scan</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {repositories.map((repo) => (
                      <tr key={repo.id}>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="text-sm font-medium text-gray-900">
                            {repo.owner}/{repo.name}
                          </div>
                          <div className="text-sm text-gray-500">{repo.url}</div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                            repo.status === 'completed' ? 'bg-green-100 text-green-800' : 
                            repo.status === 'in_progress' ? 'bg-yellow-100 text-yellow-800' : 
                            'bg-gray-100 text-gray-800'
                          }`}>
                            {repo.status || 'Unknown'}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {repo.last_scan_at ? new Date(repo.last_scan_at).toLocaleString() : 'Never'}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                          <Link
                            href={`/repositories/${repo.id}`}
                            className="text-indigo-600 hover:text-indigo-900 mr-4"
                          >
                            View
                          </Link>
                          <button
                            onClick={async () => {
                              setIsLoading(true);
                              try {
                                await repositoryApi.submitRepository(repo.url);
                                loadRepositories();
                              } catch (error) {
                                console.error('Failed to rescan:', error);
                              } finally {
                                setIsLoading(false);
                              }
                            }}
                            disabled={isLoading}
                            className="text-indigo-600 hover:text-indigo-900"
                          >
                            Rescan
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
} 