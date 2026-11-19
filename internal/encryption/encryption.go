// Package encryption encrypts request bodies from the agent to the server with
// the server's RSA key pair.
//
// Since RSA can encrypt only a few hundred bytes, the body is encrypted with a
// random AES-256-GCM key, and only that key is encrypted with RSA-OAEP
// (SHA-256). A new key and nonce from crypto/rand are used for every message.
// The encrypted message is:
//
//	RSA-OAEP(AES key) || GCM nonce || AES-GCM(body)
//
// The first part is as long as the RSA key, so the receiver can split the
// message with its private key alone. The encrypted AES key is also the GCM
// additional data, which binds it to the body.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

const (
	Header = "X-Encryption"
	Scheme = "rsa-oaep-sha256+aes-256-gcm"
	aesKeySize = 32
)

var (
	// ErrInvalidMessage is returned when a message cannot be decrypted: it was
	// not encrypted for this key, or it was truncated or modified.
	ErrInvalidMessage = errors.New("invalid encrypted message")

	errNotRSAKey = errors.New("not an RSA key")
)

// LoadPublicKey reads an RSA public key from a PEM file in PKIX
// ("PUBLIC KEY") or PKCS #1 ("RSA PUBLIC KEY") format.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	block, err := readPEM(path)
	if err != nil {
		return nil, err
	}

	switch block.Type {
	case "PUBLIC KEY":
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse public key %s: %w", path, err)
		}
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key %s: %w", path, errNotRSAKey)
		}
		return rsaKey, nil
	case "RSA PUBLIC KEY":
		key, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse public key %s: %w", path, err)
		}
		return key, nil
	default:
		return nil, fmt.Errorf("public key %s: unsupported PEM type %q", path, block.Type)
	}
}

// LoadPrivateKey reads an RSA private key from a PEM file in PKCS #8
// ("PRIVATE KEY") or PKCS #1 ("RSA PRIVATE KEY") format.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	block, err := readPEM(path)
	if err != nil {
		return nil, err
	}

	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key %s: %w", path, err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key %s: %w", path, errNotRSAKey)
		}
		return rsaKey, nil
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key %s: %w", path, err)
		}
		return key, nil
	default:
		return nil, fmt.Errorf("private key %s: unsupported PEM type %q", path, block.Type)
	}
}

func readPEM(path string) (*pem.Block, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("key file %s: no PEM data found", path)
	}

	return block, nil
}

// Encrypt encrypts plaintext for the owner of the private key matching pub.
func Encrypt(pub *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	key := make([]byte, aesKeySize)
	rand.Read(key) // Never returns an error.

	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, key, nil)
	if err != nil {
		return nil, fmt.Errorf("encrypt session key: %w", err)
	}

	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)

	msg := make([]byte, 0, len(encryptedKey)+len(nonce)+len(plaintext)+gcm.Overhead())
	msg = append(msg, encryptedKey...)
	msg = append(msg, nonce...)

	return gcm.Seal(msg, nonce, plaintext, encryptedKey), nil
}

// Decrypt decrypts a message made by Encrypt with the public part of priv.
// It returns an error wrapping ErrInvalidMessage if the message is not valid.
func Decrypt(priv *rsa.PrivateKey, msg []byte) ([]byte, error) {
	keyLen := priv.Size()
	if len(msg) < keyLen {
		return nil, fmt.Errorf("%w: too short", ErrInvalidMessage)
	}
	encryptedKey, rest := msg[:keyLen], msg[keyLen:]

	key, err := rsa.DecryptOAEP(sha256.New(), nil, priv, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: decrypt session key: %v", ErrInvalidMessage, err)
	}
	if len(key) != aesKeySize {
		return nil, fmt.Errorf("%w: session key has %d bytes", ErrInvalidMessage, len(key))
	}

	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	if len(rest) < gcm.NonceSize()+gcm.Overhead() {
		return nil, fmt.Errorf("%w: too short", ErrInvalidMessage)
	}
	nonce, ciphertext := rest[:gcm.NonceSize()], rest[gcm.NonceSize():]

	plaintext := make([]byte, 0, len(ciphertext)-gcm.Overhead())
	plaintext, err = gcm.Open(plaintext, nonce, ciphertext, encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidMessage, err)
	}

	return plaintext, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}
