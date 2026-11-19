package encryption_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/encryption"
	"github.com/paveltovchigrechko/metrics-service/internal/encryption/encryptiontest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := encryptiontest.Key(t)

	tests := map[string][]byte{
		"empty":                       {},
		"small":                       []byte(`[{"id":"PollCount","type":"counter","delta":5}]`),
		"larger than RSA can encrypt": bytes.Repeat([]byte("metrics"), 150_000),
	}

	for name, plaintext := range tests {
		t.Run(name, func(t *testing.T) {
			msg, err := encryption.Encrypt(&key.PublicKey, plaintext)
			require.NoError(t, err)
			if len(plaintext) > 0 {
				assert.NotContains(t, string(msg), string(plaintext[:min(len(plaintext), 32)]))
			}

			got, err := encryption.Decrypt(key, msg)
			require.NoError(t, err)
			assert.Equal(t, plaintext, got)
		})
	}
}

func TestEncrypt_IsRandomized(t *testing.T) {
	key := encryptiontest.Key(t)
	plaintext := []byte("same message")

	first, err := encryption.Encrypt(&key.PublicKey, plaintext)
	require.NoError(t, err)
	second, err := encryption.Encrypt(&key.PublicKey, plaintext)
	require.NoError(t, err)

	assert.NotEqual(t, first, second, "a new session key and nonce must be used every time")
}

func TestDecrypt_RejectsInvalidMessages(t *testing.T) {
	key := encryptiontest.Key(t)
	plaintext := []byte(`[{"id":"Alloc","type":"gauge","value":1.5}]`)

	msg, err := encryption.Encrypt(&key.PublicKey, plaintext)
	require.NoError(t, err)

	flipByte := func(i int) []byte {
		modified := bytes.Clone(msg)
		modified[i] ^= 0xFF
		return modified
	}

	tests := map[string][]byte{
		"empty":                 nil,
		"shorter than RSA key":  msg[:key.Size()-1],
		"no nonce and tag":      msg[:key.Size()+4],
		"truncated ciphertext":  msg[:len(msg)-1],
		"modified session key":  flipByte(0),
		"modified nonce":        flipByte(key.Size()),
		"modified ciphertext":   flipByte(len(msg) - 20),
		"modified tag":          flipByte(len(msg) - 1),
		"plaintext, not cipher": bytes.Repeat([]byte("x"), 400),
	}

	for name, invalid := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := encryption.Decrypt(key, invalid)
			assert.ErrorIs(t, err, encryption.ErrInvalidMessage)
		})
	}

	t.Run("encrypted for another key", func(t *testing.T) {
		other, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		_, err = encryption.Decrypt(other, msg)
		assert.ErrorIs(t, err, encryption.ErrInvalidMessage)
	})
}

func TestLoadKeys(t *testing.T) {
	publicPath, privatePath, key := encryptiontest.WriteKeyFiles(t)

	t.Run("PKIX public key", func(t *testing.T) {
		got, err := encryption.LoadPublicKey(publicPath)
		require.NoError(t, err)
		assert.True(t, key.PublicKey.Equal(got))
	})

	t.Run("PKCS #8 private key", func(t *testing.T) {
		got, err := encryption.LoadPrivateKey(privatePath)
		require.NoError(t, err)
		assert.True(t, key.Equal(got))
	})

	t.Run("PKCS #1 public key", func(t *testing.T) {
		path := encryptiontest.WritePEM(t, filepath.Join(t.TempDir(), "pub.pem"),
			"RSA PUBLIC KEY", x509.MarshalPKCS1PublicKey(&key.PublicKey))

		got, err := encryption.LoadPublicKey(path)
		require.NoError(t, err)
		assert.True(t, key.PublicKey.Equal(got))
	})

	t.Run("PKCS #1 private key", func(t *testing.T) {
		path := encryptiontest.WritePEM(t, filepath.Join(t.TempDir(), "priv.pem"),
			"RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))

		got, err := encryption.LoadPrivateKey(path)
		require.NoError(t, err)
		assert.True(t, key.Equal(got))
	})

	t.Run("keys from files work together", func(t *testing.T) {
		pub, err := encryption.LoadPublicKey(publicPath)
		require.NoError(t, err)
		priv, err := encryption.LoadPrivateKey(privatePath)
		require.NoError(t, err)

		msg, err := encryption.Encrypt(pub, []byte("hello"))
		require.NoError(t, err)
		got, err := encryption.Decrypt(priv, msg)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(got))
	})
}

func TestLoadKeys_Errors(t *testing.T) {
	publicPath, privatePath, _ := encryptiontest.WriteKeyFiles(t)
	dir := t.TempDir()

	notPEM := filepath.Join(dir, "not.pem")
	require.NoError(t, os.WriteFile(notPEM, []byte("not a key"), 0600))

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ecPublicDER, err := x509.MarshalPKIXPublicKey(&ecKey.PublicKey)
	require.NoError(t, err)
	ecPrivateDER, err := x509.MarshalPKCS8PrivateKey(ecKey)
	require.NoError(t, err)
	ecPublic := encryptiontest.WritePEM(t, filepath.Join(dir, "ec_pub.pem"), "PUBLIC KEY", ecPublicDER)
	ecPrivate := encryptiontest.WritePEM(t, filepath.Join(dir, "ec_priv.pem"), "PRIVATE KEY", ecPrivateDER)
	corrupted := encryptiontest.WritePEM(t, filepath.Join(dir, "bad.pem"), "PUBLIC KEY", []byte("garbage"))

	publicTests := map[string]string{
		"missing file":        filepath.Join(dir, "missing.pem"),
		"not PEM":             notPEM,
		"private key file":    privatePath,
		"ECDSA key":           ecPublic,
		"corrupted key bytes": corrupted,
	}
	for name, path := range publicTests {
		t.Run("public/"+name, func(t *testing.T) {
			_, err := encryption.LoadPublicKey(path)
			assert.Error(t, err)
		})
	}

	privateTests := map[string]string{
		"missing file":    filepath.Join(dir, "missing.pem"),
		"not PEM":         notPEM,
		"public key file": publicPath,
		"ECDSA key":       ecPrivate,
	}
	for name, path := range privateTests {
		t.Run("private/"+name, func(t *testing.T) {
			_, err := encryption.LoadPrivateKey(path)
			assert.Error(t, err)
		})
	}
}

func BenchmarkEncrypt(b *testing.B) {
	key := encryptiontest.Key(b)
	plaintext := bytes.Repeat([]byte("m"), 4<<10)

	for b.Loop() {
		if _, err := encryption.Encrypt(&key.PublicKey, plaintext); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := encryptiontest.Key(b)
	msg, err := encryption.Encrypt(&key.PublicKey, bytes.Repeat([]byte("m"), 4<<10))
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if _, err := encryption.Decrypt(key, msg); err != nil {
			b.Fatal(err)
		}
	}
}
