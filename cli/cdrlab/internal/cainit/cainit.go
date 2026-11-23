// Package cainit makes the mock CDR CA and writes its intermediate Secret and its CDR CA bundle Secret, only when
// both are absent.
package cainit

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

type Config struct {
	IntermediateSecret string
	CABundleSecret     string
	KeyBits            int
	Now                func() time.Time
}

// Run gives created false when both Secrets exist.
func Run(ctx context.Context, secrets *kube.Client, cfg Config) (created bool, err error) {
	if cfg.IntermediateSecret == cfg.CABundleSecret {
		return false, errors.New("the intermediate Secret and the CA bundle Secret must have different names")
	}

	_, hasIntermediate, err := secrets.Get(ctx, cfg.IntermediateSecret)
	if err != nil {
		return false, err
	}

	_, hasBundle, err := secrets.Get(ctx, cfg.CABundleSecret)
	if err != nil {
		return false, err
	}

	switch {
	case hasIntermediate && hasBundle:
		return false, nil
	case hasIntermediate:
		return false, absent(cfg.CABundleSecret, cfg.IntermediateSecret)
	case hasBundle:
		return false, absent(cfg.IntermediateSecret, cfg.CABundleSecret)
	}

	intermediate, bundle, err := newCA(cfg)
	if err != nil {
		return false, err
	}

	if err := secrets.Create(ctx, intermediate); err != nil {
		return false, err
	}

	if err := secrets.Create(ctx, bundle); err != nil {
		// With only one of the two Secrets, each later run fails. Thus the run removes the Secret that it wrote.
		return false, errors.Join(err, secrets.Delete(ctx, intermediate.Name))
	}

	return true, nil
}

func absent(missing, present string) error {
	return fmt.Errorf("the Secret %q is absent, but the Secret %q exists: "+
		"give both Secrets of the same CA, or delete the Secret %q to make a new CA", missing, present, present)
}

// newCA keeps the root key only in memory, because no component needs it after the set-up.
func newCA(cfg Config) (intermediate, bundle kube.Secret, err error) {
	rootKey, err := rsa.GenerateKey(rand.Reader, cfg.KeyBits)
	if err != nil {
		return kube.Secret{}, kube.Secret{}, err
	}

	intermediateKey, err := rsa.GenerateKey(rand.Reader, cfg.KeyBits)
	if err != nil {
		return kube.Secret{}, kube.Secret{}, err
	}

	// X.509 times have no fractional seconds.
	root, authority, err := ca.NewCDR(rootKey, intermediateKey, cfg.Now().UTC().Truncate(time.Second))
	if err != nil {
		return kube.Secret{}, kube.Secret{}, err
	}

	keyPEM, err := ca.EncodeKey(authority.Key)
	if err != nil {
		return kube.Secret{}, kube.Secret{}, err
	}

	intermediate = kube.Secret{
		Name: cfg.IntermediateSecret,
		Type: kube.TLSType,
		Data: map[string][]byte{"tls.crt": ca.EncodeCertificates(authority.Certificate), "tls.key": keyPEM},
	}
	bundle = kube.Secret{
		Name: cfg.CABundleSecret,
		Type: "Opaque",
		Data: map[string][]byte{"ca.crt": ca.EncodeCertificates(authority.Certificate, root.Certificate)},
	}

	return intermediate, bundle, nil
}
