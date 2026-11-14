package app

import (
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var DataHolderType = eventsourcing.AggregateType[
	domain.DataHolderState, domain.DataHolderCommand, domain.DataHolderEvent,
]{
	StreamType: "data-holder",
	Handle:     domain.HandleDataHolder,
	Apply:      domain.ApplyDataHolder,
	Codec:      dataHolderCodec(),
}

var DataRecipientType = eventsourcing.AggregateType[
	domain.DataRecipientState, domain.DataRecipientCommand, domain.DataRecipientEvent,
]{
	StreamType: "data-recipient",
	Handle:     domain.HandleDataRecipient,
	Apply:      domain.ApplyDataRecipient,
	Codec:      dataRecipientCodec(),
}

func dataHolderCodec() *eventsourcing.JSONCodec[domain.DataHolderEvent] {
	c := eventsourcing.NewJSONCodec[domain.DataHolderEvent]()
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.DataHolderImported](c, "DataHolderImported")
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.DataHolderCreated](c, "DataHolderCreated")
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.DataHolderDetailsChanged](c, "DataHolderDetailsChanged")
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.DataHolderLegalEntityRemoved](
		c,
		"DataHolderLegalEntityRemoved",
	)
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.DataHolderBrandAdded](c, "DataHolderBrandAdded")
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.DataHolderBrandDetailsChanged](
		c,
		"DataHolderBrandDetailsChanged",
	)
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.DataHolderBrandActivated](c, "DataHolderBrandActivated")
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.DataHolderBrandDeactivated](
		c,
		"DataHolderBrandDeactivated",
	)
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.DataHolderBrandRemoved](c, "DataHolderBrandRemoved")
	eventsourcing.RegisterJSON[domain.DataHolderEvent, domain.ServerCertificateIssued](c, "ServerCertificateIssued")

	return c
}

func dataRecipientCodec() *eventsourcing.JSONCodec[domain.DataRecipientEvent] {
	c := eventsourcing.NewJSONCodec[domain.DataRecipientEvent]()
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.DataRecipientImported](c, "DataRecipientImported")
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.AccreditationSuspended](c, "AccreditationSuspended")
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.AccreditationRevoked](c, "AccreditationRevoked")
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.AccreditationSurrendered](
		c,
		"AccreditationSurrendered",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.AccreditationReinstated](c, "AccreditationReinstated")
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.SoftwareProductActivated](
		c,
		"SoftwareProductActivated",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.SoftwareProductDeactivated](
		c,
		"SoftwareProductDeactivated",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.SoftwareProductRemoved](c, "SoftwareProductRemoved")
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.ClientCertificateRegistered](
		c,
		"ClientCertificateRegistered",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.ClientCertificateRemoved](
		c,
		"ClientCertificateRemoved",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.ClientCertificateIssued](
		c,
		"ClientCertificateIssued",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.DataRecipientCreated](
		c,
		"DataRecipientCreated",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.DataRecipientDetailsChanged](
		c,
		"DataRecipientDetailsChanged",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.DataRecipientBrandAdded](
		c,
		"DataRecipientBrandAdded",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.DataRecipientBrandDetailsChanged](
		c,
		"DataRecipientBrandDetailsChanged",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.DataRecipientBrandActivated](
		c,
		"DataRecipientBrandActivated",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.DataRecipientBrandDeactivated](
		c,
		"DataRecipientBrandDeactivated",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.DataRecipientBrandRemoved](
		c,
		"DataRecipientBrandRemoved",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.SoftwareProductAdded](
		c,
		"SoftwareProductAdded",
	)
	eventsourcing.RegisterJSON[domain.DataRecipientEvent, domain.SoftwareProductDetailsChanged](
		c,
		"SoftwareProductDetailsChanged",
	)

	return c
}
