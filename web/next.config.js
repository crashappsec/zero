/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async redirects() {
    return [
      // Old scanner pages → New scanner pages
      {
        source: '/vulnerabilities',
        destination: '/scanners/code-security?feature=vulns',
        permanent: true,
      },
      {
        source: '/secrets',
        destination: '/scanners/code-security?feature=secrets',
        permanent: true,
      },
      {
        source: '/dependencies',
        destination: '/scanners/packages',
        permanent: true,
      },
      {
        source: '/ownership',
        destination: '/scanners/code-ownership',
        permanent: true,
      },
      {
        source: '/technology',
        destination: '/scanners/tech-id',
        permanent: true,
      },
      {
        source: '/devops',
        destination: '/scanners/devops',
        permanent: true,
      },
      {
        source: '/quality',
        destination: '/scanners/code-quality',
        permanent: true,
      },
      {
        source: '/devx',
        destination: '/scanners/devx',
        permanent: true,
      },
      // Projects → Repos rename
      {
        source: '/projects',
        destination: '/repos',
        permanent: true,
      },
      {
        source: '/projects/:id*',
        destination: '/repos/:id*',
        permanent: true,
      },
    ];
  },
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
