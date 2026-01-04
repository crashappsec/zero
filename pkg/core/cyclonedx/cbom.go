package cyclonedx

import (
	"fmt"
	"strings"
)

// CryptoProperties represents CycloneDX CBOM cryptoProperties object
type CryptoProperties struct {
	AssetType                       string                           `json:"assetType"`
	AlgorithmProperties             *AlgorithmProperties             `json:"algorithmProperties,omitempty"`
	CertificateProperties           *CertificateProperties           `json:"certificateProperties,omitempty"`
	ProtocolProperties              *ProtocolProperties              `json:"protocolProperties,omitempty"`
	RelatedCryptoMaterialProperties *RelatedCryptoMaterialProperties `json:"relatedCryptoMaterialProperties,omitempty"`
	OID                             string                           `json:"oid,omitempty"`
}

// CryptoAssetType constants
const (
	CryptoAssetAlgorithm            = "algorithm"
	CryptoAssetCertificate          = "certificate"
	CryptoAssetProtocol             = "protocol"
	CryptoAssetRelatedCryptoMaterial = "related-crypto-material"
)

// AlgorithmProperties contains algorithm-specific properties
type AlgorithmProperties struct {
	Primitive              string   `json:"primitive,omitempty"`              // ae, mac, hash, pke, kem, dsa, xof, kdf, other
	ParameterSetIdentifier string   `json:"parameterSetIdentifier,omitempty"` // key size (128, 256, 2048)
	Mode                   string   `json:"mode,omitempty"`                   // gcm, cbc, ctr, ecb
	ExecutionEnvironment   string   `json:"executionEnvironment,omitempty"`   // software-plain-ram, hardware, tee, hsm, tpm
	ImplementationPlatform string   `json:"implementationPlatform,omitempty"` // x86_64, arm64
	CertificationLevel     []string `json:"certificationLevel,omitempty"`     // fips-140-2, fips-140-3, common-criteria
	CryptoFunctions        []string `json:"cryptoFunctions,omitempty"`        // keygen, encrypt, decrypt, sign, verify, hash
	ClassicalSecurityLevel int      `json:"classicalSecurityLevel,omitempty"` // bits of security
	NISTQuantumSecurityLevel int    `json:"nistQuantumSecurityLevel,omitempty"` // 1-5
}

// CryptoPrimitive constants
const (
	PrimitiveAE    = "ae"    // Authenticated Encryption
	PrimitiveMAC   = "mac"   // Message Authentication Code
	PrimitiveHash  = "hash"  // Hash Function
	PrimitivePKE   = "pke"   // Public Key Encryption
	PrimitiveKEM   = "kem"   // Key Encapsulation Mechanism
	PrimitiveDSA   = "dsa"   // Digital Signature Algorithm
	PrimitiveXOF   = "xof"   // Extendable Output Function
	PrimitiveKDF   = "kdf"   // Key Derivation Function
	PrimitiveOther = "other"
)

// CryptoFunction constants
const (
	FunctionKeyGen      = "keygen"
	FunctionEncrypt     = "encrypt"
	FunctionDecrypt     = "decrypt"
	FunctionSign        = "sign"
	FunctionVerify      = "verify"
	FunctionHash        = "hash"
	FunctionEncapsulate = "encapsulate"
	FunctionDecapsulate = "decapsulate"
)

// CertificateProperties contains certificate-specific properties
type CertificateProperties struct {
	SubjectName           string `json:"subjectName,omitempty"`
	IssuerName            string `json:"issuerName,omitempty"`
	NotValidBefore        string `json:"notValidBefore,omitempty"`
	NotValidAfter         string `json:"notValidAfter,omitempty"`
	SignatureAlgorithmRef string `json:"signatureAlgorithmRef,omitempty"`
	SubjectPublicKeyRef   string `json:"subjectPublicKeyRef,omitempty"`
	CertificateFormat     string `json:"certificateFormat,omitempty"`
	CertificateExtension  string `json:"certificateExtension,omitempty"`
}

// ProtocolProperties contains protocol-specific properties
type ProtocolProperties struct {
	Type          string        `json:"type,omitempty"`          // tls, ssh, ipsec, ikev2, sstp, wpa
	Version       string        `json:"version,omitempty"`       // 1.2, 1.3
	CipherSuites  []CipherSuite `json:"cipherSuites,omitempty"`
	CryptoRefArray []string     `json:"cryptoRefArray,omitempty"` // refs to certificates/keys
}

// ProtocolType constants
const (
	ProtocolTLS   = "tls"
	ProtocolSSH   = "ssh"
	ProtocolIPSec = "ipsec"
	ProtocolIKEv2 = "ikev2"
	ProtocolSSTP  = "sstp"
	ProtocolWPA   = "wpa"
)

// CipherSuite represents a cipher suite
type CipherSuite struct {
	Name        string   `json:"name,omitempty"`
	Algorithms  []string `json:"algorithms,omitempty"` // bom-refs to algorithm components
	Identifiers []string `json:"identifiers,omitempty"`
}

// RelatedCryptoMaterialProperties contains key material properties
type RelatedCryptoMaterialProperties struct {
	Type           string     `json:"type,omitempty"`           // public-key, private-key, secret-key, key-pair, certificate, password, token
	ID             string     `json:"id,omitempty"`
	State          string     `json:"state,omitempty"`          // pre-activation, active, suspended, deactivated, compromised, destroyed
	Size           int        `json:"size,omitempty"`           // key size in bits
	AlgorithmRef   string     `json:"algorithmRef,omitempty"`   // bom-ref to algorithm
	CreationDate   string     `json:"creationDate,omitempty"`
	ActivationDate string     `json:"activationDate,omitempty"`
	ExpirationDate string     `json:"expirationDate,omitempty"`
	SecuredBy      *SecuredBy `json:"securedBy,omitempty"`
}

// KeyMaterialType constants
const (
	KeyTypePublicKey  = "public-key"
	KeyTypePrivateKey = "private-key"
	KeyTypeSecretKey  = "secret-key"
	KeyTypeKeyPair    = "key-pair"
	KeyTypeCertificate = "certificate"
	KeyTypePassword   = "password"
	KeyTypeToken      = "token"
)

// KeyState constants
const (
	KeyStatePreActivation = "pre-activation"
	KeyStateActive        = "active"
	KeyStateSuspended     = "suspended"
	KeyStateDeactivated   = "deactivated"
	KeyStateCompromised   = "compromised"
	KeyStateDestroyed     = "destroyed"
)

// SecuredBy describes how a key is secured
type SecuredBy struct {
	Mechanism    string `json:"mechanism,omitempty"` // Software, HSM, TPM, TEE
	AlgorithmRef string `json:"algorithmRef,omitempty"`
}

// NewAlgorithmComponent creates a new algorithm cryptographic asset component
func NewAlgorithmComponent(algorithm, mode string, keySize int) Component {
	name := algorithm
	if mode != "" {
		name = fmt.Sprintf("%s-%s", algorithm, strings.ToUpper(mode))
	}
	if keySize > 0 {
		name = fmt.Sprintf("%s-%d", name, keySize)
	}

	c := NewCryptoComponent(name)
	c.BOMRef = fmt.Sprintf("crypto/algorithm/%s", strings.ToLower(name))
	c.CryptoProperties = &CryptoProperties{
		AssetType: CryptoAssetAlgorithm,
		AlgorithmProperties: &AlgorithmProperties{
			Primitive:              inferPrimitive(algorithm),
			ParameterSetIdentifier: fmt.Sprintf("%d", keySize),
			Mode:                   mode,
			CryptoFunctions:        inferCryptoFunctions(algorithm),
			ClassicalSecurityLevel: inferSecurityLevel(algorithm, keySize),
		},
		OID: lookupAlgorithmOID(algorithm, mode, keySize),
	}

	return c
}

// NewCertificateComponent creates a new certificate cryptographic asset component
func NewCertificateComponent(subject, issuer, notBefore, notAfter, sigAlgo, keyType string, keySize int) Component {
	c := NewCryptoComponent(fmt.Sprintf("cert-%s", subject))
	c.BOMRef = fmt.Sprintf("crypto/certificate/%s", subject)
	c.CryptoProperties = &CryptoProperties{
		AssetType: CryptoAssetCertificate,
		CertificateProperties: &CertificateProperties{
			SubjectName:          subject,
			IssuerName:           issuer,
			NotValidBefore:       notBefore,
			NotValidAfter:        notAfter,
			CertificateFormat:    "X.509",
			CertificateExtension: "pem",
		},
	}

	return c
}

// NewProtocolComponent creates a new protocol cryptographic asset component
func NewProtocolComponent(protocolType, version string) Component {
	name := fmt.Sprintf("%s-%s", strings.ToUpper(protocolType), version)
	c := NewCryptoComponent(name)
	c.BOMRef = fmt.Sprintf("crypto/protocol/%s@%s", protocolType, version)
	c.CryptoProperties = &CryptoProperties{
		AssetType: CryptoAssetProtocol,
		ProtocolProperties: &ProtocolProperties{
			Type:    protocolType,
			Version: version,
		},
	}

	return c
}

// NewKeyComponent creates a new key material cryptographic asset component
func NewKeyComponent(keyType string, size int, algorithm string) Component {
	name := fmt.Sprintf("%s-%s-%d", keyType, algorithm, size)
	c := NewCryptoComponent(name)
	c.BOMRef = fmt.Sprintf("crypto/key/%s@%d", algorithm, size)
	c.CryptoProperties = &CryptoProperties{
		AssetType: CryptoAssetRelatedCryptoMaterial,
		RelatedCryptoMaterialProperties: &RelatedCryptoMaterialProperties{
			Type:  keyType,
			Size:  size,
			State: KeyStateActive,
		},
	}

	return c
}

// CipherFindingToComponent converts a cipher finding to a CycloneDX component
func CipherFindingToComponent(algorithm, mode, severity, file string, line int, description string) Component {
	keySize := inferKeySizeFromName(algorithm)
	c := NewAlgorithmComponent(algorithm, mode, keySize)

	c.Description = description
	c.AddProperty("zero:severity", severity)
	c.AddProperty("zero:file", file)
	c.AddProperty("zero:line", fmt.Sprintf("%d", line))

	// Add evidence of where it was found
	c.Evidence = &Evidence{
		Occurrences: []Occurrence{
			{Location: file, Line: line},
		},
	}

	return c
}

// TLSFindingToComponent converts a TLS finding to a CycloneDX component
func TLSFindingToComponent(tlsType, version, severity, file string, line int, description string) Component {
	c := NewProtocolComponent(ProtocolTLS, version)

	c.Description = description
	c.AddProperty("zero:severity", severity)
	c.AddProperty("zero:finding_type", tlsType)
	c.AddProperty("zero:file", file)
	c.AddProperty("zero:line", fmt.Sprintf("%d", line))

	c.Evidence = &Evidence{
		Occurrences: []Occurrence{
			{Location: file, Line: line},
		},
	}

	return c
}

// CertInfoToComponent converts certificate info to a CycloneDX component
func CertInfoToComponent(subject, issuer, notBefore, notAfter, keyType string, keySize int, sigAlgo, file string, isSelfSigned bool) Component {
	c := NewCertificateComponent(subject, issuer, notBefore, notAfter, sigAlgo, keyType, keySize)

	c.AddProperty("zero:key_type", keyType)
	c.AddProperty("zero:key_size", fmt.Sprintf("%d", keySize))
	c.AddProperty("zero:signature_algorithm", sigAlgo)
	c.AddProperty("zero:file", file)
	if isSelfSigned {
		c.AddProperty("zero:self_signed", "true")
	}

	return c
}

// inferPrimitive infers the cryptographic primitive from algorithm name
func inferPrimitive(algorithm string) string {
	alg := strings.ToLower(algorithm)

	// Authenticated encryption
	if strings.Contains(alg, "gcm") || strings.Contains(alg, "ccm") ||
	   strings.Contains(alg, "chacha20-poly1305") {
		return PrimitiveAE
	}

	// Hash functions
	if strings.HasPrefix(alg, "sha") || strings.HasPrefix(alg, "md5") ||
	   strings.HasPrefix(alg, "md4") || strings.Contains(alg, "blake") ||
	   strings.Contains(alg, "ripemd") {
		return PrimitiveHash
	}

	// MACs
	if strings.HasPrefix(alg, "hmac") || strings.Contains(alg, "poly1305") {
		return PrimitiveMAC
	}

	// Digital signatures
	if strings.Contains(alg, "rsa") || strings.Contains(alg, "ecdsa") ||
	   strings.Contains(alg, "ed25519") || strings.Contains(alg, "dsa") ||
	   strings.Contains(alg, "dilithium") || strings.Contains(alg, "ml-dsa") {
		return PrimitiveDSA
	}

	// Key encapsulation
	if strings.Contains(alg, "kyber") || strings.Contains(alg, "ml-kem") {
		return PrimitiveKEM
	}

	// Key derivation
	if strings.Contains(alg, "hkdf") || strings.Contains(alg, "pbkdf") ||
	   strings.Contains(alg, "scrypt") || strings.Contains(alg, "argon") {
		return PrimitiveKDF
	}

	// Block ciphers (default for AES, DES, etc. without mode)
	if strings.Contains(alg, "aes") || strings.Contains(alg, "des") ||
	   strings.Contains(alg, "blowfish") || strings.Contains(alg, "rc4") {
		return PrimitiveAE
	}

	return PrimitiveOther
}

// inferCryptoFunctions infers crypto functions from algorithm
func inferCryptoFunctions(algorithm string) []string {
	primitive := inferPrimitive(algorithm)

	switch primitive {
	case PrimitiveAE:
		return []string{FunctionKeyGen, FunctionEncrypt, FunctionDecrypt}
	case PrimitiveHash:
		return []string{FunctionHash}
	case PrimitiveMAC:
		return []string{FunctionKeyGen, FunctionSign, FunctionVerify}
	case PrimitiveDSA:
		return []string{FunctionKeyGen, FunctionSign, FunctionVerify}
	case PrimitiveKEM:
		return []string{FunctionKeyGen, FunctionEncapsulate, FunctionDecapsulate}
	case PrimitiveKDF:
		return []string{FunctionKeyGen}
	default:
		return []string{}
	}
}

// inferSecurityLevel infers classical security level from algorithm and key size
func inferSecurityLevel(algorithm string, keySize int) int {
	alg := strings.ToLower(algorithm)

	// Broken algorithms
	if strings.Contains(alg, "md5") || strings.Contains(alg, "md4") ||
	   strings.Contains(alg, "des") && !strings.Contains(alg, "3des") ||
	   strings.Contains(alg, "rc4") || strings.Contains(alg, "rc2") {
		return 0
	}

	// SHA family
	if strings.Contains(alg, "sha1") || strings.Contains(alg, "sha-1") {
		return 80 // Deprecated but not completely broken
	}
	if strings.Contains(alg, "sha256") || strings.Contains(alg, "sha-256") {
		return 128
	}
	if strings.Contains(alg, "sha384") || strings.Contains(alg, "sha-384") {
		return 192
	}
	if strings.Contains(alg, "sha512") || strings.Contains(alg, "sha-512") {
		return 256
	}

	// Symmetric ciphers
	if strings.Contains(alg, "aes") || strings.Contains(alg, "chacha20") {
		if keySize >= 256 {
			return 256
		}
		if keySize >= 192 {
			return 192
		}
		if keySize >= 128 {
			return 128
		}
	}

	// 3DES
	if strings.Contains(alg, "3des") || strings.Contains(alg, "triple") {
		return 112
	}

	// RSA (security is roughly key_size / 2 for symmetric equivalent)
	if strings.Contains(alg, "rsa") {
		if keySize >= 4096 {
			return 140
		}
		if keySize >= 3072 {
			return 128
		}
		if keySize >= 2048 {
			return 112
		}
		if keySize >= 1024 {
			return 80
		}
		return 0
	}

	// ECC
	if strings.Contains(alg, "ecdsa") || strings.Contains(alg, "ecdh") {
		if keySize >= 521 {
			return 256
		}
		if keySize >= 384 {
			return 192
		}
		if keySize >= 256 {
			return 128
		}
	}

	// Ed25519
	if strings.Contains(alg, "ed25519") {
		return 128
	}

	return keySize / 2 // Default estimate
}

// inferKeySizeFromName tries to extract key size from algorithm name
func inferKeySizeFromName(algorithm string) int {
	alg := strings.ToLower(algorithm)

	// Common patterns
	patterns := []struct {
		contains string
		size     int
	}{
		{"aes-256", 256},
		{"aes-192", 192},
		{"aes-128", 128},
		{"aes256", 256},
		{"aes192", 192},
		{"aes128", 128},
		{"rsa-4096", 4096},
		{"rsa-3072", 3072},
		{"rsa-2048", 2048},
		{"rsa-1024", 1024},
		{"rsa4096", 4096},
		{"rsa3072", 3072},
		{"rsa2048", 2048},
		{"rsa1024", 1024},
		{"sha-256", 256},
		{"sha-384", 384},
		{"sha-512", 512},
		{"sha256", 256},
		{"sha384", 384},
		{"sha512", 512},
		{"p-256", 256},
		{"p-384", 384},
		{"p-521", 521},
		{"secp256", 256},
		{"secp384", 384},
		{"secp521", 521},
	}

	for _, p := range patterns {
		if strings.Contains(alg, p.contains) {
			return p.size
		}
	}

	return 0
}

// lookupAlgorithmOID returns the OID for common algorithms
func lookupAlgorithmOID(algorithm, mode string, keySize int) string {
	key := strings.ToLower(fmt.Sprintf("%s-%s-%d", algorithm, mode, keySize))

	// Common OIDs
	oids := map[string]string{
		"aes-gcm-256": "2.16.840.1.101.3.4.1.46",
		"aes-gcm-192": "2.16.840.1.101.3.4.1.26",
		"aes-gcm-128": "2.16.840.1.101.3.4.1.6",
		"aes-cbc-256": "2.16.840.1.101.3.4.1.42",
		"aes-cbc-192": "2.16.840.1.101.3.4.1.22",
		"aes-cbc-128": "2.16.840.1.101.3.4.1.2",
		"sha-256-0":   "2.16.840.1.101.3.4.2.1",
		"sha-384-0":   "2.16.840.1.101.3.4.2.2",
		"sha-512-0":   "2.16.840.1.101.3.4.2.3",
		"rsa--2048":   "1.2.840.113549.1.1.1",
		"rsa--4096":   "1.2.840.113549.1.1.1",
	}

	if oid, ok := oids[key]; ok {
		return oid
	}
	return ""
}
