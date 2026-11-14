// Command pki creates the lab CAs and issues certificates from them.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pki/internal/labdir"
)

const usage = `usage: pki <command> [flags]

commands:
  init          create the lab CAs, if they do not exist
  issue         issue a CDR certificate from a CSR: pki issue -type server|client -crl-url URL -out FILE CSR
  issue-public  issue a TLS server certificate from the mock public CA:
                pki issue-public -cert FILE -key FILE NAME [NAME]
`

type config struct {
	now       func() time.Time
	caKeyBits int
}

// issueTime has no fractional seconds, because X.509 times do not keep them.
func (c config) issueTime() time.Time {
	return c.now().UTC().Truncate(time.Second)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, config{now: time.Now, caKeyBits: 4096}))
}

func run(args []string, stdout, stderr io.Writer, cfg config) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}

	commands := map[string]func([]string, io.Writer, config) error{
		"init":         initLab,
		"issue":        issue,
		"issue-public": issuePublic,
	}

	command, ok := commands[args[0]]
	if !ok {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}

	if err := command(args[1:], stdout, cfg); err != nil {
		_, _ = fmt.Fprintf(stderr, "pki %s: %v\n", args[0], err)
		return 1
	}

	return 0
}

func flags(name string) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	return fs, fs.String("dir", "", "directory of the lab CAs (required)")
}

func parse(fs *flag.FlagSet, args []string, dir *string) error {
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *dir == "" {
		return errors.New("-dir is required")
	}

	return nil
}

func initLab(args []string, stdout io.Writer, cfg config) error {
	fs, dir := flags("init")
	if err := parse(fs, args, dir); err != nil {
		return err
	}

	state, err := labdir.State(*dir)
	if err != nil {
		return err
	}

	if state == labdir.Complete {
		_, err = fmt.Fprintf(stdout, "the lab CAs exist in %s\n", *dir)
		return err
	}

	var keys [3]*rsa.PrivateKey
	for i := range keys {
		if keys[i], err = rsa.GenerateKey(rand.Reader, cfg.caKeyBits); err != nil {
			return err
		}
	}

	lab, err := ca.NewLab(
		ca.LabKeys{CDRRoot: keys[0], CDRIntermediate: keys[1], PublicRoot: keys[2]},
		cfg.issueTime(),
	)
	if err != nil {
		return err
	}

	if err := labdir.Save(*dir, lab); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "created the lab CAs in %s\ntrust bundles: %s, %s\n", *dir,
		filepath.Join(*dir, labdir.CDRBundle), filepath.Join(*dir, labdir.PublicRootCert))

	return err
}

func issue(args []string, stdout io.Writer, cfg config) error {
	fs, dir := flags("issue")
	kindName := fs.String("type", "", "server or client")
	out := fs.String("out", "", "file for the certificate")

	crlURL := fs.String("crl-url", "", "URL of the CRL of the mock CDR intermediate (required)")
	if err := parse(fs, args, dir); err != nil {
		return err
	}

	kinds := map[string]ca.Kind{"server": ca.Server, "client": ca.Client}

	kind, ok := kinds[*kindName]
	if !ok || *out == "" || fs.NArg() != 1 {
		return errors.New("usage: pki issue -type server|client -crl-url URL -out FILE CSR")
	}

	if err := checkCRLURL(*crlURL); err != nil {
		return err
	}

	lab, err := labdir.Load(*dir)
	if err != nil {
		return err
	}

	csr, err := labdir.ReadCSR(fs.Arg(0))
	if err != nil {
		return err
	}

	lab.CDRIntermediate.CRLURL = *crlURL

	cert, err := lab.CDRIntermediate.Issue(csr, kind, cfg.issueTime())
	if err != nil {
		return fmt.Errorf("the CSR does not follow the CDS profile:\n%w", err)
	}

	if err := labdir.WriteCertificates(*out, cert); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "issued %s, valid until %s\n", *out, cert.NotAfter.Format(time.DateOnly))

	return err
}

func issuePublic(args []string, stdout io.Writer, cfg config) error {
	fs, dir := flags("issue-public")
	certOut := fs.String("cert", "", "file for the certificate")

	keyOut := fs.String("key", "", "file for the private key")
	if err := parse(fs, args, dir); err != nil {
		return err
	}

	if *certOut == "" || *keyOut == "" || fs.NArg() == 0 {
		return errors.New("usage: pki issue-public -cert FILE -key FILE NAME [NAME]")
	}

	publicRoot, err := labdir.LoadPublicRoot(*dir)
	if err != nil {
		return err
	}

	key, err := rsa.GenerateKey(rand.Reader, ca.LeafKeyBits)
	if err != nil {
		return err
	}

	cert, err := publicRoot.IssuePublicServer(&key.PublicKey, fs.Args(), cfg.issueTime())
	if err != nil {
		return err
	}
	// Lab containers read the key as another user (pki/DESIGN.md: Keys).
	if err := labdir.WriteKey(*keyOut, key, 0o644); err != nil {
		return err
	}

	if err := labdir.WriteCertificates(*certOut, cert); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "issued %s, valid until %s\n", *certOut, cert.NotAfter.Format(time.DateOnly))

	return err
}

func checkCRLURL(value string) error {
	if value == "" {
		return errors.New("-crl-url is required")
	}

	return ca.CheckCRLURL(value)
}
