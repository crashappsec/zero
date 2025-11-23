# OpenSSL Versions

## Current Supported Versions (2025)

### OpenSSL 3.2 (Current Stable)
- **First Release**: 3.2.0 (November 23, 2023)
- **Latest**: 3.2.x
- **Status**: Active development
- **Support Level**: Full support
- **End of Life**: TBD (approximately 1 year after 3.3 release)
- **LTS**: No

### OpenSSL 3.1
- **First Release**: 3.1.0 (March 14, 2023)
- **Latest**: 3.1.x
- **Status**: Security fixes only
- **Support Level**: Security fixes only
- **End of Life**: March 14, 2025 (approximately)
- **LTS**: No

### OpenSSL 3.0 (LTS)
- **First Release**: 3.0.0 (September 7, 2021)
- **Latest**: 3.0.x
- **Status**: Long-term support
- **Support Level**: Security and bug fixes
- **End of Life**: September 7, 2026
- **LTS**: Yes (5 years)
- **Major Features**: New provider architecture, deprecation of legacy APIs

## End-of-Life Versions

### OpenSSL 1.1.1 (EOL)
- **First Release**: 1.1.1 (September 11, 2018)
- **Final Release**: 1.1.1w (September 11, 2023)
- **Status**: END OF LIFE
- **End of Life**: September 11, 2023
- **LTS**: Was LTS (5 years)
- **Security Risk**: HIGH - No longer receiving security updates
- **Recommendation**: Upgrade to OpenSSL 3.0+ immediately

### OpenSSL 1.1.0 (EOL)
- **First Release**: 1.1.0 (August 25, 2016)
- **Final Release**: 1.1.0l (September 10, 2019)
- **Status**: END OF LIFE
- **End of Life**: September 11, 2019
- **LTS**: No
- **Security Risk**: CRITICAL
- **Recommendation**: Do not use

### OpenSSL 1.0.2 (EOL)
- **First Release**: 1.0.2 (January 22, 2015)
- **Final Release**: 1.0.2u (December 20, 2019)
- **Status**: END OF LIFE
- **End of Life**: December 31, 2019
- **LTS**: Was LTS
- **Security Risk**: CRITICAL - Multiple unpatched vulnerabilities
- **Recommendation**: Do not use

### OpenSSL 1.0.1 (EOL)
- **First Release**: 1.0.1 (March 14, 2012)
- **Final Release**: 1.0.1u (September 22, 2016)
- **Status**: END OF LIFE
- **End of Life**: December 31, 2016
- **LTS**: Was LTS
- **Security Risk**: CRITICAL - Includes Heartbleed
- **Known Vulnerabilities**: Heartbleed (CVE-2014-0160), many others
- **Recommendation**: Do not use

### OpenSSL 1.0.0 (EOL)
- **First Release**: 1.0.0 (March 29, 2010)
- **Final Release**: 1.0.0t (December 3, 2015)
- **Status**: END OF LIFE
- **End of Life**: December 31, 2015
- **Security Risk**: CRITICAL
- **Recommendation**: Do not use

### OpenSSL 0.9.8 and Earlier (EOL)
- **Status**: END OF LIFE
- **Security Risk**: CRITICAL
- **Recommendation**: Do not use

## Version Detection

### Command Line
```bash
# Get OpenSSL version
openssl version

# Get detailed version information
openssl version -a

# Version output formats:
# OpenSSL 3.2.0 23 Nov 2023
# OpenSSL 3.1.4 24 Oct 2023
# OpenSSL 3.0.13 30 Jan 2024
# OpenSSL 1.1.1w 11 Sep 2023
```

### Version String Parsing
```
OpenSSL X.Y.Z DD Mon YYYY
    X = Major version
    Y = Minor version
    Z = Patch version
```

### C API Version Detection

#### Compile-Time Version
```c
#include <openssl/opensslv.h>

// OpenSSL 3.x
#define OPENSSL_VERSION_MAJOR  3
#define OPENSSL_VERSION_MINOR  0
#define OPENSSL_VERSION_PATCH  0

// OpenSSL 1.1.1 and earlier
#define OPENSSL_VERSION_NUMBER 0x1010101fL
```

#### Runtime Version
```c
#include <openssl/crypto.h>

// OpenSSL 3.x
const char *version = OPENSSL_VERSION_STR;

// OpenSSL 1.1.0+
const char *version = OpenSSL_version(OPENSSL_VERSION);
unsigned long version_num = OpenSSL_version_num();

// OpenSSL 1.0.x and earlier
const char *version = SSLeay_version(SSLEAY_VERSION);
unsigned long version_num = SSLeay();
```

### Python Detection
```python
import ssl
print(ssl.OPENSSL_VERSION)
# Output: 'OpenSSL 3.0.2 15 Mar 2022'

import OpenSSL
print(OpenSSL.__version__)
```

### Ruby Detection
```ruby
require 'openssl'
puts OpenSSL::OPENSSL_VERSION
# Output: "OpenSSL 3.0.2 15 Mar 2022"
```

### Node.js Detection
```javascript
const crypto = require('crypto');
console.log(process.versions.openssl);
// Output: '3.0.2'
```

### Go Detection
```go
// Go uses BoringSSL or system OpenSSL
// Check at runtime
import "crypto/tls"
// No direct version query
```

## Version Number Format

### OpenSSL 3.x
```
X.Y.Z
X = Major version (3)
Y = Minor version (0, 1, 2, etc.)
Z = Patch version (0, 1, 2, etc.)
```

### OpenSSL 1.1.1 and earlier
```
Hexadecimal: 0xMNNFFPPS
M = Major version (1)
NN = Minor version (01 = 0.1, 10 = 1.0, 11 = 1.1)
FF = Fix/Patch level (00-FF)
PP = Patch level (00-FF)
S = Status (0-f)
    0 = development
    1-e = beta 1-14
    f = release

Example: 0x1010101fL = 1.1.1 release
```

## Major Version Differences

### OpenSSL 3.x vs 1.1.1

#### New in 3.x
- Provider architecture (replaces ENGINE)
- Deprecation of low-level APIs
- FIPS 140-3 compliance path
- Improved algorithm flexibility
- Better support for algorithm implementations

#### Breaking Changes
- Many low-level APIs deprecated
- ENGINE API deprecated
- Some algorithms deprecated/removed
- API changes in certificate handling

#### Migration Path
- Use EVP APIs instead of low-level APIs
- Replace ENGINE with providers
- Update certificate validation code
- Test thoroughly

### OpenSSL 1.1.1 vs 1.0.2

#### New in 1.1.1
- TLS 1.3 support
- New threading API
- Automatic initialization
- Improved PRNG
- ChaCha20-Poly1305 cipher suites

#### Breaking Changes
- Opaque structures (cannot access internals directly)
- Different initialization
- Threading changes
- Some API changes

### OpenSSL 1.0.2 vs 0.9.8

#### New in 1.0.2
- DTLS 1.2 support
- TLS extension support
- Suite B support
- Custom extension handling
- Certificate Transparency

## Version in Package Managers

### Debian/Ubuntu
```bash
# Check installed version
dpkg -l | grep libssl

# Available versions
apt-cache policy libssl-dev

# Version patterns
libssl3 = OpenSSL 3.x
libssl1.1 = OpenSSL 1.1.x
libssl1.0.0 = OpenSSL 1.0.x
```

### Red Hat/CentOS/Fedora
```bash
# Check installed version
rpm -qa | grep openssl

# Available versions
yum info openssl
dnf info openssl
```

### Alpine
```bash
# Check installed version
apk info openssl

# Pattern: openssl-3.x.x
```

### macOS (Homebrew)
```bash
# Check installed version
brew list --versions openssl

# Multiple versions can coexist
openssl@3
openssl@1.1
```

## Docker Base Images

### Official Images
```dockerfile
# Alpine with OpenSSL 3.x
FROM alpine:3.19

# Ubuntu 24.04 with OpenSSL 3.x
FROM ubuntu:24.04

# Ubuntu 22.04 with OpenSSL 3.x
FROM ubuntu:22.04

# Ubuntu 20.04 with OpenSSL 1.1.1 (EOL)
FROM ubuntu:20.04

# Debian Bookworm with OpenSSL 3.x
FROM debian:bookworm

# Debian Bullseye with OpenSSL 1.1.1 (EOL)
FROM debian:bullseye
```

## Language Runtime Versions

### Python
- **Python 3.12+**: Typically OpenSSL 3.x
- **Python 3.10-3.11**: OpenSSL 1.1.1 or 3.x
- **Python 3.9**: OpenSSL 1.1.1
- **Python 3.8 and earlier**: May use older OpenSSL

### Node.js
- **Node.js 20+**: OpenSSL 3.x
- **Node.js 18**: OpenSSL 3.x
- **Node.js 17**: OpenSSL 3.x
- **Node.js 16**: OpenSSL 1.1.1
- **Node.js 14 and earlier**: OpenSSL 1.1.1 or older

### Ruby
- **Ruby 3.2+**: OpenSSL 3.x support
- **Ruby 3.1**: OpenSSL 1.1.1 or 3.x
- **Ruby 3.0**: OpenSSL 1.1.1
- **Ruby 2.x**: OpenSSL 1.1.1 or older

## Version Compatibility Matrix

| OpenSSL Version | TLS 1.3 | TLS 1.2 | TLS 1.1 | TLS 1.0 | SSLv3 | SSLv2 |
|----------------|---------|---------|---------|---------|-------|-------|
| 3.2.x          | ✅      | ✅      | ✅*     | ✅*     | ❌**  | ❌    |
| 3.1.x          | ✅      | ✅      | ✅*     | ✅*     | ❌**  | ❌    |
| 3.0.x          | ✅      | ✅      | ✅*     | ✅*     | ❌**  | ❌    |
| 1.1.1 (EOL)    | ✅      | ✅      | ✅*     | ✅*     | ❌**  | ❌    |
| 1.1.0 (EOL)    | ❌      | ✅      | ✅*     | ✅*     | ❌**  | ❌    |
| 1.0.2 (EOL)    | ❌      | ✅      | ✅      | ✅      | ✅*** | ❌    |

\* Deprecated, must be explicitly enabled
\*\* Removed, not available
\*\*\* Available but not recommended

## Release Cadence

### Major Versions (X.0.0)
- Infrequent (years between releases)
- May include breaking changes
- Extended support (LTS)

### Minor Versions (3.X.0)
- Regular releases (months)
- New features
- No breaking changes within major version
- Short support window (until next minor + 1 year)

### Patch Versions (3.0.X)
- Frequent releases (weeks/months)
- Bug fixes and security updates
- No API changes
- Critical security fixes released immediately

## Version Selection Guidelines

### For New Projects
```
✅ RECOMMENDED: OpenSSL 3.0.x (LTS)
✅ ACCEPTABLE: OpenSSL 3.2.x (latest stable)
```

### For Existing Projects
```
If using 1.1.1: Upgrade to 3.0.x (1.1.1 is EOL)
If using 1.1.0 or earlier: Upgrade urgently (critical security risk)
If using 3.0.x: Continue using, plan upgrade before EOL (2026)
If using 3.1.x: Upgrade to 3.2.x or stay on 3.0.x LTS
```

### For Compliance/Regulatory
```
✅ Use LTS version (3.0.x)
✅ Stay on supported version
✅ Apply security updates promptly
❌ Never use EOL versions
```

## Detection Confidence

- **HIGH**: openssl version command output
- **HIGH**: OPENSSL_VERSION_NUMBER or OPENSSL_VERSION_STR in code
- **HIGH**: Package manager version information
- **MEDIUM**: Version inferred from features used
- **MEDIUM**: Base image or runtime environment
- **LOW**: No explicit version information available
