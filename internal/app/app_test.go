package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPrivateKey_PKCS8(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "private.pem")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}), 0o600))

	got, err := loadPrivateKey(path)

	require.NoError(t, err)
	assert.Equal(t, key.N, got.N)
}

func TestLoadPublicKey_PKIX(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pkix, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "public.pem")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pkix}), 0o644))

	got, err := loadPublicKey(path)

	require.NoError(t, err)
	assert.Equal(t, key.PublicKey.N, got.N)
}

func TestLoadPublicKey_PKCS1(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pkcs1 := x509.MarshalPKCS1PublicKey(&key.PublicKey)

	path := filepath.Join(t.TempDir(), "public.pem")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pkcs1}), 0o644))

	got, err := loadPublicKey(path)

	require.NoError(t, err)
	assert.Equal(t, key.PublicKey.N, got.N)
}

func TestLoadPrivateKey_InvalidPEM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private.pem")
	require.NoError(t, os.WriteFile(path, []byte("not-a-pem"), 0o600))

	_, err := loadPrivateKey(path)

	assert.Error(t, err)
}

func TestLoadPublicKey_InvalidPEM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "public.pem")
	require.NoError(t, os.WriteFile(path, []byte("not-a-pem"), 0o644))

	_, err := loadPublicKey(path)

	assert.Error(t, err)
}
