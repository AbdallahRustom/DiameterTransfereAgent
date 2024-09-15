package diameter

import (
	"diametertransfereagent/pkg/models"
	"diametertransfereagent/pkg/radius"
	"log"
	"net"
	"strings"

	"github.com/fiorix/go-diameter/v4/diam"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

func ConvertToRadius(messagetype string, m *diam.Message, c diam.Conn) (radius.Request, models.DiameterRequest) {

	// var radiuspacket radius.Request

	switch messagetype {
	case diam.AIR:
		var radiuspacket radius.AuthRequest
		var req models.AuthenticationInformationRequest
		err := m.Unmarshal(&req)
		if err != nil {
			log.Printf("Failed to unmarshal AIR: %s", err)
			return nil, req
		}
		radiuspacket.Type = radius.AccessRequest
		usernameparts := strings.Split(string(req.UserName), "@")
		username := usernameparts[0]
		radiuspacket.Username = string(username)
		radiuspacket.Password = "12345"
		radiuspacket.NASIPAddress = c.RemoteAddr().String()
		radiuspacket.NASPortType = rfc2865.NASPortType_Value_Virtual
		radiuspacket.ServiceType = rfc2865.ServiceType_Value_FramedUser
		radiuspacket.CalledStationID = "00-14-22-01-23-45"
		radiuspacket.CallingStationID = "00-14-22-67-89-AB"
		return &radiuspacket, req

	case diam.AAR:
		var radiuspacket radius.AuthRequest
		var req models.AuthenticationAuthorizationRequest
		err := m.Unmarshal(&req)
		if err != nil {
			log.Printf("Failed to unmarshal AAR: %s", err)
			return nil, req
		}
		radiuspacket.Type = radius.AccessRequest
		usernameparts := strings.Split(string(req.UserName), "@")
		username := usernameparts[0]
		radiuspacket.Username = string(username)
		radiuspacket.Password = "12345"
		radiuspacket.NASIPAddress = c.RemoteAddr().String()
		radiuspacket.NASPortType = rfc2865.NASPortType_Value_Virtual
		radiuspacket.ServiceType = rfc2865.ServiceType_Value_FramedUser
		radiuspacket.CalledStationID = "00-14-22-01-23-45"
		radiuspacket.CallingStationID = "00-14-22-67-89-AB"
		return &radiuspacket, req

	case diam.CCR:
		var req models.CreditControlRequest
		var radiuspacket radius.AccRequest
		err := m.Unmarshal(&req)
		if err != nil {
			log.Printf("Failed to unmarshal CCR: %s", err)
			return nil, req
		}

		radiuspacket.Type = radius.AccountingRequest
		radiuspacket.Username = string(req.SubscriptionId.SubscriptionIDData)

		if req.CCRequestType == 1 {
			radiuspacket.AcctStatus = rfc2866.AcctStatusType_Value_Start
		} else {
			//to be implemented
		}

		radiuspacket.PDPType = uint8(req.ServiceInformation.PsInformation.PDPType)

		if req.ServiceInformation.PsInformation.PDPType == 0 {
			radiuspacket.Ipv4FramedIP = net.IP(req.ServiceInformation.PsInformation.PDPAddress[0])
			// radiuspacket.Ipv6FramedIP = nil
		} else {
			radiuspacket.Ipv4FramedIP = net.IP(req.ServiceInformation.PsInformation.PDPAddress[0])
			radiuspacket.Ipv6FramedIP = net.IP(req.ServiceInformation.PsInformation.PDPAddress[1])
		}

		radiuspacket.CalledStationID = string(req.ServiceInformation.PsInformation.CalledStationId)
		radiuspacket.AcctDelayTime = rfc2866.AcctDelayTime(0)

		parts := strings.Split(string(req.SessionID), ";")
		if len(parts) > 2 {
			session_id := parts[len(parts)-2]
			radiuspacket.AcctSessionID = string(session_id)
		}
		// radiuspacket.AcctSessionID = string(req.SessionID)
		radiuspacket.IMSI = string(req.SubscriptionId.SubscriptionIDData)
		radiuspacket.ULAMBR = string(req.MultipleServiceCreditControl.Qos.APNAggregateMaxBitrateUL)
		radiuspacket.DLAMBR = string(req.MultipleServiceCreditControl.Qos.APNAggregateMaxBitrateDL)
		radiuspacket.SGSNAddress = net.IP(req.ServiceInformation.PsInformation.SGSNAddress)
		radiuspacket.GGSNAddress = net.IP(req.ServiceInformation.PsInformation.GGSNAddress)
		radiuspacket.MCCMNC = string(req.ServiceInformation.PsInformation.TGPPSGSNMCCMNC)
		radiuspacket.IMEISV = string(req.ServiceInformation.PsInformation.UserEquipment.UserEquipmentInfoValue)
		radiuspacket.RATType = string(req.MultipleServiceCreditControl.TGPPRATType)
		radiuspacket.UserLocationInfo = string(req.ServiceInformation.PsInformation.ThreeGPPUserLocationInfo)
		radiuspacket.Timezone = string(req.ServiceInformation.PsInformation.ThreeGPPMSTimeZone)
		radiuspacket.EventTimestamp = req.ServiceInformation.PsInformation.EventTimestamp.String()

		return &radiuspacket, req

	case diam.DPR:

		var req models.DisconnectPeerRequest
		err := m.Unmarshal(&req)
		if err != nil {
			log.Printf("Failed to unmarshal DPR: %s", err)
			return nil, &req
		}
		return nil, req
	}

	return nil, nil
}
