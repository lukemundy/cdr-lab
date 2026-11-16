package smtpmail_test

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/smtpmail"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// received is one message that the test SMTP server got.
type received struct {
	from, to string
	data     string
}

// smtpServer is a minimal SMTP server that accepts one message for each connection.
type smtpServer struct {
	address  string
	messages chan received
	// refuseRecipient makes the server answer RCPT with a permanent error.
	refuseRecipient bool
}

func newSMTPServer(t *testing.T) *smtpServer {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	server := &smtpServer{address: listener.Addr().String(), messages: make(chan received, 4)}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go server.serve(conn)
		}
	}()

	return server
}

func (s *smtpServer) serve(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	reply := func(line string) { _, _ = conn.Write([]byte(line + "\r\n")) }

	var message received

	reply("220 test ESMTP")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		command := strings.TrimRight(line, "\r\n")

		switch {
		case strings.HasPrefix(command, "EHLO"), strings.HasPrefix(command, "HELO"):
			reply("250 test")
		case strings.HasPrefix(command, "MAIL FROM:"):
			message.from = strings.Trim(strings.TrimPrefix(command, "MAIL FROM:"), "<>")

			reply("250 ok")
		case strings.HasPrefix(command, "RCPT TO:") && s.refuseRecipient:
			reply("550 no such mailbox")
		case strings.HasPrefix(command, "RCPT TO:"):
			message.to = strings.Trim(strings.TrimPrefix(command, "RCPT TO:"), "<>")

			reply("250 ok")
		case command == "DATA":
			reply("354 send the message")

			message.data = readData(reader)
			s.messages <- message

			reply("250 accepted")
		case command == "QUIT":
			reply("221 bye")
			return
		default:
			reply("502 not implemented")
		}
	}
}

func readData(reader *bufio.Reader) string {
	var data strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == ".\r\n" {
			return data.String()
		}

		// The client adds a dot before each line that starts with a dot (RFC 5321, section 4.5.2).
		data.WriteString(strings.TrimPrefix(line, "."))
	}
}

var coastal = smtpmail.Brand{
	Name: "Coastal Bank",
	Logo: pngLogo(480, 96),
	Colours: smtpmail.Colours{
		Brand: "#1d4ed8", Page: "#f5f5f5", Surface: "#ffffff", Border: "#d4d4d4",
		Text: "#1a1a1a", Heading: "#0a0a0a", Muted: "#525252",
	},
}

func pngLogo(width, height int) []byte {
	var logo bytes.Buffer
	if err := png.Encode(&logo, image.NewNRGBA(image.Rect(0, 0, width, height))); err != nil {
		panic(err)
	}

	return logo.Bytes()
}

var otpEmail = domain.Email{
	Subject:    "Your one-time password",
	Heading:    "Your one-time password is 042137.",
	Paragraphs: []string{"It is valid for 5 minutes.", "Budget App does not need your password."},
}

func newMailer(t *testing.T, address string, brand smtpmail.Brand) smtpmail.Mailer {
	t.Helper()

	mailer, err := smtpmail.New(address, "no-reply@bank.test", brand)
	require.NoError(t, err)

	return mailer
}

// email is a message that the test SMTP server got, with its parts.
type email struct {
	received

	header mail.Header
	// raw is the data of the message as the server got it.
	raw        string
	text       string
	html       string
	logo       []byte
	logoHeader mail.Header
	// related is the Content-Type of the HTML part and the logo.
	related string
}

func send(t *testing.T, brand smtpmail.Brand, message domain.Email) email {
	t.Helper()

	server := newSMTPServer(t)
	require.NoError(t, newMailer(t, server.address, brand).Send(t.Context(), "alex@example.com", message))

	return parse(t, <-server.messages)
}

func parse(t *testing.T, r received) email {
	t.Helper()

	msg, err := mail.ReadMessage(strings.NewReader(r.data))
	require.NoError(t, err, r.data)

	e := email{received: r, header: msg.Header, raw: r.data}

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "multipart/alternative", mediaType)

	parts := multipart.NewReader(msg.Body, params["boundary"])

	text, err := parts.NextPart()
	require.NoError(t, err)
	assert.Equal(t, "text/plain; charset=utf-8", text.Header.Get("Content-Type"))
	e.text = readAll(t, text)

	related, err := parts.NextPart()
	require.NoError(t, err)

	e.related = related.Header.Get("Content-Type")
	mediaType, params, err = mime.ParseMediaType(e.related)
	require.NoError(t, err)
	require.Equal(t, "multipart/related", mediaType)

	relatedParts := multipart.NewReader(related, params["boundary"])

	html, err := relatedParts.NextPart()
	require.NoError(t, err)
	assert.Equal(t, "text/html; charset=utf-8", html.Header.Get("Content-Type"))
	e.html = readAll(t, html)

	logo, err := relatedParts.NextPart()
	require.NoError(t, err)

	e.logoHeader = mail.Header(logo.Header)
	e.logo = []byte(readAll(t, logo))

	_, err = relatedParts.NextPart()
	require.ErrorIs(t, err, io.EOF, "the logo is the only image")

	_, err = parts.NextPart()
	require.ErrorIs(t, err, io.EOF, "the email has only the text part and the HTML part")

	return e
}

// readAll decodes the part: the multipart reader decodes quoted-printable, and the test decodes base64.
func readAll(t *testing.T, part *multipart.Part) string {
	t.Helper()

	data, err := io.ReadAll(part)
	require.NoError(t, err)

	if part.Header.Get("Content-Transfer-Encoding") == "base64" {
		decoded, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, bytes.NewReader(data)))
		require.NoError(t, err)

		return string(decoded)
	}

	return string(data)
}

// Verifies: DH-1262
func TestAnEmailGoesToTheSMTPServerFromTheBrandNameWithItsRecipientAndSubject(t *testing.T) {
	e := send(t, coastal, otpEmail)

	assert.Equal(t, "no-reply@bank.test", e.from)
	assert.Equal(t, "alex@example.com", e.to)
	assert.Equal(t, `"Coastal Bank" <no-reply@bank.test>`, e.header.Get("From"))
	assert.Equal(t, "alex@example.com", e.header.Get("To"))
	assert.Equal(t, "Your one-time password", e.header.Get("Subject"))
	assert.Equal(t, "1.0", e.header.Get("MIME-Version"))
	assert.NotEmpty(t, e.header.Get("Date"))
}

// Verifies: DH-1257
func TestTheTextPartGivesTheHeadingTheParagraphsAndTheDemonstrationNotice(t *testing.T) {
	e := send(t, coastal, otpEmail)

	assert.Equal(t, "Your one-time password is 042137.\r\n\r\n"+
		"It is valid for 5 minutes.\r\n\r\n"+
		"Budget App does not need your password.\r\n\r\n"+
		domain.DemonstrationNotice("Coastal Bank")+"\r\n", e.text)
}

// Verifies: DH-1247
func TestTheHTMLPartGivesTheSameInformationAsTheTextPart(t *testing.T) {
	e := send(t, coastal, otpEmail)

	assert.Contains(t, e.html, "<h1")
	assert.Contains(t, e.html, ">Your one-time password is 042137.</h1>")
	assert.Contains(t, e.html, ">It is valid for 5 minutes.</p>")
	assert.Contains(t, e.html, ">Budget App does not need your password.</p>")
	assert.Contains(t, e.html, ">"+domain.DemonstrationNotice("Coastal Bank")+"</p>")
	assert.Contains(t, e.html, "<title>Your one-time password</title>")
	assert.Contains(t, e.html, `<html lang="en-AU">`)
}

// Verifies: DH-1248, DH-1252
func TestTheHTMLPartShowsTheLogoOfTheBrandFromAPartOfTheEmail(t *testing.T) {
	e := send(t, coastal, otpEmail)

	assert.Equal(t, coastal.Logo, e.logo)
	assert.Equal(t, "image/png", e.logoHeader.Get("Content-Type"))
	assert.Equal(t, "<logo@bank.test>", e.logoHeader.Get("Content-Id"))
	assert.Contains(t, e.related, `type="text/html"`)
	assert.Contains(t, e.html, `<img src="cid:logo@bank.test" alt="Coastal Bank" width="240" height="48"`)
	assert.Contains(t, e.html, "background-color:#1d4ed8")
}

// The email refers to no other document and loads no content from a different location.
// Verifies: DH-1251, DH-1256
func TestTheHTMLPartHasNoLinkAndNoRemoteContent(t *testing.T) {
	e := send(t, coastal, otpEmail)

	assert.NotContains(t, e.html, "<a ")
	assert.NotContains(t, e.html, "http:")
	assert.NotContains(t, e.html, "https:")
	assert.NotContains(t, e.html, "<style")
	assert.NotContains(t, e.html, "<link")
	assert.NotContains(t, e.html, "<script")
}

// Verifies: DH-1259
func TestEachLayoutTableOfTheHTMLPartHasThePresentationRole(t *testing.T) {
	e := send(t, coastal, otpEmail)

	assert.Positive(t, strings.Count(e.html, "<table"))
	assert.Equal(t, strings.Count(e.html, "<table"), strings.Count(e.html, `<table role="presentation"`))
}

func TestTheBrandNameAndTheContentAreEscapedInTheHTMLPart(t *testing.T) {
	brand := coastal
	brand.Name = "Coastal <Test> & Bank"
	message := otpEmail
	message.Paragraphs = []string{"<b>bold</b>"}

	e := send(t, brand, message)

	assert.Contains(t, e.html, `alt="Coastal &lt;Test&gt; &amp; Bank"`)
	assert.Contains(t, e.html, "&lt;b&gt;bold&lt;/b&gt;")
	assert.NotContains(t, e.html, "<b>")
}

// Verifies: DH-1249
func TestNoLineOfTheMessageIsLongerThan998Characters(t *testing.T) {
	message := otpEmail
	message.Paragraphs = []string{strings.Repeat("long ", 400)}

	e := send(t, coastal, message)

	for line := range strings.SplitSeq(e.raw, "\r\n") {
		assert.LessOrEqual(t, len(line), 998, line)
	}

	assert.Contains(t, e.text, strings.Repeat("long ", 400), "the long line is a soft line break only")
}

// Verifies: DH-1245
func TestABrandNameWithSpecialCharactersIsOneValidFromHeader(t *testing.T) {
	for name, brandName := range map[string]string{
		"specials":       `Wattle, "Test" Bank`,
		"not ASCII":      "Banque Café",
		"line break":     "Wattle Bank\r\nBcc: attacker@example.com",
		"line feed only": "Wattle Bank\nBcc: attacker@example.com",
		"encoded-word":   "=?utf-8?q?Bcc=3A_x?=",
		"angle brackets": "Wattle <attacker@example.com>",
	} {
		t.Run(name, func(t *testing.T) {
			brand := coastal
			brand.Name = brandName

			e := send(t, brand, otpEmail)

			headers, _, _ := strings.Cut(e.raw, "\r\n\r\n")
			assert.NotContains(t, headers, "\r\nBcc:")
			assert.NotContains(t, headers, "\nBcc:")
			assert.Equal(t, 6, strings.Count(headers, "\r\n")+1, "the message has only the headers of the mailer")

			from, err := mail.ParseAddress(e.header.Get("From"))
			require.NoError(t, err)
			assert.Equal(t, "no-reply@bank.test", from.Address)
		})
	}
}

// Verifies: DH-1244
func TestTheDisplayNameOfTheSenderIsTheBrandName(t *testing.T) {
	for _, brandName := range []string{`Wattle, "Test" Bank`, "Banque Café"} {
		brand := coastal
		brand.Name = brandName

		from, err := mail.ParseAddress(send(t, brand, otpEmail).header.Get("From"))

		require.NoError(t, err)
		assert.Equal(t, brandName, from.Name)
	}
}

func TestALineBreakInTheSubjectDoesNotAddAHeader(t *testing.T) {
	message := otpEmail
	message.Subject = "Share data with Budget\r\nBcc: attacker@example.com"

	e := send(t, coastal, message)

	headers, _, _ := strings.Cut(e.raw, "\r\n\r\n")
	assert.NotContains(t, headers, "\r\nBcc:")
	assert.Equal(t, 6, strings.Count(headers, "\r\n")+1, "the message has only the headers of the mailer")
}

func TestASubjectWithCharactersThatAreNotASCIIIsEncoded(t *testing.T) {
	message := otpEmail
	message.Subject = "Données de Café"

	assert.Contains(t, send(t, coastal, message).raw, "Subject: =?utf-8?q?Donn=C3=A9es_de_Caf=C3=A9?=\r\n")
}

func TestALogoThatIsNotAPNGIsAnError(t *testing.T) {
	brand := coastal
	brand.Logo = []byte("GIF89a")

	_, err := smtpmail.New("127.0.0.1:25", "no-reply@bank.test", brand)

	require.ErrorContains(t, err, "the logo is not a PNG image")
}

func TestARecipientThatTheServerRefusesIsAnError(t *testing.T) {
	server := newSMTPServer(t)
	server.refuseRecipient = true

	err := newMailer(t, server.address, coastal).Send(t.Context(), "alex@example.com", otpEmail)

	require.ErrorContains(t, err, "give the recipient to the SMTP server")
	assert.Empty(t, server.messages)
}

func TestAServerThatDoesNotListenIsAnError(t *testing.T) {
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	err = newMailer(t, address, coastal).Send(t.Context(), "alex@example.com", otpEmail)

	require.ErrorContains(t, err, "connect to the SMTP server")
}
