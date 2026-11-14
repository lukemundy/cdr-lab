package app

// ParticipantSort is a sort key. Each participant list accepts only some of the keys.
type ParticipantSort string

const (
	SortByName                 ParticipantSort = "name"
	SortByAccreditationNumber  ParticipantSort = "accreditation-number"
	SortByStatus               ParticipantSort = "status"
	SortByBrandCount           ParticipantSort = "brand-count"
	SortBySoftwareProductCount ParticipantSort = "software-product-count"
)
