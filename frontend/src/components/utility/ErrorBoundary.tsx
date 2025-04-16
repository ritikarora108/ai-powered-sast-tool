'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
}

interface State {
    hasError: boolean;
}

class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = { hasError: false };
    }

    static getDerivedStateFromError(error: Error): State {
        // Log the error here if needed
        console.error('Error caught by ErrorBoundary:', error.message);
        return { hasError: true };
    }

    componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        // You can log the error to an error reporting service
        console.error('ErrorBoundary caught an error:', error, errorInfo);
    }

    render() {
        if (this.state.hasError) {
            // You can render any custom fallback UI
            return (
                this.props.fallback || (
                    <div className="p-4 bg-red-50 dark:bg-red-900/20 rounded-md m-4">
                        <h2 className="text-lg font-semibold text-red-800 dark:text-red-200">
                            Something went wrong
                        </h2>
                        <p className="text-sm text-red-700 dark:text-red-300">
                            The application encountered an error. Please try refreshing the page.
                        </p>
                    </div>
                )
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
