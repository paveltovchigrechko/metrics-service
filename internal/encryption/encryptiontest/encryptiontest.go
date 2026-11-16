// Package encryptiontest provides RSA keys and key files for tests.
package encryptiontest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// Generating an RSA key takes long, so all tests of a package share one.
var generate = sync.OnceValues(func() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
})

// Key returns a 2048-bit RSA key shared by all tests of the package.
func Key(t testing.TB) *rsa.PrivateKey {
	t.Helper()

	key, err := generate()
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	return key
}

// WriteKeyFiles writes the key from Key to a temporary directory as PEM files
// and returns the paths of the public key (PKIX) and the private key (PKCS #8).
func WriteKeyFiles(t testing.TB) (publicPath, privatePath string, key *rsa.PrivateKey) {
	t.Helper()

	key = Key(t)

	publicDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}

	dir := t.TempDir()
	publicPath = WritePEM(t, filepath.Join(dir, "public.pem"), "PUBLIC KEY", publicDER)
	privatePath = WritePEM(t, filepath.Join(dir, "private.pem"), "PRIVATE KEY", privateDER)

	return publicPath, privatePath, key
}

// WritePEM writes der to path as a PEM block of the given type and returns
// the path.
func WritePEM(t testing.TB, path, blockType string, der []byte) string {
	t.Helper()

	data := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	return path
}
