/** @type {import('next').NextConfig} */
const nextConfig = {
    reactStrictMode: true,
    images: {
        domains: [
            'lh3.googleusercontent.com', // Google profile images
            'github.com',                // GitHub profile images
            'avatars.githubusercontent.com', // GitHub avatars
            'githubusercontent.com',      // GitHub user content
            'gitlab.com',                // GitLab profile images
            's.gravatar.com',            // Gravatar images
            'secure.gravatar.com',       // Secure Gravatar images
            'res.cloudinary.com'         // Cloudinary hosted images
        ],
        remotePatterns: [
            {
                protocol: 'https',
                hostname: '**',
                pathname: '**',
            }
        ],
    },
    // Disable the development indicator that appears in the bottom right corner
    devIndicators: {
        position: 'bottom-right',
    },
};

export default nextConfig;
