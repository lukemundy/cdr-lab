package cds

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

const certificateCheckInterval = 10 * time.Second

// ServerCertificate gives each new handshake the certificate that is current at that time.
type ServerCertificate struct {
	name    string
	current atomic.Pointer[tls.Certificate]
}

// NewServerCertificate gives a certificate that only the service replaces. The name identifies it in the log.
func NewServerCertificate(ctx context.Context, name string, cert tls.Certificate) (*ServerCertificate, error) {
	c := &ServerCertificate{name: name}
	if err := c.Replace(ctx, cert); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *ServerCertificate) Replace(ctx context.Context, cert tls.Certificate) error {
	if cert.Leaf == nil {
		if len(cert.Certificate) == 0 {
			return fmt.Errorf("replace the %s TLS certificate: the replacement has no certificate", c.name)
		}

		leaf, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return fmt.Errorf("replace the %s TLS certificate: %w", c.name, err)
		}

		cert.Leaf = leaf
	}

	c.current.Store(&cert)
	slog.InfoContext(ctx, "the TLS listeners use a new certificate",
		slog.String("app.tls.certificate.name", c.name),
		slog.String("app.tls.server.serial_number", strings.ToUpper(cert.Leaf.SerialNumber.Text(16))),
		slog.String("tls.server.not_after", cert.Leaf.NotAfter.UTC().Format(time.RFC3339)),
	)

	return nil
}

func (c *ServerCertificate) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return c.current.Load(), nil
}

// WatchCertificateFiles reads the files again at each check until ctx ends, because a file system watch
// does not always see the symlink swap of a mounted Kubernetes Secret.
func WatchCertificateFiles(ctx context.Context, name, certFile, keyFile string) (*ServerCertificate, error) {
	return watchCertificateFiles(ctx, name, certFile, keyFile, certificateCheckInterval)
}

func watchCertificateFiles(
	ctx context.Context, name, certFile, keyFile string, interval time.Duration,
) (*ServerCertificate, error) {
	files, err := readCertificateFiles(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load the %s TLS certificate: %w", name, err)
	}

	cert, err := files.keyPair()
	if err != nil {
		return nil, fmt.Errorf("load the %s TLS certificate: %w", name, err)
	}

	c, err := NewServerCertificate(ctx, name, cert)
	if err != nil {
		return nil, err
	}

	w := &certificateWatch{cert: c, certFile: certFile, keyFile: keyFile, last: files}
	go w.run(ctx, interval)

	return c, nil
}

type certificateWatch struct {
	cert              *ServerCertificate
	certFile, keyFile string
	last              certificateFiles
}

func (w *certificateWatch) run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.check(ctx)
		}
	}
}

func (w *certificateWatch) check(ctx context.Context) {
	// A file that cannot be read gives no content, so a file that stays missing gives only one warning.
	files, err := readCertificateFiles(w.certFile, w.keyFile)
	if files == w.last {
		return
	}

	w.last = files
	if err != nil {
		w.warn(ctx, err)
		return
	}

	cert, err := files.keyPair()
	if err != nil {
		w.warn(ctx, err)
		return
	}

	if err := w.cert.Replace(ctx, cert); err != nil {
		w.warn(ctx, err)
	}
}

func (w *certificateWatch) warn(ctx context.Context, err error) {
	slog.WarnContext(ctx, "the TLS certificate files are not valid; the last valid certificate stays in use",
		slog.String("app.tls.certificate.name", w.cert.name), slog.Any("error", err))
}

// certificateFiles holds the content as strings, so that a check can compare two reads.
type certificateFiles struct{ cert, key string }

func readCertificateFiles(certFile, keyFile string) (certificateFiles, error) {
	cert, certErr := os.ReadFile(certFile) //nolint:gosec // G304: the operator of the service gives the file name
	key, keyErr := os.ReadFile(keyFile)    //nolint:gosec // G304: the operator of the service gives the file name

	if err := errors.Join(certErr, keyErr); err != nil {
		return certificateFiles{}, err
	}

	return certificateFiles{cert: string(cert), key: string(key)}, nil
}

func (f certificateFiles) keyPair() (tls.Certificate, error) {
	return tls.X509KeyPair([]byte(f.cert), []byte(f.key))
}
