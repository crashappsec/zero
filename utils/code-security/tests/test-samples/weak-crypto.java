// Test sample: Cryptographic weakness vulnerabilities
// This file contains intentionally vulnerable code for testing

import java.security.MessageDigest;
import java.security.SecureRandom;
import java.util.Random;
import javax.crypto.Cipher;
import javax.crypto.spec.SecretKeySpec;

public class WeakCrypto {

    // VULNERABLE: MD5 for password hashing
    public String hashPasswordMD5(String password) throws Exception {
        // Vulnerable - MD5 is cryptographically broken
        MessageDigest md = MessageDigest.getInstance("MD5");
        byte[] hash = md.digest(password.getBytes());
        return bytesToHex(hash);
    }

    // VULNERABLE: SHA-1 for security purposes
    public String hashDataSHA1(String data) throws Exception {
        // Vulnerable - SHA-1 has known collision attacks
        MessageDigest md = MessageDigest.getInstance("SHA-1");
        byte[] hash = md.digest(data.getBytes());
        return bytesToHex(hash);
    }

    // VULNERABLE: DES encryption
    public byte[] encryptDES(String data, String key) throws Exception {
        // Vulnerable - DES has only 56-bit key, easily broken
        Cipher cipher = Cipher.getInstance("DES/ECB/PKCS5Padding");
        SecretKeySpec secretKey = new SecretKeySpec(key.getBytes(), "DES");
        cipher.init(Cipher.ENCRYPT_MODE, secretKey);
        return cipher.doFinal(data.getBytes());
    }

    // VULNERABLE: ECB mode encryption
    public byte[] encryptECB(String data, byte[] key) throws Exception {
        // Vulnerable - ECB mode leaks patterns
        Cipher cipher = Cipher.getInstance("AES/ECB/PKCS5Padding");
        SecretKeySpec secretKey = new SecretKeySpec(key, "AES");
        cipher.init(Cipher.ENCRYPT_MODE, secretKey);
        return cipher.doFinal(data.getBytes());
    }

    // VULNERABLE: Weak random number generator
    public String generateToken() {
        // Vulnerable - java.util.Random is not cryptographically secure
        Random random = new Random();
        StringBuilder token = new StringBuilder();
        for (int i = 0; i < 32; i++) {
            token.append(Integer.toHexString(random.nextInt(16)));
        }
        return token.toString();
    }

    // VULNERABLE: Hardcoded encryption key
    private static final String SECRET_KEY = "MySecretKey12345";

    public byte[] encryptWithHardcodedKey(String data) throws Exception {
        // Vulnerable - key is in source code
        Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
        SecretKeySpec secretKey = new SecretKeySpec(SECRET_KEY.getBytes(), "AES");
        cipher.init(Cipher.ENCRYPT_MODE, secretKey);
        return cipher.doFinal(data.getBytes());
    }

    // VULNERABLE: Short RSA key
    public void generateWeakRSAKey() throws Exception {
        java.security.KeyPairGenerator keyGen = java.security.KeyPairGenerator.getInstance("RSA");
        // Vulnerable - 1024-bit RSA is no longer considered secure
        keyGen.initialize(1024);
        keyGen.generateKeyPair();
    }

    // SECURE versions for comparison

    // SECURE: bcrypt or Argon2 should be used for passwords
    // (Using SHA-256 here as placeholder - real code should use bcrypt)
    public String hashPasswordSecure(String password) throws Exception {
        MessageDigest md = MessageDigest.getInstance("SHA-256");
        byte[] hash = md.digest(password.getBytes());
        return bytesToHex(hash);
    }

    // SECURE: AES-GCM encryption
    public byte[] encryptSecure(String data, byte[] key, byte[] iv) throws Exception {
        Cipher cipher = Cipher.getInstance("AES/GCM/NoPadding");
        SecretKeySpec secretKey = new SecretKeySpec(key, "AES");
        javax.crypto.spec.GCMParameterSpec gcmSpec = new javax.crypto.spec.GCMParameterSpec(128, iv);
        cipher.init(Cipher.ENCRYPT_MODE, secretKey, gcmSpec);
        return cipher.doFinal(data.getBytes());
    }

    // SECURE: SecureRandom for cryptographic purposes
    public String generateTokenSecure() {
        SecureRandom random = new SecureRandom();
        byte[] token = new byte[32];
        random.nextBytes(token);
        return bytesToHex(token);
    }

    // SECURE: Strong RSA key length
    public void generateStrongRSAKey() throws Exception {
        java.security.KeyPairGenerator keyGen = java.security.KeyPairGenerator.getInstance("RSA");
        keyGen.initialize(4096);  // 4096-bit is recommended
        keyGen.generateKeyPair();
    }

    private String bytesToHex(byte[] bytes) {
        StringBuilder sb = new StringBuilder();
        for (byte b : bytes) {
            sb.append(String.format("%02x", b));
        }
        return sb.toString();
    }
}
