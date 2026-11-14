package domain

type SoftwareProductStatus string

const (
	SoftwareProductActive   SoftwareProductStatus = "ACTIVE"
	SoftwareProductInactive SoftwareProductStatus = "INACTIVE"
	SoftwareProductRemoved  SoftwareProductStatus = "REMOVED"
)

func ParseSoftwareProductStatus(s string) (SoftwareProductStatus, bool) {
	status := SoftwareProductStatus(s)

	switch status {
	case SoftwareProductActive, SoftwareProductInactive, SoftwareProductRemoved:
		return status, true
	default:
		return "", false
	}
}
