# TLS Misconfiguration

**Category**: cryptography/tls-misconfig
**Description**: Detection of insecure TLS/SSL configurations
**CWE**: CWE-295 (Improper Certificate Validation), CWE-757 (Selection of Less-Secure Algorithm)

---

## Import Detection

### Python
**Pattern**: `ssl\._create_unverified_context`
- Disabled certificate verification
- Example: `ssl._create_unverified_context()`

**Pattern**: `verify\s*=\s*False`
- Requests/urllib3 cert verification disabled
- Example: `requests.get(url, verify=False)`

**Pattern**: `CERT_NONE`
- SSL context with no cert verification
- Example: `context.verify_mode = ssl.CERT_NONE`

**Pattern**: `check_hostname\s*=\s*False`
- Hostname verification disabled
- Example: `context.check_hostname = False`

**Pattern**: `verify_mode\s*=\s*ssl\.CERT_NONE`
- Explicit CERT_NONE assignment
- Example: `ctx.verify_mode = ssl.CERT_NONE`

**Pattern**: `ssl\.PROTOCOL_SSLv2`
- SSLv2 protocol (broken)
- Example: `ssl.wrap_socket(sock, ssl_version=ssl.PROTOCOL_SSLv2)`

**Pattern**: `ssl\.PROTOCOL_SSLv3`
- SSLv3 protocol (broken - POODLE)
- Example: `ssl.wrap_socket(sock, ssl_version=ssl.PROTOCOL_SSLv3)`

**Pattern**: `ssl\.PROTOCOL_TLSv1\b`
- TLS 1.0 protocol (deprecated)
- Example: `ssl.wrap_socket(sock, ssl_version=ssl.PROTOCOL_TLSv1)`

**Pattern**: `ssl\.PROTOCOL_TLSv1_1`
- TLS 1.1 protocol (deprecated)
- Example: `ssl.wrap_socket(sock, ssl_version=ssl.PROTOCOL_TLSv1_1)`

**Pattern**: `urllib3\.disable_warnings`
- Disabling SSL warnings
- Example: `urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)`

### Javascript
**Pattern**: `rejectUnauthorized\s*:\s*false`
- Node.js TLS cert verification disabled
- Example: `{ rejectUnauthorized: false }`

**Pattern**: `NODE_TLS_REJECT_UNAUTHORIZED.*['"]?0['"]?`
- Environment variable disabling TLS verification
- Example: `process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0'`

**Pattern**: `agent:\s*new https\.Agent\(.*rejectUnauthorized`
- Custom agent with disabled verification
- Example: `agent: new https.Agent({ rejectUnauthorized: false })`

**Pattern**: `minVersion\s*:\s*['"]TLSv1['"]`
- Deprecated TLS 1.0 as minimum
- Example: `{ minVersion: 'TLSv1' }`

**Pattern**: `minVersion\s*:\s*['"]TLSv1\.1['"]`
- Deprecated TLS 1.1 as minimum
- Example: `{ minVersion: 'TLSv1.1' }`

**Pattern**: `secureProtocol\s*:\s*['"]SSLv3`
- Broken SSLv3 protocol
- Example: `{ secureProtocol: 'SSLv3_method' }`

**Pattern**: `secureProtocol\s*:\s*['"]SSLv2`
- Broken SSLv2 protocol
- Example: `{ secureProtocol: 'SSLv2_method' }`

**Pattern**: `secureProtocol\s*:\s*['"]TLSv1_method`
- Deprecated TLS 1.0
- Example: `{ secureProtocol: 'TLSv1_method' }`

**Pattern**: `checkServerIdentity:\s*\(\)\s*=>`
- Empty hostname verification callback
- Example: `checkServerIdentity: () => undefined`

### Java
**Pattern**: `setHostnameVerifier\(.*ALLOW_ALL`
- Hostname verification disabled (Apache HttpClient)
- Example: `setHostnameVerifier(SSLConnectionSocketFactory.ALLOW_ALL_HOSTNAME_VERIFIER)`

**Pattern**: `setHostnameVerifier\(.*NoopHostnameVerifier`
- Noop hostname verifier
- Example: `setHostnameVerifier(NoopHostnameVerifier.INSTANCE)`

**Pattern**: `TrustManager.*X509Certificate.*return`
- Custom trust manager accepting all certs
- Example: `public X509Certificate[] getAcceptedIssuers() { return new X509Certificate[0]; }`

**Pattern**: `checkServerTrusted.*\{\s*\}`
- Empty trust check implementation
- Example: `public void checkServerTrusted(X509Certificate[] chain, String authType) { }`

**Pattern**: `checkClientTrusted.*\{\s*\}`
- Empty client trust check
- Example: `public void checkClientTrusted(X509Certificate[] chain, String authType) { }`

**Pattern**: `SSLContext\.getInstance\(["']SSL["']\)`
- Generic SSL protocol (may use SSLv3)
- Example: `SSLContext.getInstance("SSL")`

**Pattern**: `SSLContext\.getInstance\(["']SSLv3["']\)`
- Broken SSLv3
- Example: `SSLContext.getInstance("SSLv3")`

**Pattern**: `SSLContext\.getInstance\(["']TLSv1["']\)`
- Deprecated TLS 1.0
- Example: `SSLContext.getInstance("TLSv1")`

**Pattern**: `SSLContext\.getInstance\(["']TLSv1\.1["']\)`
- Deprecated TLS 1.1
- Example: `SSLContext.getInstance("TLSv1.1")`

**Pattern**: `setEnabledProtocols.*SSLv3`
- Enabling SSLv3
- Example: `socket.setEnabledProtocols(new String[] { "SSLv3" })`

**Pattern**: `setEnabledProtocols.*TLSv1[^.]`
- Enabling TLS 1.0
- Example: `socket.setEnabledProtocols(new String[] { "TLSv1" })`

### Go
**Pattern**: `InsecureSkipVerify\s*:\s*true`
- Go TLS skip certificate verification
- Example: `InsecureSkipVerify: true`

**Pattern**: `MinVersion\s*:\s*tls\.VersionSSL30`
- Broken SSL 3.0 in Go
- Example: `MinVersion: tls.VersionSSL30`

**Pattern**: `MinVersion\s*:\s*tls\.VersionTLS10`
- Deprecated TLS 1.0 in Go
- Example: `MinVersion: tls.VersionTLS10`

**Pattern**: `MinVersion\s*:\s*tls\.VersionTLS11`
- Deprecated TLS 1.1 in Go
- Example: `MinVersion: tls.VersionTLS11`

**Pattern**: `MaxVersion\s*:\s*tls\.VersionTLS10`
- Maximum TLS 1.0 (should support higher)
- Example: `MaxVersion: tls.VersionTLS10`

**Pattern**: `VerifyPeerCertificate:\s*func.*nil`
- Empty certificate verification function
- Example: `VerifyPeerCertificate: func([][]byte, [][]*x509.Certificate) error { return nil }`

**Pattern**: `ServerName:\s*["']["']`
- Empty server name (disables SNI verification)
- Example: `ServerName: ""`

### Ruby
**Pattern**: `verify_mode\s*=\s*OpenSSL::SSL::VERIFY_NONE`
- Certificate verification disabled
- Example: `http.verify_mode = OpenSSL::SSL::VERIFY_NONE`

**Pattern**: `VERIFY_NONE`
- VERIFY_NONE constant usage
- Example: `ssl_context.verify_mode = OpenSSL::SSL::VERIFY_NONE`

**Pattern**: `ssl_version\s*=\s*['":]*SSLv3`
- SSLv3 in Ruby
- Example: `http.ssl_version = :SSLv3`

**Pattern**: `ssl_version\s*=\s*['":]*TLSv1\b`
- TLS 1.0 in Ruby
- Example: `http.ssl_version = :TLSv1`

### PHP
**Pattern**: `CURLOPT_SSL_VERIFYPEER.*false`
- cURL SSL verification disabled
- Example: `curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false)`

**Pattern**: `CURLOPT_SSL_VERIFYHOST.*0`
- cURL hostname verification disabled
- Example: `curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, 0)`

**Pattern**: `verify_peer.*false`
- Stream context SSL verification disabled
- Example: `'verify_peer' => false`

**Pattern**: `verify_peer_name.*false`
- Stream context peer name verification disabled
- Example: `'verify_peer_name' => false`

**Pattern**: `allow_self_signed.*true`
- Allowing self-signed certificates
- Example: `'allow_self_signed' => true`

### C/C++
**Pattern**: `SSL_CTX_set_verify.*SSL_VERIFY_NONE`
- OpenSSL verification disabled
- Example: `SSL_CTX_set_verify(ctx, SSL_VERIFY_NONE, NULL)`

**Pattern**: `SSL_set_verify.*SSL_VERIFY_NONE`
- SSL connection verification disabled
- Example: `SSL_set_verify(ssl, SSL_VERIFY_NONE, NULL)`

**Pattern**: `SSL_CTX_set_min_proto_version.*SSL3_VERSION`
- SSLv3 minimum version
- Example: `SSL_CTX_set_min_proto_version(ctx, SSL3_VERSION)`

**Pattern**: `SSL_CTX_set_min_proto_version.*TLS1_VERSION\b`
- TLS 1.0 minimum version
- Example: `SSL_CTX_set_min_proto_version(ctx, TLS1_VERSION)`

### C#
**Pattern**: `ServicePointManager\.ServerCertificateValidationCallback.*true`
- Certificate validation callback always returns true
- Example: `ServicePointManager.ServerCertificateValidationCallback = (s, c, ch, e) => true`

**Pattern**: `ServerCertificateCustomValidationCallback.*true`
- HttpClient certificate validation disabled
- Example: `ServerCertificateCustomValidationCallback = (msg, cert, chain, errors) => true`

**Pattern**: `SecurityProtocolType\.Ssl3`
- SSLv3 in .NET
- Example: `ServicePointManager.SecurityProtocol = SecurityProtocolType.Ssl3`

**Pattern**: `SecurityProtocolType\.Tls\b`
- TLS 1.0 only in .NET
- Example: `ServicePointManager.SecurityProtocol = SecurityProtocolType.Tls`

---

## Secrets Detection

#### Self-Signed Certificate Bypass
**Pattern**: `(?:self.signed|selfsigned|self_signed)\s*[=:]\s*(?:true|True|1)`
**Severity**: high
**Description**: Explicitly allowing self-signed certificates

#### Certificate Pinning Disabled
**Pattern**: `(?:pinning|PINNING|pin)\s*[=:]\s*(?:false|False|0|disabled|none)`
**Severity**: medium
**Description**: Certificate pinning explicitly disabled

#### Trust All Certificates Comment
**Pattern**: `(?:trust|accept)\s*all\s*(?:cert|certificate)`
**Severity**: high
**Description**: Code comment indicating trust-all behavior

---

## Environment Variables

- `SSL_CERT_FILE`
- `SSL_CERT_DIR`
- `REQUESTS_CA_BUNDLE`
- `CURL_CA_BUNDLE`
- `NODE_TLS_REJECT_UNAUTHORIZED`
- `NODE_EXTRA_CA_CERTS`
- `PYTHONHTTPSVERIFY`

---

## Detection Confidence

**Certificate Validation Bypass**: 98%
**Protocol Version Detection**: 95%
**Hostname Verification Bypass**: 95%
