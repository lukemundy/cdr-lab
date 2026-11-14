package domain

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxNameCharacters = 100

// A customer number and the account number of an open on the online bank site have 8 decimal digits.
const randomNumberDigits = 8

var randomNumbers = uint64(math.Pow10(randomNumberDigits))

// NewCustomerNumber makes a customer number from a random value, which the app layer gives.
func NewCustomerNumber(random uint64) string {
	return randomNumber(random)
}

// NewAccountNumber makes the account number of an open from a random value, which the app layer gives.
func NewAccountNumber(random uint64) string {
	return randomNumber(random)
}

func randomNumber(random uint64) string {
	return fmt.Sprintf("%0*d", randomNumberDigits, random%randomNumbers)
}

func ValidBSB(bsb string) bool {
	return bsbDigits.MatchString(bsb)
}

var (
	bsbDigits           = regexp.MustCompile(`^[0-9]{6}$`)
	accountNumberDigits = regexp.MustCompile(`^[0-9]{6,10}$`)
	// The Mailbox of RFC 5321, section 4.1.2, without the quoted local part and the address literal, which a lab
	// customer does not need.
	emailAddress = regexp.MustCompile(
		`^[A-Za-z0-9!#$%&'*+/=?^_{|}~-]+(\.[A-Za-z0-9!#$%&'*+/=?^_{|}~-]+)*` +
			`@[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?)+$`,
	)
)

// Customer is a person, and the data holder creates its CustomerNumber.
type Customer struct {
	CustomerNumber string `json:"customerNumber"`
	FirstName      string `json:"firstName,omitempty"`
	LastName       string `json:"lastName"`
	EmailAddress   string `json:"emailAddress"`
	// LastUpdated is the time of the last change. A customer does not change, so it is the time of its creation.
	LastUpdated time.Time `json:"-"`
}

var ErrAccountClosed = errors.New("the account is closed")

// OpenStatus has the values of openStatus in the CDS BankingAccountV3.
type OpenStatus string

const (
	OpenStatusOpen   OpenStatus = "OPEN"
	OpenStatusClosed OpenStatus = "CLOSED"
)

// Account has an AccountID that the data holder creates and that a data recipient does not get.
type Account struct {
	AccountID       string          `json:"accountId"`
	DisplayName     string          `json:"displayName"`
	ProductCategory ProductCategory `json:"productCategory"`
	ProductName     string          `json:"productName"`
	BSB             string          `json:"bsb"`
	AccountNumber   string          `json:"accountNumber"`
	CreditLimit     Amount          `json:"creditLimit"`
	OpenStatus      OpenStatus      `json:"openStatus"`
	CreatedAt       time.Time       `json:"-"`
}

// Close gives ErrAccountClosed, because a closed account cannot open again.
func (a Account) Close() (Account, error) {
	if a.OpenStatus == OpenStatusClosed {
		return Account{}, ErrAccountClosed
	}

	a.OpenStatus = OpenStatusClosed

	return a, nil
}

// AcceptTransaction gives ErrAccountClosed, because a closed account gets no new transaction.
func (a Account) AcceptTransaction() error {
	if a.OpenStatus == OpenStatusClosed {
		return ErrAccountClosed
	}

	return nil
}

// OpenAccounts are the accounts that a customer can select for a consent.
func OpenAccounts(accounts []Account) []Account {
	open := []Account{}

	for _, account := range accounts {
		if account.OpenStatus == OpenStatusOpen {
			open = append(open, account)
		}
	}

	return open
}

// Validate gives a ValidationError, and does not examine the CustomerNumber.
func (c Customer) Validate() error {
	var rules fieldProblems

	rules.check(
		utf8.RuneCountInString(c.FirstName) <= MaxNameCharacters,
		"firstName", "firstName must have %d characters at most", MaxNameCharacters,
	)
	rules.name("lastName", c.LastName)
	rules.check(emailAddress.MatchString(c.EmailAddress), "emailAddress", "emailAddress must be one email address")

	return rules.validationError()
}

const unmaskedDigits = 4

// MaskedNumber has the format of the CDS MaskedAccountString.
func (a Account) MaskedNumber() string {
	masked := max(len(a.AccountNumber)-unmaskedDigits, 0)

	return "xxx-xxx " + strings.Repeat("x", masked) + a.AccountNumber[masked:]
}

// Validate gives a ValidationError, and does not examine the AccountID, the OpenStatus and the CreatedAt, which the
// data holder sets.
func (a Account) Validate() error {
	var rules fieldProblems

	rules.name("displayName", a.DisplayName)
	rules.check(
		slices.Contains(productCategories, a.ProductCategory),
		"productCategory", "productCategory must be a value of BankingProductCategoryV2",
	)
	rules.name("productName", a.ProductName)
	rules.check(ValidBSB(a.BSB), "bsb", "bsb must have 6 digits")
	rules.check(
		accountNumberDigits.MatchString(a.AccountNumber), "accountNumber", "accountNumber must have 6 to 10 digits",
	)
	rules.check(a.CreditLimit >= 0, "creditLimit", "creditLimit must be zero or more")

	return rules.validationError()
}
