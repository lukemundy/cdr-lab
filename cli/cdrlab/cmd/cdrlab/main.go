// Command cdrlab is the CLI of the lab. Its commands make the mock CDR CA and the mock public CA, get CDR
// certificates from the Register and issue public TLS certificates, with Kubernetes Secrets or files as the output.
// They also add participants to the Register and customers to a data holder brand (cli/cdrlab/design/README.md).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	cdrlab "github.com/lukemundy/cdr-lab"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/cainit"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certcheck"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certissue"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/compose"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/dataholder"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/pemfile"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/publiccert"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register"
	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/platform/buildinfo"
)

const usage = `usage: cdrlab <command> [flags]

With Secrets as the output, the commands read and write the Secrets of the namespace of the pod, as its
ServiceAccount. With files as the output, they make no call to Kubernetes.

commands:
  cdrlab version            give the version, the commit and the build time
  cdrlab ca init            make the mock CDR CA, if its output is absent:
                            cdrlab ca init -intermediate-secret NAME -ca-bundle-secret NAME
                            cdrlab ca init -dir DIR    (also makes the mock public CA)
  cdrlab cert issue         get a CDR certificate from the Register and write it to the output, if the output has no
                            current certificate. The operator token is in the variable REGISTER_OPERATOR_TOKEN.
                            cdrlab cert issue -register-url URL -kind server -legal-entity-id ID -brand-id ID
                              -cn NAME -o NAME [-dns NAME]... OUTPUT [-renew-before DURATION]
                            cdrlab cert issue -register-url URL -kind client -legal-entity-id ID -software-product-id ID
                              -cn NAME -o NAME OUTPUT [-renew-before DURATION]
                            OUTPUT is -secret NAME, or -cert FILE -key FILE [-key-readable-by-all].
                            With -csr FILE in place of -cn, -o, -dns and -key, the command sends that CSR and writes
                            only the -cert FILE.
  cdrlab public-cert issue  issue a TLS server certificate from the mock public CA of a directory, if the
                            certificate file has no current certificate. The first NAME is the CN.
                            cdrlab public-cert issue -dir DIR -cert FILE -key FILE [-key-readable-by-all]
                              [-renew-before DURATION] NAME...
  cdrlab participants add   add the participants of a load file to the Register, and change no participant that
                            exists. The operator token is in the variable REGISTER_OPERATOR_TOKEN.
                            cdrlab participants add -register-url URL -file FILE
  cdrlab customers add      add the customers of a customer load file to a data holder brand, and change no customer
                            that exists. The operator token is in the variable DATAHOLDER_OPERATOR_TOKEN.
                            cdrlab customers add -dataholder-url URL -file FILE
  cdrlab up                 make the CAs and the certificates of the compose lab, and start the compose lab. The
                            certificate list lab/certificates.json of the lab directory gives the certificates.
                            In a checkout, -source uses the compose project of the checkout and builds the images:
                            cdrlab up -source [-dir DIR] [-all-banks] [-no-build] [-no-telemetry]
                              [-wait-timeout DURATION] [SERVICE]...
                            With no -source, cdrlab writes its own compose project into the lab directory, and
                            Compose pulls the release images of the version of cdrlab:
                            cdrlab up [-dir DIR] [-image-prefix PREFIX] [-image-tag TAG] [-all-banks]
                              [-no-telemetry] [-wait-timeout DURATION] [SERVICE]...
  cdrlab down               stop the compose lab, and keep its data: cdrlab down [-dir DIR]
  cdrlab reset              stop the compose lab, and delete its data and its CAs, and the compose project that
                            cdrlab up wrote: cdrlab reset [-dir DIR]

The compose commands run Docker Compose as "docker compose", or as the command in the variable CDRLAB_COMPOSE.
`

// requestTimeout is more than the time of a Register issue, which signs one certificate.
const requestTimeout = 30 * time.Second

type deps struct {
	secrets    func() (*kube.Client, error)
	httpClient *http.Client
	getenv     func(string) string
	now        func() time.Time
	caKeyBits  int
	environ    func() []string
	compose    compose.Exec
	stderr     io.Writer
	// project is the compose project of the release mode, and version is the version of the binary.
	project fs.FS
	version string
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
		environ:    os.Environ,
		compose:    compose.OSExec,
		project:    cdrlab.ComposeProject,
		version:    buildinfo.Get().Version,
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
		_, _ = fmt.Fprintf(stdout, "cdrlab %s, commit %s, built %s\n", info.Version, info.Commit, info.Time)

		return 0
	}

	commands := map[string]command{
		"ca init":           caInit,
		"cert issue":        certIssue,
		"public-cert issue": publicCertIssue,
		"participants add":  participantsAdd,
		"customers add":     customersAdd,
		"up":                up,
		"down":              down,
		"reset":             reset,
	}

	name, rest := commandName(args, commands)
	if name == "" {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}

	d.stderr = stderr
	err := commands[name](ctx, rest, stdout, d)

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

type command func(context.Context, []string, io.Writer, deps) error

// commandName gives the command of one or two words at the start of args, and the other arguments.
func commandName(args []string, commands map[string]command) (string, []string) {
	if len(args) >= 1 {
		if _, ok := commands[args[0]]; ok {
			return args[0], args[1:]
		}
	}

	if len(args) >= 2 {
		if _, ok := commands[args[0]+" "+args[1]]; ok {
			return args[0] + " " + args[1], args[2:]
		}
	}

	return "", nil
}

func newFlags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	return fs
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}

	return nil
}

// parse also refuses arguments that are not flags.
func parse(fs *flag.FlagSet, args []string, required map[string]*string) error {
	if err := parseFlags(fs, args); err != nil {
		return err
	}

	if fs.NArg() != 0 {
		return fmt.Errorf("%w: unexpected arguments: %s", errUsage, strings.Join(fs.Args(), " "))
	}

	return requireFlags(fs, required)
}

func requireFlags(fs *flag.FlagSet, required map[string]*string) error {
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

func refuseFlags(fs *flag.FlagSet, reason string, names ...string) error {
	var set []string

	fs.Visit(func(f *flag.Flag) {
		for _, name := range names {
			if f.Name == name {
				set = append(set, "-"+f.Name)
			}
		}
	})

	if len(set) > 0 {
		return fmt.Errorf("%w: %s %s", errUsage, reason, strings.Join(set, ", "))
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
	dir := fs.String("dir", "", "the directory of the files of the mock CDR CA and the mock public CA")

	if err := parse(fs, args, nil); err != nil {
		return err
	}

	if *dir != "" {
		if err := refuseFlags(fs, "-dir cannot go with", "intermediate-secret", "ca-bundle-secret"); err != nil {
			return err
		}

		return caInitDir(*dir, stdout, d)
	}

	if err := requireFlags(
		fs,
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

func caInitDir(dir string, stdout io.Writer, d deps) error {
	created, err := cainit.RunDir(dir, d.caKeyBits, d.now)
	if err != nil {
		return err
	}

	if !created {
		_, err = fmt.Fprintf(stdout, "the lab CAs exist in %s: no change\n", dir)
		return err
	}

	_, err = fmt.Fprintf(stdout, "created the lab CAs in %s\ntrust bundles: %s, %s\n", dir,
		filepath.Join(dir, "cdr/ca-bundle.pem"), filepath.Join(dir, "public/root.crt"))

	return err
}

type dnsNames []string

func (n *dnsNames) String() string { return strings.Join(*n, ",") }

func (n *dnsNames) Set(v string) error {
	*n = append(*n, v)
	return nil
}

// certOutput is a Secret, or a certificate file and a key file.
type certOutput struct {
	secret, cert, key, csr string
	anyUser                bool
}

func certIssue(ctx context.Context, args []string, stdout io.Writer, d deps) error {
	cfg, output, registerURL, err := parseCertIssue(args, d.now)
	if err != nil {
		return err
	}

	token, err := operatorToken(d, "REGISTER_OPERATOR_TOKEN")
	if err != nil {
		return err
	}

	var (
		out  certissue.Output
		name string
	)

	switch {
	case output.secret != "":
		secrets, err := d.secrets()
		if err != nil {
			return err
		}

		out, name = certissue.NewSecret(secrets, output.secret), fmt.Sprintf("the Secret %q", output.secret)
	default:
		if output.csr != "" {
			if cfg.CSR, err = readCSR(output.csr); err != nil {
				return err
			}
		}

		out, name = certissue.Files{Cert: output.cert, Key: output.key, KeyPerm: keyPerm(output.anyUser)}, output.cert
	}

	cert, err := certissue.Run(ctx, out, register.New(registerURL, token, d.httpClient), cfg)
	if err != nil {
		return err
	}

	if cert == nil {
		_, err = fmt.Fprintf(stdout, "%s has a current certificate: no change\n", name)
		return err
	}

	_, err = fmt.Fprintf(stdout, "wrote a new certificate to %s, valid until %s\n",
		name, cert.NotAfter.Format(time.DateOnly))

	return err
}

func operatorToken(d deps, variable string) (string, error) {
	token := d.getenv(variable)
	if token == "" {
		return "", fmt.Errorf("%w: the variable %s is not set", errUsage, variable)
	}

	return token, nil
}

func participantsAdd(ctx context.Context, args []string, stdout io.Writer, d deps) error {
	fs := newFlags("participants add")
	registerURL := fs.String("register-url", "", "the URL of the Register operator interface")
	file := fs.String("file", "", "the participant load file")

	if err := parse(fs, args, map[string]*string{"register-url": registerURL, "file": file}); err != nil {
		return err
	}

	token, data, err := loadInputs(d, "REGISTER_OPERATOR_TOKEN", *file)
	if err != nil {
		return err
	}

	load, err := register.New(*registerURL, token, d.httpClient).LoadParticipants(ctx, data)
	if err != nil {
		return err
	}

	var out strings.Builder
	for _, id := range load.Added {
		fmt.Fprintf(&out, "added %s\n", id)
	}

	for _, id := range load.Existing {
		fmt.Fprintf(&out, "existing %s\n", id)
	}

	_, err = io.WriteString(stdout, out.String())

	return err
}

func customersAdd(ctx context.Context, args []string, stdout io.Writer, d deps) error {
	fs := newFlags("customers add")
	holderURL := fs.String("dataholder-url", "", "the URL of the operator interface of the data holder brand")
	file := fs.String("file", "", "the customer load file")

	if err := parse(fs, args, map[string]*string{"dataholder-url": holderURL, "file": file}); err != nil {
		return err
	}

	token, data, err := loadInputs(d, "DATAHOLDER_OPERATOR_TOKEN", *file)
	if err != nil {
		return err
	}

	load, err := dataholder.New(*holderURL, token, d.httpClient).LoadCustomers(ctx, data)
	if err != nil {
		return err
	}

	noun := "customers"
	if load.Created == 1 {
		noun = "customer"
	}

	_, err = fmt.Fprintf(stdout, "%d %s created, %d existed\n", load.Created, noun, load.Existed)

	return err
}

func loadInputs(d deps, tokenVariable, file string) (token string, data []byte, err error) {
	token, err = operatorToken(d, tokenVariable)
	if err != nil {
		return "", nil, err
	}

	data, err = os.ReadFile(filepath.Clean(file))
	if err != nil {
		return "", nil, fmt.Errorf("read the load file: %w", err)
	}

	return token, data, nil
}

func readCSR(path string) ([]byte, error) {
	csr, found, err := pemfile.Read(path)
	if err != nil {
		return nil, fmt.Errorf("read the CSR file: %w", err)
	}

	if !found {
		return nil, fmt.Errorf("the CSR file %s is absent", path)
	}

	return csr, nil
}

func keyPerm(anyUser bool) os.FileMode {
	if anyUser {
		return pemfile.AnyUser
	}

	return pemfile.OwnerOnly
}

func parseCertIssue(args []string, now func() time.Time) (certissue.Config, certOutput, string, error) {
	fs := newFlags("cert issue")
	registerURL := fs.String("register-url", "", "the URL of the Register operator interface")
	kindName := fs.String("kind", "", "server or client")
	legalEntity := fs.String("legal-entity-id", "", "the ID of the legal entity")
	brand := fs.String("brand-id", "", "the ID of the data holder brand, for a server certificate")
	product := fs.String("software-product-id", "", "the ID of the software product, for a client certificate")
	cn := fs.String("cn", "", "the CN of the subject")
	org := fs.String("o", "", "the O of the subject")

	var output certOutput
	fs.StringVar(&output.secret, "secret", "", "the name of the TLS Secret")
	fs.StringVar(&output.cert, "cert", "", "the certificate file")
	fs.StringVar(&output.key, "key", "", "the key file")
	fs.StringVar(&output.csr, "csr", "", "a PEM CSR file to send in place of a new key and CSR")
	fs.BoolVar(&output.anyUser, "key-readable-by-all", false, "let any user read the key file")
	renewBefore := fs.Duration("renew-before", certcheck.DefaultRenewBefore,
		"get a new certificate when the current one expires within this time")

	var dns dnsNames
	fs.Var(&dns, "dns", "a DNS name of a server certificate (repeat for each name)")

	required := map[string]*string{"register-url": registerURL, "kind": kindName, "legal-entity-id": legalEntity}
	if err := parse(fs, args, required); err != nil {
		return certissue.Config{}, certOutput{}, "", err
	}

	if err := checkCertOutput(fs, output, cn, org); err != nil {
		return certissue.Config{}, certOutput{}, "", err
	}

	cfg := certissue.Config{
		LegalEntityID: *legalEntity, CommonName: *cn, Organisation: *org, DNSNames: dns,
		RenewBefore: *renewBefore, Now: now,
	}

	switch {
	case *kindName == "server" && *brand != "" && *product == "":
		cfg.Kind, cfg.RecordID = ca.Server, *brand
	case *kindName == "client" && *product != "" && *brand == "" && len(dns) == 0:
		cfg.Kind, cfg.RecordID = ca.Client, *product
	default:
		return certissue.Config{}, certOutput{}, "", fmt.Errorf(
			"%w: -kind server needs -brand-id, and -kind client needs -software-product-id and no -dns", errUsage)
	}

	if *renewBefore <= 0 {
		return certissue.Config{}, certOutput{}, "", fmt.Errorf("%w: -renew-before must be more than 0", errUsage)
	}

	return cfg, output, *registerURL, nil
}

func checkCertOutput(fs *flag.FlagSet, output certOutput, cn, org *string) error {
	if (output.secret == "") == (output.cert == "") {
		return fmt.Errorf("%w: give -secret or -cert, not both", errUsage)
	}

	if output.secret != "" {
		return refuseFlags(fs, "-secret cannot go with", "key", "csr", "key-readable-by-all")
	}

	if output.csr != "" {
		// The CSR gives the subject and the DNS names, and the user keeps the key (CLI-0128).
		return refuseFlags(fs, "-csr cannot go with", "cn", "o", "dns", "key", "key-readable-by-all")
	}

	return requireFlags(fs, map[string]*string{"cn": cn, "o": org, "key": &output.key})
}

func publicCertIssue(_ context.Context, args []string, stdout io.Writer, d deps) error {
	fs := newFlags("public-cert issue")
	dir := fs.String("dir", "", "the directory of the files of the CA set-up")
	cert := fs.String("cert", "", "the certificate file")
	key := fs.String("key", "", "the key file")
	anyUser := fs.Bool("key-readable-by-all", false, "let any user read the key file")
	renewBefore := fs.Duration("renew-before", certcheck.DefaultRenewBefore,
		"issue a new certificate when the current one expires within this time")

	if err := parseFlags(fs, args); err != nil {
		return err
	}

	if err := requireFlags(fs, map[string]*string{"dir": dir, "cert": cert, "key": key}); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("%w: give at least one DNS name", errUsage)
	}

	if *renewBefore <= 0 {
		return fmt.Errorf("%w: -renew-before must be more than 0", errUsage)
	}

	issued, err := publiccert.Run(publiccert.Config{
		Dir: *dir, Cert: *cert, Key: *key, KeyPerm: keyPerm(*anyUser), Names: fs.Args(),
		RenewBefore: *renewBefore, Now: d.now,
	})
	if err != nil {
		return err
	}

	if issued == nil {
		_, err = fmt.Fprintf(stdout, "%s has a current certificate: no change\n", *cert)
		return err
	}

	_, err = fmt.Fprintf(stdout, "wrote a new certificate to %s, valid until %s\n",
		*cert, issued.NotAfter.Format(time.DateOnly))

	return err
}
