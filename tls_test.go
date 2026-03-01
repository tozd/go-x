package x_test

import (
	"crypto/tls"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/tozd/go/x"
)

func TestCreateTempCertificateFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	errE := x.CreateTempCertificateFiles(certPath, keyPath, []string{"localhost"})
	require.NoError(t, errE, "% -+#.1v", errE)

	// Verify the files can be loaded as a valid TLS certificate pair.
	_, err := tls.LoadX509KeyPair(certPath, keyPath)
	assert.NoError(t, err)
}

func TestCreateTempCertificateFilesErrors(t *testing.T) {
	t.Parallel()

	t.Run("bad cert path", func(t *testing.T) {
		t.Parallel()

		errE := x.CreateTempCertificateFiles("/nonexistent/path/cert.pem", "/nonexistent/path/key.pem", []string{"localhost"})
		require.Error(t, errE)
	})

	t.Run("bad key path", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		certPath := filepath.Join(dir, "cert.pem")

		errE := x.CreateTempCertificateFiles(certPath, "/nonexistent/path/key.pem", []string{"localhost"})
		require.Error(t, errE)
	})
}
