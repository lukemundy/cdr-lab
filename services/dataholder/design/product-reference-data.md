# Mock data holder: Product reference data

This area gives the products of the brand, their load, and the public APIs Get Products, Get Product Detail, Get Status and Get Outages. The endpoint table of the API standards gives the path, the version and the base URI of each API: [Get Products](api-standards.md#dh-0400), [Get Product Detail](api-standards.md#dh-0401), [Get Status](api-standards.md#dh-0402) and [Get Outages](api-standards.md#dh-0403).

## Product fields

- <a id="dh-0200"></a>**DH-0200** A product has the required fields of the `BankingProductDetailV7` schema. ([CDS: BankingProductDetailV7][cds-schema-product-detail])
- <a id="dh-0201"></a>**DH-0201** A product can have these optional fields of the schema, and no other optional field. For example, a product has no `brandGroup`. ([CDS: BankingProductDetailV7][cds-schema-product-detail]; [Judgement: the optional fields of a product](#dh-0260))
  - `effectiveFrom`
  - `effectiveTo`
  - `features`
  - `fees`
  - `depositRates`
  - `lendingRates`
- <a id="dh-0202"></a>**DH-0202** A feature has the fields `featureType`, `additionalValue` and `additionalInfo`. ([CDS: BankingProductDetailV7][cds-schema-product-detail]; [Judgement: the optional fields of a product](#dh-0260))
- <a id="dh-0203"></a>**DH-0203** A fee has the fields `name`, `feeType`, `feeMethodUType`, `fixedAmount`, `rateBased`, `variable`, `additionalValue` and `additionalInfo`. ([CDS: BankingProductDetailV7][cds-schema-product-detail]; [Judgement: the optional fields of a product](#dh-0260))
- <a id="dh-0204"></a>**DH-0204** A deposit rate has the fields `depositRateType`, `rate`, `applicationType`, `applicationFrequency`, `additionalValue` and `additionalInfo`. ([CDS: BankingProductDetailV7][cds-schema-product-detail]; [Judgement: the optional fields of a product](#dh-0260))
- <a id="dh-0205"></a>**DH-0205** A lending rate has the fields of a deposit rate, with `lendingRateType` in place of `depositRateType`. It also has `comparisonRate`, `repaymentType` and `loanPurpose`. ([CDS: BankingProductDetailV7][cds-schema-product-detail]; [Judgement: the optional fields of a product](#dh-0260))
- <a id="dh-0206"></a>**DH-0206** When a field has a CDS enumeration, the data holder accepts each value of that enumeration. For example, a feature accepts each `featureType` of the CDS. ([CDS: BankingProductDetailV7][cds-schema-product-detail])
- <a id="dh-0207"></a>**DH-0207** A rate is a CDS RateString. For example, `0.0125`. ([CDS: Common Field Types][cds-common-field-types])
- <a id="dh-0208"></a>**DH-0208** An amount is a CDS AmountString. For example, `10.00`. ([CDS: Common Field Types][cds-common-field-types])
- <a id="dh-0209"></a>**DH-0209** A product has one of the 13 categories of `BankingProductCategoryV2`. ([CDS: Product Categories][cds-product-categories])
- <a id="dh-0210"></a>**DH-0210** The `lastUpdated`, `effectiveFrom` and `effectiveTo` of a product are RFC 3339 date-times. ([CDS: Common Field Types][cds-common-field-types])
- <a id="dh-0211"></a>**DH-0211** When a load file gives no `isTailored` for a product, the value is `false`. ([CDS: BankingProductV6][cds-schema-product]; [Judgement: an absent `isTailored` is false](#dh-0261))

## Product validity

- <a id="dh-0212"></a>**DH-0212** A `productId` is unique in the brand. ([CDS: BankingProductV6][cds-schema-product])
- <a id="dh-0213"></a>**DH-0213** A `productId` has only ASCII characters. ([CDS: BankingProductV6][cds-schema-product])
- <a id="dh-0214"></a>**DH-0214** A `productId` that is empty is not valid. ([CDS: BankingProductV6][cds-schema-product]; [Judgement: no empty `productId`](#dh-0262))
- <a id="dh-0215"></a>**DH-0215** A `productId` does not follow the ID permanence rules. ([CDS: BankingProductV6][cds-schema-product]; [CDS: ID Permanence][cds-id-permanence])
- <a id="dh-0216"></a>**DH-0216** A product obeys these conditional rules of the schema. ([CDS: BankingProductDetailV7][cds-schema-product-detail])
  - A fee has the object that its `feeMethodUType` names.
  - When a feature or a fee has the type `OTHER`, it has an `additionalInfo`.
  - When a deposit rate or a lending rate has the application type `PERIODIC`, it has an `applicationFrequency`.
- <a id="dh-0217"></a>**DH-0217** The `brand` of a product is the brand name. ([CDR Support: Brands in the CDR ecosystem][support-brands]; [the brand name](brand.md#dh-0304))
- <a id="dh-0218"></a>**DH-0218** A product `name` has 100 characters or fewer. For example, the load refuses a name of 101 characters. ([CDS: BankingProductV6][cds-schema-product]; [Judgement: the maximum length of a product name](#dh-0263))
- <a id="dh-0219"></a>**DH-0219** A product is current when the time of the request is at or after its `effectiveFrom` and before its `effectiveTo`. A date that is absent has no limit. ([CDS: Get Products][cds-get-products]; [Judgement: the effective period of a product](#dh-0264))
- <a id="dh-0220"></a>**DH-0220** A product is future when its `effectiveFrom` is after the time of the request. ([CDS: Get Products][cds-get-products]; [Judgement: the effective period of a product](#dh-0264))

## Preset catalogues

- <a id="dh-0221"></a>**DH-0221** The product data of each preset brand is invented. It copies no product of a real data holder. ([Judgement: invented product data](#dh-0265))
- <a id="dh-0222"></a>**DH-0222** The catalogue of each preset brand is the catalogue of an Australian retail bank. It has products in five or more product categories. ([CDS: Product Categories][cds-product-categories]; [Judgement: the products of the preset catalogues](#dh-0266))
- <a id="dh-0223"></a>**DH-0223** The catalogue of each preset brand has one future product and one product whose `effectiveTo` is in the past. ([CDS: Get Products][cds-get-products]; [Judgement: the products of the preset catalogues](#dh-0266))
- <a id="dh-0224"></a>**DH-0224** Each `additionalValue` of a preset catalogue obeys the rule that the CDS gives for its type. For example, a fee of the type `PERIODIC` has a duration. ([CDS: Product & Account Components][cds-product-components])
- <a id="dh-0225"></a>**DH-0225** The Get Products response and the Get Product Detail response of each product of each preset brand are valid against the CDS 1.36.0 OpenAPI specification. This includes a preset brand that the lab does not run. ([CDS OpenAPI: banking][cds-openapi-banking]; [Judgement: a schema check of each preset catalogue](#dh-0267))

## Product load

- <a id="dh-0226"></a>**DH-0226** The lab operator can load a full set of products into a data holder that has no products. The set is one of these. ([README: Component types][readme-component-types])
  - the product catalogue of the brand
  - the products of a load file
- <a id="dh-0227"></a>**DH-0227** When the data holder has products, the load refuses the set and changes nothing. ([Judgement: no load into a data holder that has products](#dh-0268))
- <a id="dh-0228"></a>**DH-0228** The load is separate from the start of the data holder. The start of the data holder needs no load file. ([README: Component types][readme-component-types])
- <a id="dh-0229"></a>**DH-0229** The load validates the full set against the product fields and the product validity rules of this area. When one product is not valid, the load changes nothing. ([the product fields](#dh-0200); [the product validity rules](#dh-0212))
- <a id="dh-0230"></a>**DH-0230** When a product is not valid, the error of the load names each product and each rule that fails. For example, the error names the second product and its empty `name`. ([the product validity rules](#dh-0212))
- <a id="dh-0231"></a>**DH-0231** The load refuses a product with a field that the data holder does not hold. For example, the load refuses a product with `brandGroup`. ([the optional fields that a product can have](#dh-0201); [Judgement: the optional fields of a product](#dh-0260))
- <a id="dh-0232"></a>**DH-0232** The load refuses a product whose `brand` is not the brand name. ([the brand of a product](#dh-0217))
- <a id="dh-0233"></a>**DH-0233** The load does not compare an `additionalValue` with the rule of its type. ([CDS: Product & Account Components][cds-product-components]; [Judgement: no check of an `additionalValue` in the load](#dh-0269))
- <a id="dh-0234"></a>**DH-0234** A load file uses the field names of the CDS product schemas. For example, a product in the file has `productCategory`. ([CDS: BankingProductDetailV7][cds-schema-product-detail]; [Judgement: the CDS field names in a load file](#dh-0270))
- <a id="dh-0235"></a>**DH-0235** The lab loads the product catalogue of the default preset brand. ([the default preset brand](brand.md#dh-0346); [Judgement: realistic products in the lab](#dh-0271))

## Get Products

- <a id="dh-0236"></a>**DH-0236** Get Products returns the products of the brand as `BankingProductV6` records. ([CDS: Get Products][cds-get-products])
- <a id="dh-0237"></a>**DH-0237** Get Products supports these filters. A product in the response matches each filter of the request. ([CDS: Get Products][cds-get-products])
  - `effective`, with the default `CURRENT`
  - `updated-since`
  - `brand`
  - `product-category`
- <a id="dh-0238"></a>**DH-0238** The `effective` filter gives these products. ([CDS: Get Products][cds-get-products])
  - `CURRENT`: the current products
  - `FUTURE`: the future products
  - `ALL`: all products, also a product whose `effectiveTo` is in the past
- <a id="dh-0239"></a>**DH-0239** The `updated-since` filter gives the products with a `lastUpdated` after that time. ([CDS: Get Products][cds-get-products])
- <a id="dh-0240"></a>**DH-0240** The `brand` filter gives the products whose `brand` is equal to the value. The comparison is exact and case sensitive. For example, `wattle bank` does not match `Wattle Bank`. ([CDS: Get Products][cds-get-products]; [Judgement: an exact comparison for the `brand` filter](#dh-0272))
- <a id="dh-0241"></a>**DH-0241** Get Products returns the products in descending order of `lastUpdated`. ([CDS: Get Products][cds-get-products])
- <a id="dh-0242"></a>**DH-0242** When two products have the same `lastUpdated`, Get Products returns them in the order of their `productId`. ([Judgement: a stable order of the products](#dh-0273))
- <a id="dh-0243"></a>**DH-0243** Get Products supports `page` and `page-size`, with [the CDS pagination rules](../../../pkg/cds/DESIGN.md#cdsl-0017). Each link of a page keeps the filters of the request. ([the filters in each link of a page](../../../pkg/cds/DESIGN.md#cdsl-0029))

## Get Product Detail

- <a id="dh-0244"></a>**DH-0244** Get Product Detail returns one product as a `BankingProductDetailV7` record, in the response envelope. ([CDS: Get Product Detail][cds-get-product-detail]; [the response envelope](../../../pkg/cds/DESIGN.md#cdsl-0034))
- <a id="dh-0245"></a>**DH-0245** Get Product Detail returns a product that is not current. For example, it returns a future product. ([Judgement: the detail of each product that Get Products can list](#dh-0274))

## Get Status and Get Outages

- <a id="dh-0246"></a>**DH-0246** Get Status returns the status `OK`. ([CDS: Get Status][cds-get-status]; [Judgement: no status other than `OK` and no outages](#dh-0275))
- <a id="dh-0247"></a>**DH-0247** The `updateTime` of Get Status is the time of the request. ([CDS: Get Status][cds-get-status]; [CDR Support: Get Status Payload Fields][support-status-fields])
- <a id="dh-0248"></a>**DH-0248** Get Outages returns an empty list of outages. ([CDS: Get Outages][cds-get-outages]; [Judgement: no status other than `OK` and no outages](#dh-0275))
- <a id="dh-0249"></a>**DH-0249** The Get Status response and the Get Outages response have `links.self` and an empty `meta`. ([the response envelope](../../../pkg/cds/DESIGN.md#cdsl-0034))

## Judgements

### <a id="dh-0260"></a>DH-0260 The optional fields of a product

- Kind: choice
- Basis: Step 3.3. The schema marks the other fields as optional, and the CDS gives a data holder more discretion for payloads with no customer ([CDS: Payload Conventions][cds-payload-conventions]).
- Reason: The effective dates, the features, the fees and the rates are sufficient for the most common products. The other fields are `brandName`, `brandGroup`, `applicationUri`, `additionalInformation`, `cardArt`, `bundles`, `constraints`, `eligibility` and `instalments`, and the fee discounts, rate tiers and rate conditions of the components. A field can come in when the lab needs it.

### <a id="dh-0261"></a>DH-0261 An absent `isTailored` is false

- Kind: choice
- Basis: Step 3.3. The CDS requires `isTailored` in the response ([CDS: BankingProductV6][cds-schema-product]), and gives no rule for a load file.
- Reason: The response always has the field. `false` is correct for a product with standard prices.

### <a id="dh-0262"></a>DH-0262 No empty `productId`

- Kind: choice
- Basis: Step 3.3. The schema requires `productId`, and gives no minimum length ([CDS: BankingProductV6][cds-schema-product]).
- Reason: An empty value cannot identify a product.

### <a id="dh-0263"></a>DH-0263 The maximum length of a product name

- Kind: choice
- Basis: Step 3.3. The CDS gives no maximum for the `name` of a product ([CDS: BankingProductV6][cds-schema-product]).
- Reason: An account that a consumer opens on the online bank site gets the name of its product ([Opening of an account](../DESIGN.md#opening-of-an-account)), and the `productName` of an account has 1 to 100 characters ([the length of a product name of an account](operator-interface.md#dh-0737)). With the rule in the load, each product can open, and an open cannot fail because of its name.

### <a id="dh-0264"></a>DH-0264 The effective period of a product

- Kind: choice
- Basis: Step 3.3. The CDS says only that the `effective` filter uses whether the current time is in the effective period ([CDS: Get Products][cds-get-products]).
- Reason: A start that is included and an end that is not included give no time at which a product is in two states. An absent date with no limit is the usual meaning of an open period.

### <a id="dh-0265"></a>DH-0265 Invented product data

- Kind: choice
- Basis: Step 3.1. The lab is not affiliated with a CDR participant ([README][readme]).
- Reason: A copy of the products of a real data holder can make a reader think that the lab is affiliated with it.

### <a id="dh-0266"></a>DH-0266 The products of the preset catalogues

- Kind: choice
- Basis: Step 3.3. The CDS gives the product categories and the `effective` filter ([CDS: Product Categories][cds-product-categories], [CDS: Get Products][cds-get-products]), and gives no rule for the products of a mock data holder.
- Reason: Five categories with the fields that the data holder holds give a realistic catalogue. A future product and a product with an end in the past let a client test each value of the `effective` filter.

### <a id="dh-0267"></a>DH-0267 A schema check of each preset catalogue

- Kind: choice
- Basis: Step 3.3. The conformance suite runs against the default preset brand only ([CDS OpenAPI: banking][cds-openapi-banking]).
- Reason: The other preset brands must also give valid responses. A check of each preset brand finds a catalogue that is not valid before the lab runs it.

### <a id="dh-0268"></a>DH-0268 No load into a data holder that has products

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for how a data holder gets its products.
- Reason: A merge of two sets can leave a `lastUpdated` that is not correct. A reset is simple in the lab.

### <a id="dh-0269"></a>DH-0269 No check of an `additionalValue` in the load

- Kind: choice
- Basis: Step 3.3. The CDS gives the rule of each type value in a table for each component, and not in the schema ([CDS: Product & Account Components][cds-product-components]).
- Reason: The load is a lab-only guard for the lab operator, and a check of each type value does not justify its cost. The preset catalogues obey the tables.

### <a id="dh-0270"></a>DH-0270 The CDS field names in a load file

- Kind: choice
- Basis: Step 3.3. The CDS gives no format for the products that a data holder loads.
- Reason: The lab operator then sees the same names in the file and in the API responses.

### <a id="dh-0271"></a>DH-0271 Realistic products in the lab

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the products of a mock data holder.
- Reason: A client can then test each filter of Get Products against the lab with realistic products.

### <a id="dh-0272"></a>DH-0272 An exact comparison for the `brand` filter

- Kind: choice
- Basis: Step 3.3. The CDS says only that the filter gives the results of a specific brand ([CDS: Get Products][cds-get-products]).
- Reason: A client gets the value from the `brand` field of a product. An exact comparison is then sufficient, and it is the simplest.

### <a id="dh-0273"></a>DH-0273 A stable order of the products

- Kind: choice
- Basis: Step 3.3. The CDS gives the order by `lastUpdated` only ([CDS: Get Products][cds-get-products]).
- Reason: Pagination needs a stable order. Without a second key, a product with the same `lastUpdated` as a different product can be on two pages, or on none.

### <a id="dh-0274"></a>DH-0274 The detail of each product that Get Products can list

- Kind: choice
- Basis: Step 3.3. The CDS does not say whether Get Product Detail returns a product that is not current ([CDS: Get Product Detail][cds-get-product-detail]).
- Reason: Get Products with `effective=ALL` or `FUTURE` lists such a product. A client must be able to read the detail of each product that the list gives.

### <a id="dh-0275"></a>DH-0275 No status other than `OK` and no outages

- Kind: choice
- Basis: Step 3.3. The CDS gives the statuses and the outages that a data holder reports ([CDS: Get Status][cds-get-status], [CDS: Get Outages][cds-get-outages]).
- Reason: The mock has no monitor of its own endpoints and no planned outages, and a data holder that responds is functional. An operator function to set a status or an outage can come when a client needs to test it.

[cds-get-products]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-products
[cds-get-product-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-product-detail
[cds-get-status]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-status
[cds-get-outages]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-outages
[cds-schema-product]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingproductv6
[cds-schema-product-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingproductdetailv7
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-id-permanence]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#id-permanence
[cds-product-categories]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-categories
[cds-product-components]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-amp-account-components
[cds-payload-conventions]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#payload-conventions
[cds-openapi-banking]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_banking.json
[support-brands]: https://cdr-support.zendesk.com/hc/en-us/articles/900005881263
[support-status-fields]: https://cdr-support.zendesk.com/hc/en-us/articles/900005173683
[readme]: ../../../README.md
[readme-component-types]: ../../../README.md#component-types
