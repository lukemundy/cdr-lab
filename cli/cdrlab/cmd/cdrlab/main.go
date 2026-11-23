// Command cdrlab is the CLI of the lab. Its commands make the mock CDR CA and get CDR certificates from the Register
// for
// a cluster deployment (cli/cdrlab/DESIGN.md).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/cainit"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certissue"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register"
	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/platform/buildinfo"
)

const usage = `usage: cdrlab <command> [flags]

In a pod, the commands read and write the Secrets of the namespace of the pod, as its ServiceAccount.

commands:
  cdrlab version     give the commit of the build
  cdrlab ca init     make the mock CDR CA and write its Secrets, if both are absent:
                     cdrlab ca init -intermediate-secret NAME -ca-bundle-secret NAME
  cdrlab cert issue  get a CDR certificate from the Register and write it to a TLS Secret, if the Secret has no
                     current certificate. The operator token is in the variable REGISTER_OPERATOR_TOKEN.
                     cdrlab cert issue -register-url URL -kind server -legal-entity-id ID -brand-id ID
                       -cn NAME -o NAME [-dns NAME]... -secret NAME [-renew-before DURATION]
                     cdrlab cert issue -register-url URL -kind client -legal-entity-id ID -software-product-id ID
                       -cn NAME -o NAME -secret NAME [-renew-before DURATION]
`

// requestTimeout is more than the time of a Register issue, which signs one certificate.
const requestTimeout = 30 * time.Second

type deps struct {
	secrets    func() (*kube.Client, error)
	httpClient *http.Client
	getenv     func(string) string
	now        func() time.Time
	caKeyBits  int
}

// errUsage gives the exit status 2, as for a flag error.
var errUsage = errors.New("usage")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdout, os.Stderr, deps{
		secrets:    func() (*kube.Client, error) { return kube.InCluster(kube.ServiceAccountDir) },
		httpClient: &http.Client{Timeout: requestTimeout},
		getenv:     os.Getenv,
		now:        time.Now,
		// The size of the CA keys of the PKI specification (PKI-0003).
		caKeyBits: 4096,
	})

	stop()
	os.Exit(code)
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, d deps) int {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		_, _ = fmt.Fprint(stdout, usage)
		return 0
	}

	if len(args) == 1 && args[0] == "version" {
		info := buildinfo.Get()
		_, _ = fmt.Fprintf(stdout, "cdrlab commit %s, built %s, build %s\n", info.Commit, info.Time, info.ID)

		return 0
	}

	commands := map[string]func(context.Context, []string, io.Writer, deps) error{
		"ca init":    caInit,
		"cert issue": certIssue,
	}

	if len(args) < 2 || commands[args[0]+" "+args[1]] == nil {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}

	name := args[0] + " " + args[1]

	err := commands[name](ctx, args[2:], stdout, d)

	switch {
	case err == nil:
		return 0
	case errors.Is(err, errUsage):
		_, _ = fmt.Fprintf(stderr, "cdrlab %s: %v\n", name, err)
		return 2
	default:
		_, _ = fmt.Fprintf(stderr, "cdrlab %s: %v\n", name, err)
		return 1
	}
}

func newFlags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	return fs
}

func parse(fs *flag.FlagSet, args []string, required map[string]*string) error {
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}

	if fs.NArg() != 0 {
		return fmt.Errorf("%w: unexpected arguments: %s", errUsage, strings.Join(fs.Args(), " "))
	}

	var missing []string

	fs.VisitAll(func(f *flag.Flag) {
		if p, ok := required[f.Name]; ok && *p == "" {
			missing = append(missing, "-"+f.Name)
		}
	})

	if len(missing) > 0 {
		return fmt.Errorf("%w: %s is required", errUsage, strings.Join(missing, ", "))
	}

	return nil
}

func caInit(ctx context.Context, args []string, stdout io.Writer, d deps) error {
	fs := newFlags("ca init")
	intermediate := fs.String(
		"intermediate-secret",
		"",
		"the name of the Secret of the intermediate certificate and key",
	)
	bundle := fs.String("ca-bundle-secret", "", "the name of the Secret of the CDR CA bundle")

	if err := parse(
		fs,
		args,
		map[string]*string{"intermediate-secret": intermediate, "ca-bundle-secret": bundle},
	); err != nil {
		return err
	}

	secrets, err := d.secrets()
	if err != nil {
		return err
	}

	created, err := cainit.Run(ctx, secrets, cainit.Config{
		IntermediateSecret: *intermediate, CABundleSecret: *bundle, KeyBits: d.caKeyBits, Now: d.now,
	})
	if err != nil {
		return err
	}

	if !created {
		_, err = fmt.Fprintf(stdout, "the Secrets %q and %q exist: no change\n", *intermediate, *bundle)
		return err
	}

	_, err = fmt.Fprintf(stdout, "created the mock CDR CA in the Secrets %q and %q\n", *intermediate, *bundle)

	return err
}

type dnsNames []string

func (n *dnsNames) String() string { return strings.Join(*n, ",") }

func (n *dnsNames) Set(v string) error {
	*n = append(*n, v)
	return nil
}

func certIssue(ctx context.Context, args []string, stdout io.Writer, d deps) error {
	cfg, registerURL, err := parseCertIssue(args, d.now)
	if err != nil {
		return err
	}

	token := d.getenv("REGISTER_OPERATOR_TOKEN")
	if token == "" {
		return fmt.Errorf("%w: the variable REGISTER_OPERATOR_TOKEN is not set", errUsage)
	}

	secrets, err := d.secrets()
	if err != nil {
		return err
	}

	cert, err := certissue.Run(ctx, secrets, register.New(registerURL, token, d.httpClient), cfg)
	if err != nil {
		return err
	}

	if cert == nil {
		_, err = fmt.Fprintf(stdout, "the Secret %q has a current certificate: no change\n", cfg.Secret)
		return err
	}

	_, err = fmt.Fprintf(stdout, "wrote a new certificate to the Secret %q, valid until %s\n",
		cfg.Secret, cert.NotAfter.Format(time.DateOnly))

	return err
}

func parseCertIssue(args []string, now func() time.Time) (certissue.Config, string, error) {
	fs := newFlags("cert issue")
	registerURL := fs.String("register-url", "", "the URL of the Register operator interface")
	kindName := fs.String("kind", "", "server or client")
	legalEntity := fs.String("legal-entity-id", "", "the ID of the legal entity")
	brand := fs.String("brand-id", "", "the ID of the data holder brand, for a server certificate")
	product := fs.String("software-product-id", "", "the ID of the software product, for a client certificate")
	cn := fs.String("cn", "", "the CN of the subject")
	org := fs.String("o", "", "the O of the subject")
	secret := fs.String("secret", "", "the name of the TLS Secret")
	renewBefore := fs.Duration("renew-before", certissue.DefaultRenewBefore,
		"get a new certificate when the current one expires within this time")

	var dns dnsNames
	fs.Var(&dns, "dns", "a DNS name of a server certificate (repeat for each name)")

	err := parse(fs, args, map[string]*string{
		"register-url":    registerURL,
		"kind":            kindName,
		"legal-entity-id": legalEntity,
		"cn":              cn,
		"o":               org,
		"secret":          secret,
	})
	if err != nil {
		return certissue.Config{}, "", err
	}

	cfg := certissue.Config{
		LegalEntityID: *legalEntity, CommonName: *cn, Organisation: *org, DNSNames: dns, Secret: *secret,
		RenewBefore: *renewBefore, Now: now,
	}

	switch {
	case *kindName == "server" && *brand != "" && *product == "":
		cfg.Kind, cfg.RecordID = ca.Server, *brand
	case *kindName == "client" && *product != "" && *brand == "" && len(dns) == 0:
		cfg.Kind, cfg.RecordID = ca.Client, *product
	default:
		return certissue.Config{}, "", fmt.Errorf(
			"%w: -kind server needs -brand-id, and -kind client needs -software-product-id and no -dns", errUsage)
	}

	if *renewBefore <= 0 {
		return certissue.Config{}, "", fmt.Errorf("%w: -renew-before must be more than 0", errUsage)
	}

	return cfg, *registerURL, nil
}
