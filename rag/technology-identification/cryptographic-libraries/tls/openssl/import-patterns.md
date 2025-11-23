# OpenSSL Import Patterns

## Package Names

### System Packages (Linux)

#### Debian/Ubuntu (APT)
- `libssl-dev` - Development files
- `libssl3` - Runtime library (OpenSSL 3.x)
- `libssl1.1` - Runtime library (OpenSSL 1.1.x)
- `openssl` - Command-line tools

#### Red Hat/CentOS/Fedora (YUM/DNF)
- `openssl-devel` - Development files
- `openssl-libs` - Runtime libraries
- `openssl` - Command-line tools

#### Alpine (APK)
- `openssl-dev` - Development files
- `openssl` - Runtime and tools

#### Arch Linux (pacman)
- `openssl` - Includes runtime and development files

### Language-Specific Packages

#### Python (PyPI)
- `pyOpenSSL` - Python wrapper for OpenSSL
- `cryptography` - Uses OpenSSL underneath
- `ssl` - Built-in Python module (uses OpenSSL)

#### Ruby (RubyGems)
- `openssl` - Standard library (uses system OpenSSL)

#### Node.js (NPM)
- Built-in `crypto`, `tls`, `https` modules (use OpenSSL)
- `node-openssl-cert` - Certificate utilities
- `openssl-wrapper` - Command-line wrapper

#### Go
- `crypto/tls` - Standard library (can use OpenSSL via CGO)
- `crypto/x509` - X.509 certificate handling

#### Rust (Cargo)
- `openssl` - OpenSSL bindings
- `openssl-sys` - Low-level OpenSSL bindings
- `native-tls` - Can use OpenSSL as backend

#### Java
- Uses system OpenSSL via JNI
- `conscrypt` - Google's OpenSSL provider for Java
- `wildfly-openssl` - OpenSSL bindings for Java

#### PHP
- `openssl` - Built-in PHP extension
- Compiled with `--with-openssl`

#### Perl
- `Net::SSLeay` - Perl bindings for OpenSSL
- `IO::Socket::SSL` - SSL sockets using Net::SSLeay

## C/C++ Header Includes

### Core Headers
```c
#include <openssl/ssl.h>
#include <openssl/err.h>
#include <openssl/crypto.h>
#include <openssl/opensslv.h>
```

### Cryptographic Primitives
```c
#include <openssl/evp.h>        // Envelope functions (high-level crypto)
#include <openssl/aes.h>        // AES encryption
#include <openssl/des.h>        // DES encryption
#include <openssl/rsa.h>        // RSA public key
#include <openssl/dsa.h>        // DSA public key
#include <openssl/dh.h>         // Diffie-Hellman
#include <openssl/ec.h>         // Elliptic curve
#include <openssl/ecdsa.h>      // ECDSA signatures
#include <openssl/ecdh.h>       // ECDH key exchange
```

### Hashing and MAC
```c
#include <openssl/md5.h>        // MD5 hash
#include <openssl/sha.h>        // SHA hash family
#include <openssl/hmac.h>       // HMAC
```

### Certificates and Keys
```c
#include <openssl/x509.h>       // X.509 certificates
#include <openssl/x509v3.h>     // X.509 v3 extensions
#include <openssl/pem.h>        // PEM encoding
#include <openssl/pkcs12.h>     // PKCS#12
#include <openssl/pkcs7.h>      // PKCS#7
#include <openssl/asn1.h>       // ASN.1
```

### Random Number Generation
```c
#include <openssl/rand.h>       // Random number generation
```

### BIO (I/O Abstraction)
```c
#include <openssl/bio.h>        // Basic I/O abstraction
```

### Configuration and Engine
```c
#include <openssl/conf.h>       // Configuration
#include <openssl/engine.h>     // Engine interface
```

## Python Import Patterns

### PyOpenSSL
```python
import OpenSSL
from OpenSSL import SSL, crypto
from OpenSSL.SSL import Context, Connection
from OpenSSL.crypto import (
    X509, X509Name, X509Store, X509Req,
    PKey, TYPE_RSA, TYPE_DSA,
    load_certificate, dump_certificate,
    FILETYPE_PEM, FILETYPE_ASN1
)
```

### Cryptography Library
```python
from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa, ec, padding
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.backends.openssl import backend
```

### Built-in SSL Module
```python
import ssl
import hashlib
import hmac

# SSL context creation
context = ssl.SSLContext(ssl.PROTOCOL_TLS)
context = ssl.create_default_context()
```

## Ruby Import Patterns

```ruby
require 'openssl'
require 'openssl/ssl'
require 'openssl/x509'

# Common classes
OpenSSL::SSL::SSLContext
OpenSSL::SSL::SSLSocket
OpenSSL::X509::Certificate
OpenSSL::X509::Name
OpenSSL::PKey::RSA
OpenSSL::PKey::EC
OpenSSL::Cipher
OpenSSL::Digest
OpenSSL::HMAC
```

## Node.js Import Patterns

```javascript
// Built-in modules (use OpenSSL)
const crypto = require('crypto');
const tls = require('tls');
const https = require('https');

// ES6 imports
import crypto from 'crypto';
import tls from 'tls';
import https from 'https';
```

## Go Import Patterns

```go
import (
    "crypto/tls"
    "crypto/x509"
    "crypto/rsa"
    "crypto/ecdsa"
    "crypto/rand"
    "crypto/sha256"
    "crypto/hmac"
    "crypto/cipher"
)
```

## Rust Import Patterns

```rust
use openssl::ssl::{SslContext, SslMethod, SslConnector};
use openssl::x509::X509;
use openssl::pkey::PKey;
use openssl::rsa::Rsa;
use openssl::hash::MessageDigest;
use openssl::sign::{Signer, Verifier};
use openssl::encrypt::{Encrypter, Decrypter};
use openssl::symm::{Cipher, encrypt, decrypt};
```

## Java Import Patterns

```java
// Standard Java SSL/TLS
import javax.net.ssl.*;
import java.security.cert.*;
import java.security.*;

// Conscrypt (Google's OpenSSL provider)
import org.conscrypt.*;

// WildFly OpenSSL
import org.wildfly.openssl.*;
```

## PHP Extension Patterns

### php.ini Configuration
```ini
extension=openssl.so
extension=openssl.dll
```

### PHP Code
```php
// OpenSSL extension functions
openssl_encrypt()
openssl_decrypt()
openssl_sign()
openssl_verify()
openssl_public_encrypt()
openssl_private_decrypt()
openssl_x509_parse()
openssl_get_cert_locations()
```

## Perl Import Patterns

```perl
use Net::SSLeay;
use IO::Socket::SSL;
use Crypt::OpenSSL::RSA;
use Crypt::OpenSSL::X509;
use Crypt::OpenSSL::Random;
```

## Build System Patterns

### CMake
```cmake
find_package(OpenSSL REQUIRED)
include_directories(${OPENSSL_INCLUDE_DIR})
target_link_libraries(myapp ${OPENSSL_LIBRARIES})

# Or
find_package(OpenSSL 1.1.1 REQUIRED)
target_link_libraries(myapp OpenSSL::SSL OpenSSL::Crypto)
```

### Makefile
```makefile
CFLAGS += -I/usr/include/openssl
LDFLAGS += -lssl -lcrypto
```

### pkg-config
```bash
pkg-config --cflags openssl
pkg-config --libs openssl
pkg-config --modversion openssl
```

### Autotools (configure.ac)
```autoconf
PKG_CHECK_MODULES([OPENSSL], [openssl >= 1.1.1])
AC_CHECK_LIB([ssl], [SSL_library_init])
AC_CHECK_HEADERS([openssl/ssl.h])
```

### Cargo.toml (Rust)
```toml
[dependencies]
openssl = "0.10"

[build-dependencies]
openssl-sys = "0.9"
```

### requirements.txt (Python)
```
pyOpenSSL==23.0.0
cryptography==41.0.0
```

### Gemfile (Ruby)
```ruby
# OpenSSL is part of Ruby standard library
# No gem needed unless specific version required
```

### package.json (Node.js)
```json
{
  "dependencies": {
    "node-openssl-cert": "^1.0.0"
  }
}
```

## Linking Patterns

### Dynamic Linking
```bash
# Linux
-lssl -lcrypto

# macOS
-lssl -lcrypto

# Windows
ssleay32.lib libeay32.lib  # OpenSSL 1.0.x
libssl.lib libcrypto.lib   # OpenSSL 1.1.x+
```

### Static Linking
```bash
libssl.a libcrypto.a
```

## Dockerfile Patterns

```dockerfile
# Debian/Ubuntu
RUN apt-get update && apt-get install -y \
    libssl-dev \
    openssl

# Alpine
RUN apk add --no-cache \
    openssl \
    openssl-dev

# Red Hat/CentOS
RUN yum install -y \
    openssl-devel \
    openssl
```

## Detection Confidence

- **HIGH**: Direct OpenSSL header includes in C/C++
- **HIGH**: pyOpenSSL or cryptography imports in Python
- **HIGH**: OpenSSL module usage in Ruby
- **HIGH**: openssl package in system dependencies
- **MEDIUM**: Crypto/TLS modules that may use OpenSSL
- **MEDIUM**: SSL/TLS configuration that suggests OpenSSL
- **LOW**: Generic cryptographic code without clear OpenSSL markers
