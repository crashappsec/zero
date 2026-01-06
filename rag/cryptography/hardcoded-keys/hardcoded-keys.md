# Hardcoded Cryptographic Keys

**Category**: cryptography/hardcoded-keys
**Description**: Detection of cryptographic keys embedded in source code
**CWE**: CWE-321 (Use of Hard-coded Cryptographic Key), CWE-798 (Use of Hard-coded Credentials)

---

## Secrets Detection

#### RSA Private Key
**Pattern**: `-----BEGIN RSA PRIVATE KEY-----`
**Severity**: critical
**Description**: RSA private key embedded in source code

#### RSA Private Key (Encrypted)
**Pattern**: `-----BEGIN ENCRYPTED PRIVATE KEY-----`
**Severity**: high
**Description**: Encrypted private key in source (password may be nearby)

#### EC Private Key
**Pattern**: `-----BEGIN EC PRIVATE KEY-----`
**Severity**: critical
**Description**: Elliptic curve private key in source code

#### Generic Private Key (PKCS#8)
**Pattern**: `-----BEGIN PRIVATE KEY-----`
**Severity**: critical
**Description**: PKCS#8 private key in source code

#### DSA Private Key
**Pattern**: `-----BEGIN DSA PRIVATE KEY-----`
**Severity**: critical
**Description**: DSA private key in source code

#### OpenSSH Private Key
**Pattern**: `-----BEGIN OPENSSH PRIVATE KEY-----`
**Severity**: critical
**Description**: OpenSSH format private key in source code

#### PGP Private Key
**Pattern**: `-----BEGIN PGP PRIVATE KEY BLOCK-----`
**Severity**: critical
**Description**: PGP private key block in source code

#### Hardcoded AES Key (Hex 128-bit)
**Pattern**: `(?:aes|AES|encryption).*key\s*[=:]\s*['"]([0-9a-fA-F]{32})['"]`
**Severity**: critical
**Description**: AES-128 key hardcoded as hex string

#### Hardcoded AES Key (Hex 256-bit)
**Pattern**: `(?:aes|AES|encryption).*key\s*[=:]\s*['"]([0-9a-fA-F]{64})['"]`
**Severity**: critical
**Description**: AES-256 key hardcoded as hex string

#### Hardcoded AES Key (Base64)
**Pattern**: `(?:aes|AES|encryption).*key\s*[=:]\s*['"]([A-Za-z0-9+/]{22,44}={0,2})['"]`
**Severity**: critical
**Description**: AES key hardcoded as base64 string

#### Hardcoded IV (Initialization Vector)
**Pattern**: `(?:iv|IV|nonce|NONCE)\s*[=:]\s*['"]([0-9a-fA-F]{16,32})['"]`
**Severity**: high
**Description**: Hardcoded initialization vector (should be random per encryption)

#### Hardcoded IV (Bytes)
**Pattern**: `(?:iv|IV|nonce)\s*[=:]\s*b['"]`
**Severity**: high
**Description**: Hardcoded byte string IV

#### JWT Secret Key
**Pattern**: `(?:jwt|JWT).*(?:secret|SECRET|key|KEY)\s*[=:]\s*['"]([^'"]{16,})['"]`
**Severity**: critical
**Description**: Hardcoded JWT signing secret

#### HMAC Secret Key
**Pattern**: `(?:hmac|HMAC).*(?:secret|SECRET|key|KEY)\s*[=:]\s*['"]([^'"]{16,})['"]`
**Severity**: critical
**Description**: Hardcoded HMAC secret key

#### Generic Encryption Key Variable
**Pattern**: `(?:ENCRYPTION|CRYPTO|SECRET|SIGNING)_KEY\s*[=:]\s*['"]([^'"]{16,})['"]`
**Severity**: critical
**Description**: Generic encryption key in environment/config style

#### Password Salt (Hardcoded)
**Pattern**: `(?:salt|SALT)\s*[=:]\s*['"]([^'"]{8,})['"]`
**Severity**: high
**Description**: Hardcoded salt value (should be random per user)

#### Symmetric Key Bytes
**Pattern**: `key\s*=\s*b['"][^'"]{16,}['"]`
**Severity**: critical
**Description**: Hardcoded symmetric key as byte string

---

## Import Detection

### Python
**Pattern**: `RSA\.generate\(1024\)`
- Weak RSA key generation (should be >= 2048)
- Example: `key = RSA.generate(1024)`

**Pattern**: `RSA\.generate\(512\)`
- Very weak RSA key generation (trivially breakable)
- Example: `key = RSA.generate(512)`

**Pattern**: `rsa\.generate_private_key\(.*key_size=1024`
- Weak RSA in cryptography library
- Example: `rsa.generate_private_key(public_exponent=65537, key_size=1024)`

**Pattern**: `dsa\.generate_private_key\(.*key_size=1024`
- Weak DSA key size
- Example: `dsa.generate_private_key(key_size=1024)`

### Java
**Pattern**: `keyGenerator\.init\(1024\)`
- Weak RSA key size in Java KeyGenerator
- Example: `keyGenerator.init(1024)`

**Pattern**: `keyGenerator\.init\(512\)`
- Very weak RSA key size
- Example: `keyGenerator.init(512)`

**Pattern**: `KeyPairGenerator.*initialize\(1024\)`
- Weak KeyPairGenerator initialization
- Example: `keyPairGen.initialize(1024)`

**Pattern**: `KeyPairGenerator.*initialize\(512\)`
- Very weak KeyPairGenerator initialization
- Example: `keyPairGen.initialize(512)`

### Go
**Pattern**: `rsa\.GenerateKey\(.*1024\)`
- Weak RSA key in Go
- Example: `rsa.GenerateKey(rand.Reader, 1024)`

**Pattern**: `rsa\.GenerateKey\(.*512\)`
- Very weak RSA key in Go
- Example: `rsa.GenerateKey(rand.Reader, 512)`

### Javascript
**Pattern**: `modulusLength:\s*1024`
- Weak RSA modulus length in Web Crypto API
- Example: `{ modulusLength: 1024 }`

**Pattern**: `modulusLength:\s*512`
- Very weak RSA modulus length
- Example: `{ modulusLength: 512 }`

### Ruby
**Pattern**: `OpenSSL::PKey::RSA\.new\(1024\)`
- Weak RSA key in Ruby
- Example: `OpenSSL::PKey::RSA.new(1024)`

**Pattern**: `OpenSSL::PKey::RSA\.generate\(1024\)`
- Weak RSA generation in Ruby
- Example: `OpenSSL::PKey::RSA.generate(1024)`

### PHP
**Pattern**: `openssl_pkey_new\(.*['"]private_key_bits['"].*1024`
- Weak RSA in PHP
- Example: `openssl_pkey_new(['private_key_bits' => 1024])`

---

## Detection Confidence

**Private Key Detection**: 99%
**Hardcoded Key Pattern**: 90%
**Weak Key Length**: 95%
**IV/Salt Detection**: 85%
