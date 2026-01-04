/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:3001/api/:path*',
      },
      {
        source: '/ws/:path*',
        destination: 'http://localhost:3001/ws/:path*',
      },
    ];
  },
};

module.exports = nextConfig;
