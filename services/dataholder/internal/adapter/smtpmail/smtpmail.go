// Package smtpmail sends the emails of the data holder through an SMTP server, with the brand of the data holder.
package smtpmail

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"image/png"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	sendTimeout = 10 * time.Second
	// logoHeight is the height of the logo in CSS pixels. The PNG has 2 times this height.
	logoHeight = 48
	// base64LineLength is the line length of RFC 2045, section 6.8.
	base64LineLength = 76
)

//go:embed templates/otp.html
var templateFiles embed.FS

var page = template.Must(template.ParseFS(templateFiles, "templates/otp.html"))

type Brand struct {
	// Name is the brand name, which is also the display name of the sender.
	Name string
	// Logo is a PNG image, 2 times the height of the logo in the email.
	Logo    []byte
	Colours Colours
}

// Colours are hex colours of the brand. Each text colour is legible on its background: Text and Heading on Surface,
// and Muted on Page. The logo is legible on Brand.
type Colours struct {
	Brand, Page, Surface, Border, Text, Heading, Muted string
}

// RoleColours gives the colours from the hex colours of the design system roles. Each pair of the email is a pair that
// the contrast test of the design system checks.
func RoleColours(roles map[string]string) Colours {
	return Colours{
		Brand: roles["brand"], Page: roles["page"], Surface: roles["surface"], Border: roles["border"],
		Text: roles["fg"], Heading: roles["heading"], Muted: roles["fg-muted"],
	}
}

// Mailer uses SMTP with no TLS and no authentication.
type Mailer struct {
	address   string
	from      string
	envelope  string
	brand     Brand
	logoID    string
	logoWidth int
	now       func() time.Time
}

var _ app.Mailer = Mailer{}

// New gives a Mailer for the SMTP server at address (host:port), which sends each email from the address from, with
// the brand name as the display name.
func New(address, from string, brand Brand) (Mailer, error) {
	logo, err := png.DecodeConfig(bytes.NewReader(brand.Logo))
	if err != nil {
		return Mailer{}, fmt.Errorf("the logo is not a PNG image: %w", err)
	}

	_, domainPart, _ := strings.Cut(from, "@")

	return Mailer{
		address:   address,
		from:      (&mail.Address{Name: brand.Name, Address: from}).String(),
		envelope:  from,
		brand:     brand,
		logoID:    "logo@" + domainPart,
		logoWidth: logo.Width * logoHeight / logo.Height,
		now:       time.Now,
	}, nil
}

func (m Mailer) Send(ctx context.Context, to string, email domain.Email) error {
	message, err := m.message(to, email)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", m.address)
	if err != nil {
		return fmt.Errorf("connect to the SMTP server: %w", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	host, _, _ := net.SplitHostPort(m.address)

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()

		return fmt.Errorf("start the SMTP session: %w", err)
	}

	defer func() { _ = client.Close() }()

	if err := m.transfer(client, to, message); err != nil {
		return err
	}

	return client.Quit()
}

func (m Mailer) transfer(client *smtp.Client, to string, message []byte) error {
	if err := client.Mail(m.envelope); err != nil {
		return fmt.Errorf("give the sender to the SMTP server: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("give the recipient to the SMTP server: %w", err)
	}

	body, err := client.Data()
	if err != nil {
		return fmt.Errorf("start the message: %w", err)
	}

	if _, err := body.Write(message); err != nil {
		return fmt.Errorf("write the message: %w", err)
	}

	if err := body.Close(); err != nil {
		return fmt.Errorf("end the message: %w", err)
	}

	return nil
}

// message encodes the From header and the subject, so that a line break in them cannot add a header, and each part,
// so that no line is longer than 998 characters (RFC 5322, section 2.1.1).
func (m Mailer) message(to string, email domain.Email) ([]byte, error) {
	var html bytes.Buffer
	if err := page.Execute(&html, m.view(email)); err != nil {
		return nil, fmt.Errorf("make the HTML part: %w", err)
	}

	var body bytes.Buffer

	alternative := multipart.NewWriter(&body)

	headers := []string{
		"From: " + m.from,
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("utf-8", email.Subject),
		"Date: " + m.now().UTC().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: " + mime.FormatMediaType("multipart/alternative", map[string]string{
			"boundary": alternative.Boundary(),
		}),
	}

	if err := writeQuotedPrintable(alternative, "text/plain; charset=utf-8", text(email, m.brand.Name)); err != nil {
		return nil, fmt.Errorf("make the text part: %w", err)
	}

	if err := m.writeRelated(alternative, html.Bytes()); err != nil {
		return nil, fmt.Errorf("make the HTML part: %w", err)
	}

	if err := alternative.Close(); err != nil {
		return nil, fmt.Errorf("make the message: %w", err)
	}

	return append([]byte(strings.Join(headers, "\r\n")+"\r\n\r\n"), body.Bytes()...), nil
}

// writeRelated writes the HTML part and the logo, which the HTML part refers to with a cid: URL (RFC 2392).
func (m Mailer) writeRelated(alternative *multipart.Writer, html []byte) error {
	boundary := multipart.NewWriter(nil).Boundary()

	part, err := alternative.CreatePart(textproto.MIMEHeader{"Content-Type": {mime.FormatMediaType(
		"multipart/related", map[string]string{"boundary": boundary, "type": "text/html"},
	)}})
	if err != nil {
		return err
	}

	related := multipart.NewWriter(part)
	if err := related.SetBoundary(boundary); err != nil {
		return err
	}

	if err := writeQuotedPrintable(related, "text/html; charset=utf-8", html); err != nil {
		return err
	}

	logo, err := related.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"image/png"},
		"Content-Transfer-Encoding": {"base64"},
		"Content-Id":                {"<" + m.logoID + ">"},
		"Content-Disposition":       {`inline; filename="logo.png"`},
	})
	if err != nil {
		return err
	}

	if _, err := logo.Write(base64Lines(m.brand.Logo)); err != nil {
		return err
	}

	return related.Close()
}

func writeQuotedPrintable(w *multipart.Writer, contentType string, data []byte) error {
	part, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {contentType},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return err
	}

	encoder := quotedprintable.NewWriter(part)
	if _, err := encoder.Write(data); err != nil {
		return err
	}

	return encoder.Close()
}

// text gives the text part, with CRLF line ends.
func text(email domain.Email, bank string) []byte {
	paragraphs := append(append([]string{email.Heading}, email.Paragraphs...), domain.DemonstrationNotice(bank))

	return []byte(strings.Join(paragraphs, "\r\n\r\n") + "\r\n")
}

type view struct {
	domain.Email

	Bank      string
	Footer    string
	Colours   Colours
	LogoID    string
	LogoWidth int
}

func (m Mailer) view(email domain.Email) view {
	return view{
		Email:     email,
		Bank:      m.brand.Name,
		Footer:    domain.DemonstrationNotice(m.brand.Name),
		Colours:   m.brand.Colours,
		LogoID:    m.logoID,
		LogoWidth: m.logoWidth,
	}
}

func base64Lines(data []byte) []byte {
	encoded := base64.StdEncoding.EncodeToString(data)

	var lines bytes.Buffer

	for len(encoded) > base64LineLength {
		lines.WriteString(encoded[:base64LineLength] + "\r\n")
		encoded = encoded[base64LineLength:]
	}

	lines.WriteString(encoded + "\r\n")

	return lines.Bytes()
}
