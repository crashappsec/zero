# Weak Cryptographic Ciphers

**Category**: cryptography/weak-ciphers
**Description**: Detection of weak, deprecated, or broken cryptographic algorithms
**CWE**: CWE-327 (Use of a Broken or Risky Cryptographic Algorithm)

---

## Import Detection

### Python
**Pattern**: `from Crypto\.Cipher import DES`
- DES cipher import (broken - 56-bit key)
- Example: `from Crypto.Cipher import DES`

**Pattern**: `from Crypto\.Cipher import Blowfish`
- Blowfish cipher import (deprecated)
- Example: `from Crypto.Cipher import Blowfish`

**Pattern**: `from Crypto\.Cipher import ARC4`
- RC4 cipher import (broken)
- Example: `from Crypto.Cipher import ARC4`

**Pattern**: `hashlib\.md5\(`
- MD5 hash usage (broken for security)
- Example: `hashlib.md5(password.encode())`

**Pattern**: `hashlib\.sha1\(`
- SHA1 hash usage (deprecated for security)
- Example: `hashlib.sha1(data.encode())`

**Pattern**: `AES\.MODE_ECB`
- ECB mode usage (insecure - no IV, patterns visible)
- Example: `AES.new(key, AES.MODE_ECB)`

**Pattern**: `DES3\.MODE_ECB`
- Triple DES with ECB mode
- Example: `DES3.new(key, DES3.MODE_ECB)`

### Javascript
**Pattern**: `crypto\.createCipher\(['"]des`
- DES cipher in Node.js
- Example: `crypto.createCipher('des', key)`

**Pattern**: `crypto\.createCipheriv\(['"]des`
- DES cipher with IV in Node.js
- Example: `crypto.createCipheriv('des-cbc', key, iv)`

**Pattern**: `crypto\.createHash\(['"]md5`
- MD5 hash in Node.js
- Example: `crypto.createHash('md5')`

**Pattern**: `crypto\.createHash\(['"]sha1`
- SHA1 hash in Node.js
- Example: `crypto.createHash('sha1')`

**Pattern**: `CryptoJS\.DES`
- DES in CryptoJS library
- Example: `CryptoJS.DES.encrypt(message, key)`

**Pattern**: `CryptoJS\.RC4`
- RC4 in CryptoJS (broken)
- Example: `CryptoJS.RC4.encrypt(message, key)`

**Pattern**: `CryptoJS\.TripleDES`
- Triple DES (deprecated)
- Example: `CryptoJS.TripleDES.encrypt(message, key)`

**Pattern**: `CryptoJS\.MD5`
- MD5 in CryptoJS
- Example: `CryptoJS.MD5(message)`

**Pattern**: `CryptoJS\.SHA1`
- SHA1 in CryptoJS
- Example: `CryptoJS.SHA1(message)`

**Pattern**: `mode:\s*CryptoJS\.mode\.ECB`
- ECB mode in CryptoJS
- Example: `{ mode: CryptoJS.mode.ECB }`

### Java
**Pattern**: `Cipher\.getInstance\(["']DES`
- DES cipher in Java
- Example: `Cipher.getInstance("DES")`

**Pattern**: `Cipher\.getInstance\(["']DESede`
- Triple DES in Java (deprecated)
- Example: `Cipher.getInstance("DESede")`

**Pattern**: `Cipher\.getInstance\(["'].*ECB`
- ECB mode in Java
- Example: `Cipher.getInstance("AES/ECB/PKCS5Padding")`

**Pattern**: `Cipher\.getInstance\(["']RC4`
- RC4 cipher in Java
- Example: `Cipher.getInstance("RC4")`

**Pattern**: `Cipher\.getInstance\(["']ARCFOUR`
- ARCFOUR (RC4) in Java
- Example: `Cipher.getInstance("ARCFOUR")`

**Pattern**: `Cipher\.getInstance\(["']Blowfish`
- Blowfish in Java
- Example: `Cipher.getInstance("Blowfish")`

**Pattern**: `MessageDigest\.getInstance\(["']MD5`
- MD5 in Java
- Example: `MessageDigest.getInstance("MD5")`

**Pattern**: `MessageDigest\.getInstance\(["']SHA-1`
- SHA1 in Java
- Example: `MessageDigest.getInstance("SHA-1")`

**Pattern**: `MessageDigest\.getInstance\(["']SHA1`
- SHA1 in Java (alternate form)
- Example: `MessageDigest.getInstance("SHA1")`

### Go
**Pattern**: `des\.NewCipher`
- DES cipher in Go
- Example: `des.NewCipher(key)`

**Pattern**: `des\.NewTripleDESCipher`
- Triple DES in Go
- Example: `des.NewTripleDESCipher(key)`

**Pattern**: `md5\.New\(\)`
- MD5 in Go
- Example: `h := md5.New()`

**Pattern**: `md5\.Sum\(`
- MD5 sum in Go
- Example: `md5.Sum(data)`

**Pattern**: `sha1\.New\(\)`
- SHA1 in Go
- Example: `h := sha1.New()`

**Pattern**: `sha1\.Sum\(`
- SHA1 sum in Go
- Example: `sha1.Sum(data)`

**Pattern**: `rc4\.NewCipher`
- RC4 in Go (broken)
- Example: `rc4.NewCipher(key)`

**Pattern**: `cipher\.NewCBCEncrypter.*des`
- DES with CBC in Go
- Example: `cipher.NewCBCEncrypter(block, iv)`

### Ruby
**Pattern**: `OpenSSL::Cipher\.new\(['"]des`
- DES cipher in Ruby
- Example: `OpenSSL::Cipher.new('des-cbc')`

**Pattern**: `OpenSSL::Cipher\.new\(['"]rc4`
- RC4 cipher in Ruby
- Example: `OpenSSL::Cipher.new('rc4')`

**Pattern**: `Digest::MD5`
- MD5 in Ruby
- Example: `Digest::MD5.hexdigest(data)`

**Pattern**: `Digest::SHA1`
- SHA1 in Ruby
- Example: `Digest::SHA1.hexdigest(data)`

### PHP
**Pattern**: `mcrypt_encrypt\(.*MCRYPT_DES`
- DES in PHP mcrypt (deprecated API)
- Example: `mcrypt_encrypt(MCRYPT_DES, $key, $data, MCRYPT_MODE_CBC)`

**Pattern**: `openssl_encrypt\(.*['"]des`
- DES in PHP OpenSSL
- Example: `openssl_encrypt($data, 'des-cbc', $key)`

**Pattern**: `openssl_encrypt\(.*['"]rc4`
- RC4 in PHP
- Example: `openssl_encrypt($data, 'rc4', $key)`

**Pattern**: `md5\(`
- MD5 function in PHP
- Example: `md5($password)`

**Pattern**: `sha1\(`
- SHA1 function in PHP
- Example: `sha1($data)`

**Pattern**: `MCRYPT_MODE_ECB`
- ECB mode in PHP
- Example: `mcrypt_encrypt($cipher, $key, $data, MCRYPT_MODE_ECB)`

### C/C++
**Pattern**: `DES_set_key`
- DES in OpenSSL C API
- Example: `DES_set_key(&key, &schedule)`

**Pattern**: `DES_ecb_encrypt`
- DES ECB in OpenSSL
- Example: `DES_ecb_encrypt(&input, &output, &schedule, DES_ENCRYPT)`

**Pattern**: `MD5_Init`
- MD5 in OpenSSL C API
- Example: `MD5_Init(&ctx)`

**Pattern**: `SHA1_Init`
- SHA1 in OpenSSL C API
- Example: `SHA1_Init(&ctx)`

**Pattern**: `EVP_des_`
- DES EVP functions
- Example: `EVP_des_cbc()`

**Pattern**: `EVP_rc4`
- RC4 EVP function
- Example: `EVP_rc4()`

**Pattern**: `EVP_md5`
- MD5 EVP function
- Example: `EVP_md5()`

**Pattern**: `EVP_sha1`
- SHA1 EVP function
- Example: `EVP_sha1()`

---

## Secrets Detection

#### Hardcoded DES Key
**Pattern**: `DES\.new\(['"bx][0-9a-fA-F]{16}['"]`
**Severity**: critical
**Description**: Hardcoded DES key in source code

#### Hardcoded RC4 Key
**Pattern**: `[Rr][Cc]4.*key\s*[=:]\s*['"][^'"]{8,}['"]`
**Severity**: critical
**Description**: Hardcoded RC4 key in source code

#### Hardcoded 3DES Key
**Pattern**: `DES3\.new\(['"bx][0-9a-fA-F]{32,48}['"]`
**Severity**: critical
**Description**: Hardcoded Triple DES key in source code

---

## Detection Confidence

**Import Detection**: 95%
**Code Pattern Detection**: 90%
**Configuration Detection**: 85%
