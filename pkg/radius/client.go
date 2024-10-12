package radius

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"diametertransfereagent/pkg/config"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

const FramedProtocolGPRSPDPContext uint32 = 7

type RequestType int

const (
	AccessRequest     RequestType = 0
	AccountingRequest RequestType = 1
)

type AuthRequest struct {
	Type             RequestType
	Username         string
	Password         string
	NASIPAddress     string
	NASPortType      rfc2865.NASPortType
	ServiceType      rfc2865.ServiceType
	CalledStationID  string
	CallingStationID string
}
type AuthResponse struct {
	Code      radius.Code
	FramedIP  net.IP
	FramedMTU uint32
}

type AccRequest struct {
	Type             RequestType
	Username         string
	Ipv4FramedIP     net.IP
	Ipv6FramedIP     net.IP
	AcctStatus       rfc2866.AcctStatusType
	CalledStationID  string
	AcctDelayTime    rfc2866.AcctDelayTime
	AcctSessionID    string
	IMSI             string
	PDPType          int32
	ULAMBR           string
	DLAMBR           string
	SGSNAddress      net.IP
	GGSNAddress      net.IP
	MCCMNC           string
	IMEISV           string
	RATType          string
	UserLocationInfo string
	Timezone         string
	EventTimestamp   string
	UsedInputOctets  uint64
	UsedOutputOctets uint64
	Acctsessiontime  uint32
}
type AccResponse struct {
	Code radius.Code
}
type Request interface {
	GetType() RequestType
}

type Response interface {
	GetCode() radius.Code
}

func (ar AuthRequest) GetType() RequestType {
	return ar.Type
}
func (ar AccRequest) GetType() RequestType {
	return ar.Type
}
func (ar AuthResponse) GetCode() radius.Code {
	return ar.Code
}
func (ar AccResponse) GetCode() radius.Code {
	return ar.Code
}

type Client struct {
	cfg          *config.RadiusConfig
	requestChan  chan Request // Changed to interface type
	responseChan chan Response
}

func (c *Client) Start() {
	log.Println("Radius client started")

	for req := range c.requestChan {
		go func(req Request) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var err error
			switch req.GetType() {
			case AccessRequest:
				if authReq, ok := req.(*AuthRequest); ok {
					err = c.SendAccessRequest(ctx, *authReq)
				} else {
					log.Println("Failed to assert req to AuthRequest")
				}
			case AccountingRequest:
				if accReq, ok := req.(*AccRequest); ok {
					err = c.SendAcctRequest(ctx, *accReq)
				} else {
					log.Println("Failed to assert req to AccRequest")
				}
			}

			if err != nil {
				log.Println("Failed to send Radius access request:", err)
			}

		}(req)
	}
	log.Println("started")

}

func NewClient(cfg config.RadiusConfig, requestChan chan Request, responseChan chan Response) *Client {
	return &Client{cfg: &cfg, requestChan: requestChan, responseChan: responseChan}
}

func (c *Client) SendAccessRequest(ctx context.Context, req AuthRequest) error {

	packet := radius.New(radius.CodeAccessRequest, []byte(c.cfg.Secret))
	if err := rfc2865.UserName_SetString(packet, req.Username); err != nil {
		log.Printf("Error Setting Username: %v", err)
		return err
	}

	if err := rfc2865.UserPassword_SetString(packet, req.Password); err != nil {
		log.Printf("Error Setting Password: %v", err)
		return err
	}

	if err := rfc2865.NASIPAddress_Set(packet, net.ParseIP(req.NASIPAddress)); err != nil {
		log.Printf("Error Setting NASIPAddress: %v", err)
		return err
	}

	if err := rfc2865.NASPortType_Set(packet, req.NASPortType); err != nil {
		log.Printf("Error Setting NASPortType: %v", err)
		return err
	}

	if err := rfc2865.ServiceType_Set(packet, req.ServiceType); err != nil {
		log.Printf("Error Setting ServiceType: %v", err)
		return err
	}

	if err := rfc2865.CalledStationID_SetString(packet, req.CalledStationID); err != nil {
		log.Printf("Error Setting CalledStationID: %v", err)
		return err
	}

	if err := rfc2865.CallingStationID_SetString(packet, req.CallingStationID); err != nil {
		log.Printf("Error Setting CallingStationID: %v", err)
		return err
	}

	packet.Attributes.Add(rfc2865.FramedProtocol_Type, radius.NewInteger(FramedProtocolGPRSPDPContext))

	authaddress := c.cfg.Addr + ":" + "1812"

	response, err := radius.Exchange(ctx, packet, authaddress)
	if err != nil {
		log.Printf("Respone error: %v", err)
		return err
	}
	framedIP := rfc2865.FramedIPAddress_Get(response)
	framedMTU := uint32(rfc2865.FramedMTU_Get(response))

	// Send the response back through the response channel
	c.responseChan <- AuthResponse{
		Code:      response.Code,
		FramedIP:  framedIP,
		FramedMTU: framedMTU,
	}
	return nil
}

func (c *Client) SendAcctRequest(ctx context.Context, req AccRequest) error {

	packet := radius.New(radius.CodeAccountingRequest, []byte(c.cfg.Secret))

	if err := rfc2865.UserName_SetString(packet, req.Username); err != nil {
		log.Printf("Error Setting UserName: %v", err)
		return err
	}

	if err := rfc2866.AcctStatusType_Set(packet, req.AcctStatus); err != nil {
		log.Printf("Error Setting AcctStatusType: %v", err)
		return err
	}

	if req.PDPType == 0 {
		if err := rfc2865.FramedIPAddress_Set(packet, req.Ipv4FramedIP); err != nil {
			log.Printf("Error Setting FramedIPAddress: %v", err)
			return err
		}
	} else {
		if err := rfc2865.FramedIPAddress_Set(packet, req.Ipv4FramedIP); err != nil {
			log.Printf("Error Setting FramedIPAddress: %v", err)
			return err
		}
		if err := rfc2865.FramedIPAddress_Set(packet, req.Ipv6FramedIP); err != nil {
			log.Printf("Error Setting FramedIPAddressipv6: %v", err)
			return err
		}
	}

	if err := rfc2865.CalledStationID_SetString(packet, req.CalledStationID); err != nil {
		log.Printf("Error Setting CalledStationID: %v", err)
		return err
	}

	if err := rfc2866.AcctSessionID_Set(packet, []byte(req.AcctSessionID)); err != nil {
		log.Printf("Error Setting AcctSessionID: %v", err)
		return err
	}

	switch req.AcctStatus {

	case rfc2866.AcctStatusType_Value_Start:

		if err := rfc2866.AcctDelayTime_Set(packet, req.AcctDelayTime); err != nil {
			log.Printf("Error Setting AcctDelayTime: %v", err)
			return err
		}

	case rfc2866.AcctStatusType_Value_InterimUpdate:

		if req.UsedInputOctets <= uint64(^uint32(0)) {
			if err := rfc2866.AcctInputOctets_Set(packet, rfc2866.AcctInputOctets(req.UsedInputOctets)); err != nil {
				log.Printf("Error Setting AcctInputOctets: %v", err)
				return err
			}
		} else {
			log.Printf("UsedInputOctets value too large: %d", req.UsedInputOctets)
		}

		if req.UsedOutputOctets <= uint64(^uint32(0)) {
			if err := rfc2866.AcctOutputOctets_Set(packet, rfc2866.AcctOutputOctets(req.UsedOutputOctets)); err != nil {
				log.Printf("Error Setting AcctOutputOctets: %v", err)
				return err
			}
		} else {
			log.Printf("UsedOutputOctets value too large: %d", req.UsedOutputOctets)
		}

		if err := rfc2866.AcctInputPackets_Set(packet, 0); err != nil {
			log.Printf("Error Setting AcctInputPackets: %v", err)
			return err
		}

		if err := rfc2866.AcctOutputPackets_Set(packet, 0); err != nil {
			log.Printf("Error Setting AcctOutputPackets: %v", err)
			return err
		}

		if err := rfc2866.AcctSessionTime_Set(packet, rfc2866.AcctSessionTime(req.Acctsessiontime)); err != nil {
			log.Printf("Error Setting AcctSessionTime: %v", err)
			return err
		}
	}

	// Helper function to add Vendor-Specific AVPs
	vendorID := uint32(10415) // 3GPP Vendor ID

	addVendorSpecific := func(vendorID uint32, vendorType byte, value radius.Attribute) error {
		vsa, _ := radius.NewVendorSpecific(vendorID, append([]byte{vendorType, byte(len(value) + 2)}, value...))
		packet.Attributes.Add(rfc2865.VendorSpecific_Type, vsa)
		return nil
	}
	// Add 3GPP specific AVPs as Vendor-Specific Attributes (VSA)
	if err := addVendorSpecific(vendorID, 1, []byte(req.IMSI)); err != nil { // IMSI
		return fmt.Errorf("failed to add IMSI: %v", err)
	}

	if err := addVendorSpecific(vendorID, 3, []byte(string(req.PDPType))); err != nil { // SGSN Address
		return fmt.Errorf("failed to add PDP type: %v", err)
	}

	// if err := addVendorSpecific(vendorID, 2, []byte(req.QosInformation)); err != nil { // SGSN Address
	// 	return fmt.Errorf("failed to add SGSN Address: %v", err)
	// }

	if err := addVendorSpecific(vendorID, 6, []byte(req.SGSNAddress)); err != nil { // GGSN Address
		return fmt.Errorf("failed to add GGSN Address: %v", err)
	}

	if err := addVendorSpecific(vendorID, 7, []byte(req.GGSNAddress)); err != nil { // GGSN Address
		return fmt.Errorf("failed to add GGSN Address: %v", err)
	}

	if err := addVendorSpecific(vendorID, 18, []byte(req.MCCMNC)); err != nil { // MCC-MNC
		return fmt.Errorf("failed to add MCC-MNC: %v", err)
	}

	if err := addVendorSpecific(vendorID, 20, []byte(req.IMEISV)); err != nil { // IMEISV
		return fmt.Errorf("failed to add IMEISV: %v", err)
	}

	if err := addVendorSpecific(vendorID, 21, []byte(req.RATType)); err != nil { // RAT-Type
		return fmt.Errorf("failed to add RAT-Type: %v", err)
	}

	if err := addVendorSpecific(vendorID, 22, []byte(req.UserLocationInfo)); err != nil { // User Location Info
		return fmt.Errorf("failed to add User Location Info: %v", err)
	}

	if err := addVendorSpecific(vendorID, 23, []byte(req.Timezone)); err != nil { // MS Timezone
		return fmt.Errorf("failed to add MS Timezone: %v", err)
	}

	// if err := addVendorSpecific(vendorID, 55, []byte(req.EventTimestamp)); err != nil { // Event Timestamp
	// 	return fmt.Errorf("failed to add Event Timestamp: %v", err)
	// }
	accaddress := c.cfg.Addr + ":" + "1813"
	response, err := radius.Exchange(ctx, packet, accaddress)
	if err != nil {
		log.Printf("Respone error: %v", err)
		return err
	}

	log.Printf("Respone: %v", response.Code)
	c.responseChan <- AccResponse{
		Code: response.Code,
	}
	return nil

}
