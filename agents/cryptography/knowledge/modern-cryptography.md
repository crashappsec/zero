# Modern Cryptography Best Practices

## Recommended Algorithms

### Symmetric Encryption
| Use Case | Algorithm | Notes |
|----------|-----------|-------|
| General encryption | AES-256-GCM | Authenticated encryption, 256-bit key |
| High-performance | ChaCha20-Poly1305 | Fast in software, good for mobile |
| Disk encryption | AES-XTS | For full disk encryption |

### Hash Functions
| Use Case | Algorithm | Notes |
|----------|-----------|-------|
| General hashing | SHA-256 | NIST approved, widely supported |
| Higher security | SHA-384, SHA-512 | When 256-bit isn't enough |
| Modern alternative | SHA-3, BLAKE3 | Post-SHA-2 options |
| Password hashing | Argon2id | Winner of PHC, memory-hard |
| Legacy compatible | bcrypt | Still acceptable for passwords |
| MAC | HMAC-SHA-256 | Message authentication |

### Asymmetric Encryption
| Use Case | Algorithm | Key Size |
|----------|-----------|----------|
| Key exchange | X25519 | 256-bit curve |
| Signing | Ed25519 | 256-bit curve |
| Legacy RSA | RSA-OAEP | >= 2048 bits, prefer 4096 |
| TLS/Certificates | ECDSA P-256 | Or Ed25519 if supported |

### Key Derivation
| Use Case | Algorithm | Notes |
|----------|-----------|-------|
| Password → Key | Argon2id | Best choice |
| Password → Key | PBKDF2-SHA256 | If Argon2 unavailable, high iterations |
| Key stretching | HKDF | Expand a key to multiple |
| Legacy | scrypt | Memory-hard, pre-Argon2 |

## Deprecated/Broken Algorithms

### DO NOT USE - Broken
| Algorithm | Why | Replace With |
|-----------|-----|--------------|
| DES | 56-bit key, trivially brute-forceable | AES-256 |
| RC4 | Multiple attacks (BEAST, NOMORE) | AES-GCM, ChaCha20 |
| MD5 | Collision attacks, forgery | SHA-256 |
| SHA1 | Collision attacks demonstrated | SHA-256 |
| ECB mode | Patterns visible in ciphertext | GCM, CBC with random IV |

### Deprecated - Avoid in New Code
| Algorithm | Why | Replace With |
|-----------|-----|--------------|
| 3DES | Slow, 112-bit effective security | AES-256 |
| Blowfish | 64-bit block, birthday attacks | AES |
| RIPEMD | Less scrutinized than SHA | SHA-256 |
| DSA | Complex, error-prone | Ed25519, ECDSA |

## Minimum Key Lengths

| Algorithm | Minimum | Recommended |
|-----------|---------|-------------|
| RSA | 2048 bits | 4096 bits |
| ECC | 256 bits (P-256) | 256+ bits |
| AES | 128 bits | 256 bits |
| Symmetric MAC | 256 bits | 256 bits |

## TLS Configuration

### Required
- TLS 1.2 minimum (TLS 1.3 preferred)
- Certificate verification enabled
- Hostname verification enabled
- Strong cipher suites only

### Recommended TLS 1.3 Cipher Suites
```
TLS_AES_256_GCM_SHA384
TLS_CHACHA20_POLY1305_SHA256
TLS_AES_128_GCM_SHA256
```

### Recommended TLS 1.2 Cipher Suites
```
ECDHE-ECDSA-AES256-GCM-SHA384
ECDHE-RSA-AES256-GCM-SHA384
ECDHE-ECDSA-CHACHA20-POLY1305
ECDHE-RSA-CHACHA20-POLY1305
```

### Disable
- SSLv2, SSLv3 (broken)
- TLS 1.0, TLS 1.1 (deprecated)
- NULL ciphers
- Export ciphers
- RC4 ciphers
- DES/3DES ciphers

## Random Number Generation

### Cryptographically Secure
| Language | Use |
|----------|-----|
| Python | `secrets` module, `os.urandom()` |
| Node.js | `crypto.randomBytes()`, `crypto.getRandomValues()` |
| Java | `java.security.SecureRandom` |
| Go | `crypto/rand` package |
| Ruby | `SecureRandom` |
| PHP | `random_bytes()`, `random_int()` |
| C/C++ | `/dev/urandom`, `CryptGenRandom` |

### NOT Secure - Never Use for Security
| Language | Avoid |
|----------|-------|
| Python | `random` module |
| JavaScript | `Math.random()` |
| Java | `java.util.Random` |
| Go | `math/rand` package |
| Ruby | `rand()`, `Random` |
| PHP | `rand()`, `mt_rand()` |
| C | `rand()`, `random()` |
