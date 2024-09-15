package models

import "github.com/fiorix/go-diameter/v4/diam/datatype"

type AuthenticationInformationRequest struct {
	SessionID                   datatype.UTF8String         `avp:"Session-Id"`
	OriginHost                  datatype.DiameterIdentity   `avp:"Origin-Host"`
	OriginRealm                 datatype.DiameterIdentity   `avp:"Origin-Realm"`
	DestinationRealm            datatype.DiameterIdentity   `avp:"Destination-Realm"`
	VendorSpecificApplicationID VendorSpecificApplicationID `avp:"Vendor-Specific-Application-Id"`
	AuthSessionState            datatype.Unsigned32         `avp:"Auth-Session-State"`
	UserName                    datatype.UTF8String         `avp:"User-Name"`
	VisitedPLMNID               datatype.OctetString        `avp:"Visited-PLMN-Id"`
	RequestedEUTRANAuthInfo     RequestedEUTRANAuthInfo     `avp:"Requested-EUTRAN-Authentication-Info"`
}

type VendorSpecificApplicationID struct {
	AuthApplicationID datatype.Unsigned32 `avp:"Auth-Application-Id"`
	VendorID          datatype.Unsigned32 `avp:"Vendor-Id"`
}

type RequestedEUTRANAuthInfo struct {
	NumVectors        datatype.Unsigned32  `avp:"Number-Of-Requested-Vectors"`
	ImmediateResponse datatype.Unsigned32  `avp:"Immediate-Response-Preferred"`
	ResyncInfo        datatype.OctetString `avp:"Re-synchronization-Info"`
}

type AuthenticationAuthorizationRequest struct {
	SessionID                datatype.UTF8String       `avp:"Session-Id"`
	OriginHost               datatype.DiameterIdentity `avp:"Origin-Host"`
	OriginRealm              datatype.DiameterIdentity `avp:"Origin-Realm"`
	DestinationRealm         datatype.DiameterIdentity `avp:"Destination-Realm"`
	AuthApplicationID        datatype.Unsigned32       `avp:"Auth-Application-Id"`
	AuthRequestType          datatype.Unsigned32       `avp:"Auth-Request-Type"`
	RatType                  datatype.Unsigned32       `avp:"RAT-Type"`
	UserName                 datatype.UTF8String       `avp:"User-Name"`
	VisitedNetworkIdentifier datatype.Unsigned32       `avp:"Visited-Network-Identifier"`
	ServiceSelection         datatype.UTF8String       `avp:"Service-Selection"`
}

type DisconnectPeerRequest struct {
	OriginHost       datatype.DiameterIdentity `avp:"Origin-Host"`
	OriginRealm      datatype.DiameterIdentity `avp:"Origin-Realm"`
	DestinationRealm datatype.DiameterIdentity `avp:"Destination-Realm"`
	DisconnectCause  datatype.Unsigned32       `avp:"Disconnect-Cause"`
}

type CreditControlRequest struct {
	SessionID                    datatype.UTF8String           `avp:"Session-Id"`
	OriginHost                   datatype.DiameterIdentity     `avp:"Origin-Host"`
	OriginRealm                  datatype.DiameterIdentity     `avp:"Origin-Realm"`
	DestinationRealm             datatype.DiameterIdentity     `avp:"Destination-Realm"`
	SubscriptionId               SubscriptionIdinfo            `avp:"Subscription-Id"`
	CCRequestType                datatype.Enumerated           `avp:"CC-Request-Type"`
	CCRequestNumber              datatype.Unsigned32           `avp:"CC-Request-Number"`
	ServiceInformation           Serviceinformationinfo        `avp:"Service-Information"`
	MultipleServiceCreditControl MultipleServicesCreditControl `avp:"Multiple-Services-Credit-Control"`
	AuthApplicationID            datatype.Unsigned32           `avp:"Auth-Application-Id"`
}

type SubscriptionIdinfo struct {
	SubscriptionIDData datatype.UTF8String `avp:"Subscription-Id-Data"`
}
type Serviceinformationinfo struct {
	PsInformation Psinformationinfo `avp:"PS-Information"`
}

type Psinformationinfo struct {
	PDPAddress               []datatype.Address   `avp:"PDP-Address"`
	CalledStationId          datatype.UTF8String  `avp:"Called-Station-Id"`
	UserEquipment            Userequipmentinfo    `avp:"User-Equipment-Info"`
	PDPType                  datatype.Enumerated  `avp:"TGPP-PDP-Type"`
	SGSNAddress              datatype.Address     `avp:"SGSN-Address"`
	GGSNAddress              datatype.Address     `avp:"GGSN-Address"`
	TGPPSGSNMCCMNC           datatype.UTF8String  `avp:"TGPP-SGSN-MCC-MNC"`
	ThreeGPPUserLocationInfo datatype.OctetString `avp:"TGPP-User-Location-Info"`
	ThreeGPPMSTimeZone       datatype.OctetString `avp:"TGPP-MS-TimeZone"`
	EventTimestamp           datatype.Time        `avp:"Event-Timestamp"`
}

type Userequipmentinfo struct {
	UserEquipmentInfoValue datatype.OctetString `avp:"User-Equipment-Info-Value"`
}

type MultipleServicesCreditControl struct {
	RequestedServiceUnit ServiceUnit          `avp:"Requested-Service-Unit"`
	UsedServiceUnit      ServiceUnit          `avp:"Used-Service-Unit"`
	Qos                  Oosinformation       `avp:"QoS-Information"`
	TGPPRATType          datatype.OctetString `avp:"TGPP-RAT-Type"`
}

type ServiceUnit struct {
	CCTime         datatype.Unsigned32 `avp:"CC-Time"`
	CCInputOctets  datatype.Unsigned64 `avp:"CC-Input-Octets"`
	CCOutputOctets datatype.Unsigned64 `avp:"CC-Output-Octets"`
}

type Oosinformation struct {
	APNAggregateMaxBitrateDL    datatype.Unsigned32         `avp:"APN-Aggregate-Max-Bitrate-DL"`
	APNAggregateMaxBitrateUL    datatype.Unsigned32         `avp:"APN-Aggregate-Max-Bitrate-UL"`
	QoSClassIdentifier          datatype.Enumerated         `avp:"QoS-Class-Identifier"`
	AllocationRetentionPriority AllocationRetentionPriority `avp:"Allocation-Retention-Priority"`
}
type AllocationRetentionPriority struct {
	PriorityLevel           datatype.Unsigned32 `avp:"Priority-Level"`
	PreEmptionCapability    datatype.Enumerated `avp:"Pre-emption-Capability"`
	PreEmptionVulnerability datatype.Enumerated `avp:"Pre-emption-Vulnerability"`
}
type DiameterRequest interface{}
