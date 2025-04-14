'use client';

import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import { repositoryApi } from '@/services/api';
import { Repository } from '@/types';
import Link from 'next/link';
import DashboardLayout from '@/components/DashboardLayout';

export default function RepositoriesPage() {
  const { status } = useSession();
  const router = useRouter();
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(null);
  const [lastRefreshed, setLastRefreshed] = useState<Date | null>(null);

  useEffect(() => {
    if (status === 'unauthenticated') {
      router.push('/auth/signin');
    }
  }, [status, router]);

  useEffect(() => {
    if (status === 'authenticated') {
      loadRepositories();
    }
  }, [status]);

  // Clean up polling on unmount
  useEffect(() => {
    return () => {
      if (pollingInterval) {
        clearInterval(pollingInterval);
      }
    };
  }, [pollingInterval]);

  const loadRepositories = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await repositoryApi.getRepositories();
      if (response.data) {
        // Log the response for debugging
        console.log('Repositories response:', response.data);
        
        // Ensure each repository has a valid ID
        const validRepositories = response.data.map((repo, index) => {
          if (!repo.id) {
            // Generate a temporary ID if missing
            return { 
              ...repo, 
              id: `temp-${index}-${Date.now()}` // More unique temporary ID
            };
          }
          return repo;
        });
        
        setRepositories(validRepositories);
        setLastRefreshed(new Date());
        
        // Check if any repositories are in progress and start polling if so
        const hasInProgressScans = validRepositories.some(repo => repo.status === 'in_progress');
        if (hasInProgressScans) {
          startPolling();
        }
      } else if (response.error) {
        setError(response.error);
      }
    } catch (err) {
      setError('Failed to load repositories');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  const startPolling = () => {
    // Stop any existing polling
    if (pollingInterval) {
      clearInterval(pollingInterval);
    }
    
    // Check repository status every 5 seconds
    const interval = setInterval(async () => {
      console.log('Polling for repository status updates...');
      
      try {
        const response = await repositoryApi.getRepositories();
        if (response.data) {
          // Update repositories with latest data
          setRepositories(response.data);
          
          // Stop polling if no more in-progress scans
          const hasInProgressScans = response.data.some(repo => repo.status === 'in_progress');
          if (!hasInProgressScans) {
            console.log('No more in-progress scans. Stopping polling.');
            clearInterval(interval);
            setPollingInterval(null);
          }
        }
      } catch (err) {
        console.error('Error polling for repository updates:', err);
        // Don't stop polling on error, just continue
      }
    }, 5000);
    
    setPollingInterval(interval);
  };

  const handleRefresh = () => {
    loadRepositories();
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
          <div className="flex justify-between items-center">
            <h1 className="text-2xl font-semibold text-gray-900">Repositories</h1>
            <div className="flex space-x-3">
              <button
                onClick={handleRefresh}
                disabled={isLoading}
                className="inline-flex items-center px-3 py-2 border border-gray-300 text-sm font-medium rounded-md shadow-sm text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                {isLoading ? (
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-gray-700" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                ) : (
                  <svg xmlns="http://www.w3.org/2000/svg" className="-ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                )}
                {isLoading ? 'Refreshing...' : 'Refresh'}
              </button>
              <Link
                href="/dashboard"
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Add Repository
              </Link>
            </div>
          </div>
          
          {lastRefreshed && (
            <p className="text-xs text-gray-500 mt-1">
              Last refreshed: {lastRefreshed.toLocaleString()}
            </p>
          )}
          
          {error && (
            <div className="mt-4 p-4 bg-red-50 text-red-700 rounded-md">
              {error}
            </div>
          )}
          
          <div className="mt-8">
            {repositories.length === 0 && !isLoading ? (
              <div className="bg-white shadow rounded-lg p-6 text-gray-500">
                No repositories found. Add a repository to get started.
              </div>
            ) : (
              <div className="bg-white shadow overflow-hidden sm:rounded-lg">
                <div className="overflow-x-auto">
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
                      {repositories.map((repo, index) => (
                        <tr key={repo.id || `repo-${index}-${Date.now()}`} className="hover:bg-gray-50">
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
                              View Details
                            </Link>
                            <button
                              onClick={async () => {
                                setIsLoading(true);
                                try {
                                  // Make sure we have a valid URL to submit
                                  if (!repo.url) {
                                    throw new Error('Repository URL is missing');
                                  }
                                  const response = await repositoryApi.submitRepository(repo.url);
                                  if (response.data) {
                                    console.log('Scan initiated:', response.data);
                                  }
                                  await loadRepositories();
                                } catch (error) {
                                  console.error('Failed to rescan:', error);
                                  setError('Failed to initiate scan. Please try again.');
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
              </div>
            )}
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
} 