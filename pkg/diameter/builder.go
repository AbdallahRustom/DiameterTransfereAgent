package diameter

import (
	"diametertransfereagent/pkg/models"
	"net"

	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/avp"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"github.com/fiorix/go-diameter/v4/diam/sm"
)

// BuildDiameterResponse constructs a Diameter response message
func BuildDiameterResponse(settings sm.Settings, req models.DiameterRequest, resultCode uint32, radiusIp net.IP, radiusMtu uint32, m *diam.Message) *diam.Message {
	a := m.Answer(resultCode)

	switch r := req.(type) {
	case models.AuthenticationInformationRequest:
		// SessionID is required to be the AVP in position 1
		a.InsertAVP(diam.NewAVP(avp.SessionID, avp.Mbit, 0, r.SessionID))
		a.NewAVP(avp.OriginHost, avp.Mbit, 0, settings.OriginHost)
		a.NewAVP(avp.OriginRealm, avp.Mbit, 0, settings.OriginRealm)
		a.NewAVP(avp.VendorSpecificApplicationID, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, r.VendorSpecificApplicationID.AuthApplicationID),
				diam.NewAVP(avp.VendorID, avp.Mbit, 0, r.VendorSpecificApplicationID.VendorID),
			},
		})
		a.NewAVP(avp.AuthSessionState, avp.Mbit, 0, r.AuthSessionState)

		a.NewAVP(avp.AuthenticationInfo, avp.Mbit, VENDOR_3GPP, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.EUTRANVector, avp.Mbit, VENDOR_3GPP, &diam.GroupedAVP{
					AVP: []*diam.AVP{
						diam.NewAVP(avp.RAND, avp.Mbit|avp.Vbit, VENDOR_3GPP, datatype.OctetString("\x94\xbf/T\xc3v\xf3\x0e\x87\x83\x06k'\x18Z\x19")),
						diam.NewAVP(avp.XRES, avp.Mbit|avp.Vbit, VENDOR_3GPP, datatype.OctetString("F\xf0\"\xb9%#\xf58")),
						diam.NewAVP(avp.AUTN, avp.Mbit|avp.Vbit, VENDOR_3GPP, datatype.OctetString("\xc7G!;\xad~\x80\x00)\x08o%\x11\x0cP_")),
						diam.NewAVP(avp.KASME, avp.Mbit|avp.Vbit, VENDOR_3GPP, datatype.OctetString("\xbf\x00\xf9\x80h3\"\x0e\xa1\x1c\xfa\x93\x03@\xd6\xf8\x02\xd51Y\xeb\xc4\x9d=\t\x14{\xeb!\xec\xcb:")),
					},
				}),
			},
		})
		if radiusIp != nil {
			a.NewAVP(avp.FramedIPAddress, avp.Mbit, 0, datatype.OctetString(radiusIp.To4()))
		}
		if radiusMtu != 0 {
			a.NewAVP(avp.FramedMTU, avp.Mbit, 0, datatype.Unsigned32(radiusMtu))
		}

	case models.AuthenticationAuthorizationRequest:
		// SessionID is required to be the AVP in position 1
		a.InsertAVP(diam.NewAVP(avp.SessionID, avp.Mbit, 0, r.SessionID))
		a.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, r.AuthApplicationID)
		a.NewAVP(avp.AuthRequestType, avp.Mbit, 0, r.AuthRequestType)
		a.NewAVP(avp.SessionTimeout, avp.Mbit, 0, datatype.Unsigned32(7200))
		a.NewAVP(avp.OriginHost, avp.Mbit, 0, settings.OriginHost)
		a.NewAVP(avp.OriginRealm, avp.Mbit, 0, settings.OriginRealm)
		if radiusIp != nil {
			a.NewAVP(avp.FramedIPAddress, avp.Mbit, 0, datatype.OctetString(radiusIp.To4()))
		}
		if radiusMtu != 0 {
			a.NewAVP(avp.FramedMTU, avp.Mbit, 0, datatype.Unsigned32(radiusMtu))
		}

	case models.CreditControlRequest:

		a.InsertAVP(diam.NewAVP(avp.SessionID, avp.Mbit, 0, r.SessionID))
		a.NewAVP(avp.OriginHost, avp.Mbit, 0, settings.OriginHost)
		a.NewAVP(avp.OriginRealm, avp.Mbit, 0, settings.OriginRealm)
		a.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, r.AuthApplicationID)
		a.NewAVP(avp.CCRequestType, avp.Mbit, 0, r.CCRequestType)
		a.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, r.CCRequestNumber)

		if r.CCRequestType == models.CCRequestTypeInitial || r.CCRequestType == models.CCRequestTypeUpdate {
			a.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{
				AVP: []*diam.AVP{
					diam.NewAVP(avp.GrantedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(5)),
							diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(1024)),
							diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(1024)),
						},
					}),
				},
			})
		}

		// a.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{
		// 	AVP: []*diam.AVP{
		// 		diam.NewAVP(avp.TGPPRATType, avp.Mbit|avp.Vbit, VENDOR_3GPP, r.MultipleServiceCreditControl.TGPPRATType),
		// 		diam.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		// 			AVP: []*diam.AVP{
		// 				diam.NewAVP(avp.CCTime, avp.Mbit, 0, r.MultipleServiceCreditControl.RequestedServiceUnit.CCTime),
		// 				diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, r.MultipleServiceCreditControl.RequestedServiceUnit.CCOutputOctets),
		// 				diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, r.MultipleServiceCreditControl.RequestedServiceUnit.CCInputOctets),
		// 			},
		// 		}),
		// 		diam.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		// 			AVP: []*diam.AVP{
		// 				diam.NewAVP(avp.CCTime, avp.Mbit, 0, r.MultipleServiceCreditControl.UsedServiceUnit.CCTime),
		// 				diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, r.MultipleServiceCreditControl.UsedServiceUnit.CCOutputOctets),
		// 				diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, r.MultipleServiceCreditControl.UsedServiceUnit.CCInputOctets),
		// 			},
		// 		}),
		// 		diam.NewAVP(avp.QoSInformation, avp.Mbit|avp.Vbit, VENDOR_3GPP, &diam.GroupedAVP{
		// 			AVP: []*diam.AVP{
		// 				diam.NewAVP(avp.QoSClassIdentifier, avp.Mbit|avp.Vbit, VENDOR_3GPP, r.MultipleServiceCreditControl.Qos.QoSClassIdentifier),
		// 				diam.NewAVP(avp.APNAggregateMaxBitrateDL, avp.Vbit, VENDOR_3GPP, r.MultipleServiceCreditControl.Qos.APNAggregateMaxBitrateDL),
		// 				diam.NewAVP(avp.APNAggregateMaxBitrateUL, avp.Vbit, VENDOR_3GPP, r.MultipleServiceCreditControl.Qos.APNAggregateMaxBitrateUL),
		// 				diam.NewAVP(avp.AllocationRetentionPriority, avp.Vbit, VENDOR_3GPP, &diam.GroupedAVP{
		// 					AVP: []*diam.AVP{
		// 						diam.NewAVP(avp.PriorityLevel, avp.Vbit, VENDOR_3GPP, r.MultipleServiceCreditControl.Qos.AllocationRetentionPriority.PriorityLevel),
		// 						diam.NewAVP(avp.PreemptionCapability, avp.Vbit, VENDOR_3GPP, r.MultipleServiceCreditControl.Qos.AllocationRetentionPriority.PreEmptionCapability),
		// 						diam.NewAVP(avp.PreemptionVulnerability, avp.Vbit, VENDOR_3GPP, r.MultipleServiceCreditControl.Qos.AllocationRetentionPriority.PreEmptionVulnerability),
		// 					},
		// 				}),
		// 			},
		// 		}),
		// 	},
		// })

	case models.DisconnectPeerRequest:
		a.NewAVP(avp.OriginHost, avp.Mbit, 0, settings.OriginHost)
		a.NewAVP(avp.OriginRealm, avp.Mbit, 0, settings.OriginRealm)
	}

	return a
}
