package domain

import (
	"regexp"
	"slices"
	"time"
	"unicode/utf8"
)

const maxDescriptionCharacters = 500

type TransactionType string

const (
	TransactionDirectDebit      TransactionType = "DIRECT_DEBIT"
	TransactionFee              TransactionType = "FEE"
	TransactionInterestCharged  TransactionType = "INTEREST_CHARGED"
	TransactionInterestPaid     TransactionType = "INTEREST_PAID"
	TransactionOther            TransactionType = "OTHER"
	TransactionPayment          TransactionType = "PAYMENT"
	TransactionTransferIncoming TransactionType = "TRANSFER_INCOMING"
	TransactionTransferOutgoing TransactionType = "TRANSFER_OUTGOING"
)

var transactionTypes = []TransactionType{
	TransactionDirectDebit, TransactionFee, TransactionInterestCharged, TransactionInterestPaid, TransactionOther,
	TransactionPayment, TransactionTransferIncoming, TransactionTransferOutgoing,
}

type TransactionStatus string

const (
	TransactionPending TransactionStatus = "PENDING"
	TransactionPosted  TransactionStatus = "POSTED"
)

var transactionStatuses = []TransactionStatus{TransactionPending, TransactionPosted}

type NPPService string

const (
	NPPServiceX2P1   NPPService = "X2P1"
	NPPServiceIFTI   NPPService = "IFTI"
	NPPServiceBSCT   NPPService = "BSCT"
	NPPServiceCATSCT NPPService = "CATSCT"
)

var nppServices = []NPPService{NPPServiceX2P1, NPPServiceIFTI, NPPServiceBSCT, NPPServiceCATSCT}

// ExtensionNPPPayload is the only extensionUType of the CDS.
const ExtensionNPPPayload = "nppPayload"

var serviceVersionDigits = regexp.MustCompile(`^[0-9]{2}$`)

// ExtendedData has the fields of the extendedData of the CDS BankingTransactionDetailV3 that the data holder holds.
type ExtendedData struct {
	Payer          string     `json:"payer,omitempty"`
	Payee          string     `json:"payee,omitempty"`
	ExtensionUType string     `json:"extensionUType"`
	NPPPayload     NPPPayload `json:"nppPayload"`
}

type NPPPayload struct {
	ExtendedDescription string     `json:"extendedDescription"`
	Service             NPPService `json:"service"`
	ServiceVersion      string     `json:"serviceVersion"`
}

// Transaction has a TransactionID that the data holder creates and that a data recipient does not get.
type Transaction struct {
	TransactionID        string            `json:"transactionId"`
	Type                 TransactionType   `json:"type"`
	Status               TransactionStatus `json:"status"`
	Description          string            `json:"description"`
	Amount               Amount            `json:"amount"`
	Reference            string            `json:"reference"`
	PostingDateTime      *time.Time        `json:"postingDateTime,omitempty"`
	ExecutionDateTime    *time.Time        `json:"executionDateTime,omitempty"`
	ValueDateTime        *time.Time        `json:"valueDateTime,omitempty"`
	MerchantName         string            `json:"merchantName,omitempty"`
	MerchantCategoryCode string            `json:"merchantCategoryCode,omitempty"`
	BillerCode           string            `json:"billerCode,omitempty"`
	BillerName           string            `json:"billerName,omitempty"`
	CRN                  string            `json:"crn,omitempty"`
	APCANumber           string            `json:"apcaNumber,omitempty"`
	ExtendedData         *ExtendedData     `json:"extendedData,omitempty"`
}

// EffectiveTime is valid only for a transaction that Validate accepts.
func (t Transaction) EffectiveTime() time.Time {
	if t.PostingDateTime != nil {
		return *t.PostingDateTime
	}

	return *t.ExecutionDateTime
}

func (t Transaction) InUTC() Transaction {
	for _, at := range []**time.Time{&t.PostingDateTime, &t.ExecutionDateTime, &t.ValueDateTime} {
		if *at != nil {
			*at = new((*at).UTC())
		}
	}

	return t
}

func (t Transaction) IsDetailAvailable() bool {
	return t.ExtendedData != nil
}

// Validate gives a ValidationError, and does not examine the TransactionID.
func (t Transaction) Validate() error {
	var rules fieldProblems

	rules.check(slices.Contains(transactionTypes, t.Type), "type", "type must be a value of the CDS enumeration")
	rules.check(
		slices.Contains(transactionStatuses, t.Status), "status", "status must be PENDING or POSTED",
	)

	length := utf8.RuneCountInString(t.Description)
	rules.check(
		length >= 1 && length <= maxDescriptionCharacters,
		"description", "description must have 1 to %d characters", maxDescriptionCharacters,
	)
	// A rule of two fields is at the field that is absent.
	rules.check(
		t.Status != TransactionPosted || t.PostingDateTime != nil,
		"postingDateTime", "a POSTED transaction must have a postingDateTime",
	)
	rules.check(
		t.Status != TransactionPending || t.ExecutionDateTime != nil,
		"executionDateTime", "a PENDING transaction must have an executionDateTime",
	)

	if e := t.ExtendedData; e != nil {
		rules.check(
			e.ExtensionUType == ExtensionNPPPayload,
			"extendedData.extensionUType", "extendedData.extensionUType must be nppPayload",
		)
		rules.check(
			slices.Contains(nppServices, e.NPPPayload.Service),
			"extendedData.nppPayload.service", "extendedData.nppPayload.service must be a value of the CDS enumeration",
		)
		rules.check(
			serviceVersionDigits.MatchString(e.NPPPayload.ServiceVersion),
			"extendedData.nppPayload.serviceVersion", "extendedData.nppPayload.serviceVersion must have 2 digits",
		)
		rules.check(
			e.NPPPayload.ExtendedDescription != "",
			"extendedData.nppPayload.extendedDescription",
			"extendedData.nppPayload.extendedDescription must have 1 character or more",
		)
		rules.check(
			t.Amount >= 0 || e.Payee != "",
			"extendedData.payee", "extendedData.payee must be present for a negative amount",
		)
		rules.check(
			t.Amount <= 0 || e.Payer != "",
			"extendedData.payer", "extendedData.payer must be present for a positive amount",
		)
	}

	return rules.validationError()
}

// AccountEnumerations gives the values of each enumeration of an account and of a transaction, by the name of its
// type.
func AccountEnumerations() map[string][]string {
	return map[string][]string{
		"OpenStatus":        {string(OpenStatusOpen), string(OpenStatusClosed)},
		"TransactionType":   values(transactionTypes),
		"TransactionStatus": values(transactionStatuses),
	}
}
