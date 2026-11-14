package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"
	"unicode"
	"unicode/utf8"
)

// Product has the required fields of the CDS BankingProductDetailV7 schema, and the
// optional fields that the data holder holds. The JSON names are the CDS field names.
type Product struct {
	ProductID       string          `json:"productId"`
	LastUpdated     time.Time       `json:"lastUpdated"`
	ProductCategory ProductCategory `json:"productCategory"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Brand           string          `json:"brand"`
	IsTailored      bool            `json:"isTailored"`
	EffectiveFrom   *time.Time      `json:"effectiveFrom,omitempty"`
	EffectiveTo     *time.Time      `json:"effectiveTo,omitempty"`
	Features        []Feature       `json:"features,omitempty"`
	Fees            []Fee           `json:"fees,omitempty"`
	DepositRates    []DepositRate   `json:"depositRates,omitempty"`
	LendingRates    []LendingRate   `json:"lendingRates,omitempty"`
}

type Feature struct {
	FeatureType     FeatureType `json:"featureType"`
	AdditionalValue string      `json:"additionalValue,omitempty"`
	AdditionalInfo  string      `json:"additionalInfo,omitempty"`
}

type Fee struct {
	Name            string         `json:"name"`
	FeeType         FeeType        `json:"feeType"`
	FeeMethodUType  FeeMethodUType `json:"feeMethodUType"`
	FixedAmount     *FeeAmount     `json:"fixedAmount,omitempty"`
	RateBased       *FeeRate       `json:"rateBased,omitempty"`
	Variable        *FeeRange      `json:"variable,omitempty"`
	AdditionalValue string         `json:"additionalValue,omitempty"`
	AdditionalInfo  string         `json:"additionalInfo,omitempty"`
}

type FeeAmount struct {
	Amount string `json:"amount"`
}

type FeeRate struct {
	RateType FeeRateType `json:"rateType"`
	Rate     string      `json:"rate"`
}

type FeeRange struct {
	FeeMinimum string `json:"feeMinimum,omitempty"`
	FeeMaximum string `json:"feeMaximum,omitempty"`
}

type DepositRate struct {
	DepositRateType      DepositRateType `json:"depositRateType"`
	Rate                 string          `json:"rate"`
	ApplicationType      ApplicationType `json:"applicationType"`
	ApplicationFrequency string          `json:"applicationFrequency,omitempty"`
	AdditionalValue      string          `json:"additionalValue,omitempty"`
	AdditionalInfo       string          `json:"additionalInfo,omitempty"`
}

type LendingRate struct {
	LendingRateType      LendingRateType `json:"lendingRateType"`
	Rate                 string          `json:"rate"`
	ComparisonRate       string          `json:"comparisonRate,omitempty"`
	ApplicationType      ApplicationType `json:"applicationType"`
	ApplicationFrequency string          `json:"applicationFrequency,omitempty"`
	RepaymentType        RepaymentType   `json:"repaymentType"`
	LoanPurpose          LoanPurpose     `json:"loanPurpose"`
	AdditionalValue      string          `json:"additionalValue,omitempty"`
	AdditionalInfo       string          `json:"additionalInfo,omitempty"`
}

var (
	// The CDS AmountString and RateString formats.
	amountString = regexp.MustCompile(`^-?[0-9]{1,16}\.[0-9]{2,16}$`)
	rateString   = regexp.MustCompile(`^-?[0-9]{1,16}(\.[0-9]{1,16})?$`)
)

// ParseProduct refuses a field that Product does not have, and its error names each rule that the product breaks.
func ParseProduct(data []byte) (Product, error) {
	var p Product

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&p); err != nil {
		return Product{}, err
	}

	if err := p.Validate(); err != nil {
		return Product{}, err
	}

	return p, nil
}

// Validate returns all the rules that the product breaks, so that the operator can correct a file in one pass.
func (p Product) Validate() error {
	var rules problems

	rules.present("productId", p.ProductID)
	rules.check(isASCII(p.ProductID), "productId has a character that is not ASCII")
	rules.check(!p.LastUpdated.IsZero(), "lastUpdated is mandatory")
	checkEnum(&rules, "productCategory", p.ProductCategory, productCategories)
	rules.present("name", p.Name)
	// The account of an open gets the product name, which has the rule of an account name.
	rules.check(utf8.RuneCountInString(p.Name) <= MaxNameCharacters, "name has more than %d characters",
		MaxNameCharacters)
	rules.present("description", p.Description)
	rules.present("brand", p.Brand)

	for i, f := range p.Features {
		field := fmt.Sprintf("features[%d]", i)

		checkEnum(&rules, field+".featureType", f.FeatureType, featureTypes)
		rules.other(field, f.FeatureType == FeatureOther, f.AdditionalInfo)
	}

	for i, f := range p.Fees {
		rules.fee(fmt.Sprintf("fees[%d]", i), f)
	}

	for i, r := range p.DepositRates {
		field := fmt.Sprintf("depositRates[%d]", i)

		checkEnum(&rules, field+".depositRateType", r.DepositRateType, depositRateTypes)
		rules.rate(field, r.Rate, r.ApplicationType, r.ApplicationFrequency)
	}

	for i, r := range p.LendingRates {
		field := fmt.Sprintf("lendingRates[%d]", i)

		checkEnum(&rules, field+".lendingRateType", r.LendingRateType, lendingRateTypes)
		rules.rate(field, r.Rate, r.ApplicationType, r.ApplicationFrequency)
		rules.check(
			r.ComparisonRate == "" || rateString.MatchString(r.ComparisonRate),
			"%s.comparisonRate is not a RateString", field,
		)
		checkEnum(&rules, field+".repaymentType", r.RepaymentType, repaymentTypes)
		checkEnum(&rules, field+".loanPurpose", r.LoanPurpose, loanPurposes)
	}

	return errors.Join(rules.errs...)
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}

	return true
}

type problems struct {
	errs []error
}

func (p *problems) check(ok bool, format string, args ...any) {
	if !ok {
		p.errs = append(p.errs, fmt.Errorf(format, args...))
	}
}

func (p *problems) present(field, value string) {
	p.check(value != "", "%s is mandatory", field)
}

func checkEnum[T ~string](p *problems, field string, value T, values []T) {
	p.check(slices.Contains(values, value), "%s %q is not a value of the CDS enumeration", field, string(value))
}

func (p *problems) other(field string, isOther bool, additionalInfo string) {
	p.check(!isOther || additionalInfo != "", "%s: additionalInfo is mandatory when the type is OTHER", field)
}

func (p *problems) fee(field string, f Fee) {
	p.present(field+".name", f.Name)
	checkEnum(p, field+".feeType", f.FeeType, feeTypes)
	p.other(field, f.FeeType == FeeTypeOther, f.AdditionalInfo)
	checkEnum(p, field+".feeMethodUType", f.FeeMethodUType, feeMethodUTypes)

	present := map[FeeMethodUType]bool{
		FeeFixedAmount: f.FixedAmount != nil, FeeRateBased: f.RateBased != nil, FeeVariable: f.Variable != nil,
	}
	if has, known := present[f.FeeMethodUType]; known {
		p.check(has, "%s: %s is mandatory when feeMethodUType is %s", field, f.FeeMethodUType, f.FeeMethodUType)
	}

	if f.FixedAmount != nil {
		p.amount(field+".fixedAmount.amount", f.FixedAmount.Amount)
	}

	if f.RateBased != nil {
		checkEnum(p, field+".rateBased.rateType", f.RateBased.RateType, feeRateTypes)
		p.check(rateString.MatchString(f.RateBased.Rate), "%s.rateBased.rate is not a RateString", field)
	}

	if f.Variable != nil {
		p.optionalAmount(field+".variable.feeMinimum", f.Variable.FeeMinimum)
		p.optionalAmount(field+".variable.feeMaximum", f.Variable.FeeMaximum)
	}
}

func (p *problems) amount(field, value string) {
	p.check(amountString.MatchString(value), "%s is not an AmountString", field)
}

func (p *problems) optionalAmount(field, value string) {
	if value != "" {
		p.amount(field, value)
	}
}

func (p *problems) rate(field, rate string, applicationType ApplicationType, applicationFrequency string) {
	p.check(rateString.MatchString(rate), "%s.rate is not a RateString", field)
	checkEnum(p, field+".applicationType", applicationType, applicationTypes)
	p.check(
		applicationType != ApplicationPeriodic || applicationFrequency != "",
		"%s: applicationFrequency is mandatory when applicationType is PERIODIC", field,
	)
}

// ProductID gives the ID of a product that is possibly not valid, or an empty string when it has none.
func ProductID(data []byte) string {
	var p struct {
		ProductID string `json:"productId"`
	}

	_ = json.Unmarshal(data, &p)

	return p.ProductID
}

// DuplicateProductIDs gives each ID that is in ids more than one time, in the order of ids.
func DuplicateProductIDs(ids []string) []string {
	var duplicates []string

	count := map[string]int{}
	for _, id := range ids {
		count[id]++
		if count[id] == 2 {
			duplicates = append(duplicates, id)
		}
	}

	return duplicates
}
