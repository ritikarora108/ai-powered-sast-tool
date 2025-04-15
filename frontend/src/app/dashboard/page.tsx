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
  const [scanId, setScanId] = useState<string | null>(null);
  const [repoId, setRepoId] = useState<string | null>(null);
  const [repoDetails, setRepoDetails] = useState<{ owner: string; name: string } | null>(null);
  const [scannedRepoDetails, setScannedRepoDetails] = useState<{id: string, owner: string, name: string} | null>(null);
  const [githubUrl, setGithubUrl] = useState('');

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

  // Clean up any timers on unmount
  useEffect(() => {
    return () => {
      // Cleanup
    };
  }, []);

  const loadRepositories = async () => {
    setIsLoading(true);
    setError(null);
    
    console.log('Loading repositories...');
    
    try {
      const response = await repositoryApi.getRepositories();
      
      if (response.data) {
        console.log(`Loaded ${response.data.length} repositories`);
        setRepositories(response.data);
      } else if (response.error) {
        console.error('Error loading repositories:', response.error);
        setError(response.error);
      }
    } catch (err) {
      console.error('Failed to load repositories:', err);
      setError('Failed to load repositories. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!repoUrl) {
      return;
    }
    
    setIsLoading(true);
    setScanStatus(null);
    
    console.log(`Submitting repository: ${repoUrl}`);
    
    try {
      // Extract owner and name from repo URL
      const match = repoUrl.match(/github\.com\/([^\/]+)\/([^\/]+)/);
      if (match) {
        const owner = match[1];
        const name = match[2].replace(/\.git$/, '');
        setRepoDetails({ owner, name });
        
        console.log(`Extracted owner: ${owner}, name: ${name}`);
      }
      
      const response = await repositoryApi.submitRepository(repoUrl);
      
      if (response.data) {
        setScanStatus(response.data);
        setScanId(response.data.scan_id);
        
        // No need to start polling here
        // Instead, just update UI to show scan has started
        console.log('Scan submitted successfully:', response.data);
        setScanStatus({
          scan_id: response.data.scan_id,
          status: 'in_progress',
          run_id: response.data.run_id || '',
          repository: response.data.repository || '',
          message: 'Scan has been submitted. Please refresh the page to check for results.'
        });
        
        // Refresh repositories to see the new one
        setTimeout(() => {
          loadRepositories();
        }, 2000);
      } else if (response.error) {
        console.error('Error submitting repository:', response.error);
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
                    value={repoUrl}
                    onChange={(e) => setRepoUrl(e.target.value)}
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
              <div className={`mt-4 p-4 rounded-md ${
                scanStatus.status === 'error' ? 'bg-red-50 border border-red-200' :
                scanStatus.status === 'completed' ? 'bg-green-50 border border-green-200' :
                'bg-blue-50 border border-blue-200'
              }`}>
                <div className="flex">
                  <div className="flex-shrink-0">
                    {scanStatus.status === 'error' ? (
                      <ExclamationCircleIcon className="h-5 w-5 text-red-400" aria-hidden="true" />
                    ) : scanStatus.status === 'completed' ? (
                      <CheckCircleIcon className="h-5 w-5 text-green-400" aria-hidden="true" />
                    ) : (
                      <svg className="animate-spin h-5 w-5 text-blue-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                    )}
                  </div>
                  <div className="ml-3">
                    <h3 className={`text-sm font-medium ${
                      scanStatus.status === 'error' ? 'text-red-800' :
                      scanStatus.status === 'completed' ? 'text-green-800' :
                      'text-blue-800'
                    }`}>
                      {scanStatus.status === 'error' ? 'Error' :
                       scanStatus.status === 'completed' ? 'Success' :
                       'In Progress'}
                    </h3>
                    <div className={`mt-2 text-sm ${
                      scanStatus.status === 'error' ? 'text-red-700' :
                      scanStatus.status === 'completed' ? 'text-green-700' :
                      'text-blue-700'
                    }`}>
                      <p>{scanStatus.message}</p>
                      {scanStatus.status === 'in_progress' && (
                        <p className="mt-1 font-medium">Please refresh the page or check the repositories page to see results once the scan completes.</p>
                      )}
                    </div>
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