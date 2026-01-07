# Test file: Weak cryptography patterns that SHOULD be detected
# These demonstrate insecure cryptographic practices

import hashlib
from Crypto.Cipher import DES, AES, Blowfish
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms
import ssl

# Weak hashing algorithms (should detect)
md5_hash = hashlib.md5(b"password").hexdigest()
sha1_hash = hashlib.sha1(b"data").hexdigest()

# DES - weak cipher (should detect)
des_cipher = DES.new(b"12345678", DES.MODE_ECB)

# ECB mode - insecure (should detect)
aes_ecb = AES.new(key, AES.MODE_ECB)

# Hardcoded encryption key (should detect)
ENCRYPTION_KEY = b"hardcoded_secret_key_12345"
cipher = AES.new(ENCRYPTION_KEY, AES.MODE_CBC, iv)

# Weak key size (should detect)
small_key = os.urandom(8)  # Only 64 bits

# Insecure random (should detect)
import random
token = ''.join(random.choice('abcdef0123456789') for _ in range(32))

# SSL/TLS issues (should detect)
ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLSv1)  # TLS 1.0 is deprecated
ssl_context.check_hostname = False
ssl_context.verify_mode = ssl.CERT_NONE  # Disabling certificate verification

# RC4 - broken cipher (should detect)
rc4_cipher = algorithms.ARC4(key)

# Blowfish with small key (should detect)
bf = Blowfish.new(b"shortkey", Blowfish.MODE_CBC)

# No IV/nonce reuse protection
fixed_iv = b"1234567890123456"  # Using fixed IV
