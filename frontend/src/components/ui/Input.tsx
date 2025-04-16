import React from 'react';
import classNames from 'classnames';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
    label?: string;
    helperText?: string;
    error?: string;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
    ({ className, label, helperText, error, id, ...props }, ref) => {
        // Generate a unique ID if one isn't provided
        const inputId = id || Math.random().toString(36).substring(2, 9);

        return (
            <div className="w-full">
                {label && (
                    <label
                        htmlFor={inputId}
                        className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >
                        {label}
                    </label>
                )}
                <input
                    ref={ref}
                    id={inputId}
                    className={classNames(
                        'block w-full rounded-md shadow-sm sm:text-sm',
                        error
                            ? 'border-red-300 text-red-900 placeholder-red-300 focus:border-red-500 focus:ring-red-500 dark:border-red-700 dark:focus:border-red-600 dark:focus:ring-red-600'
                            : 'border-gray-300 focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-indigo-500',
                        className
                    )}
                    aria-invalid={error ? 'true' : 'false'}
                    aria-errormessage={error ? `${inputId}-error` : undefined}
                    {...props}
                />
                {helperText && !error && (
                    <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{helperText}</p>
                )}
                {error && (
                    <p
                        className="mt-1 text-sm text-red-600 dark:text-red-500"
                        id={`${inputId}-error`}
                    >
                        {error}
                    </p>
                )}
            </div>
        );
    }
);

Input.displayName = 'Input';

export default Input;
