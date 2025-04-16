/**
 * Format a date string to a more readable format
 * @param dateString - The date string to format
 * @returns The formatted date string
 */
export const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
    });
};

/**
 * Truncate a string to a specified length
 * @param str - The string to truncate
 * @param length - The maximum length of the output string
 * @returns The truncated string
 */
export const truncateString = (str: string, length: number): string => {
    if (str.length <= length) return str;
    return str.slice(0, length) + '...';
};

/**
 * Capitalize the first letter of a string
 * @param str - The string to capitalize
 * @returns The capitalized string
 */
export const capitalize = (str: string): string => {
    if (!str) return '';
    return str.charAt(0).toUpperCase() + str.slice(1);
};

/**
 * Safe JSON parse with error handling
 * @param json - The JSON string to parse
 * @param fallback - The fallback value to return if parsing fails
 * @returns The parsed JSON or fallback value
 */
export const safeJsonParse = <T>(json: string, fallback: T): T => {
    try {
        return JSON.parse(json) as T;
    } catch (error) {
        return fallback;
    }
};
