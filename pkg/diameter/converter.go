package diameter

import (
	reduisparams "diametertransfereagent/pkg/radius"
	"log"

	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"layeh.com/radius/rfc2865"
)

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

type DiameterRequest interface{}

func ConvertToRadius(messagetype string, m *diam.Message, c diam.Conn) (*reduisparams.Request, DiameterRequest) {

	var reduispacket reduisparams.Request

	switch messagetype {
	case diam.AIR:

		var req AuthenticationInformationRequest
		err := m.Unmarshal(&req)
		if err != nil {
			log.Printf("Failed to unmarshal AIR: %s", err)
			return nil, req
		}
		reduispacket.Type = 0
		reduispacket.Username = string(req.UserName)
		reduispacket.Password = "12345"
		reduispacket.NASIPAddress = c.RemoteAddr().String()
		reduispacket.NASPortType = rfc2865.NASPortType_Value_Virtual
		reduispacket.ServiceType = rfc2865.ServiceType_Value_FramedUser
		reduispacket.CalledStationID = "00-14-22-01-23-45"
		reduispacket.CallingStationID = "00-14-22-67-89-AB"
		return &reduispacket, req

	case diam.AAR:

		var req AuthenticationAuthorizationRequest
		err := m.Unmarshal(&req)
		if err != nil {
			log.Printf("Failed to unmarshal AAR: %s", err)
			return nil, req
		}
		reduispacket.Type = 0
		reduispacket.Username = string(req.UserName)
		reduispacket.Password = "12345"
		reduispacket.NASIPAddress = c.RemoteAddr().String()
		reduispacket.NASPortType = rfc2865.NASPortType_Value_Virtual
		reduispacket.ServiceType = rfc2865.ServiceType_Value_FramedUser
		reduispacket.CalledStationID = "00-14-22-01-23-45"
		reduispacket.CallingStationID = "00-14-22-67-89-AB"
		return &reduispacket, req

	case diam.DPR:

		var req DisconnectPeerRequest
		err := m.Unmarshal(&req)
		if err != nil {
			log.Printf("Failed to unmarshal DPR: %s", err)
			return nil, req
		}
		return nil, req
	}

	return nil, nil
}
