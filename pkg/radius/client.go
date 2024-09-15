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
	PDPType          uint8
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
	UsedInputOctets  uint32
	UsedOutputOctets uint32
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
	rfc2865.UserName_SetString(packet, req.Username)
	rfc2865.UserPassword_SetString(packet, req.Password)
	rfc2865.NASIPAddress_Set(packet, net.ParseIP(req.NASIPAddress))
	rfc2865.NASPortType_Set(packet, req.NASPortType)
	rfc2865.ServiceType_Set(packet, req.ServiceType)
	rfc2865.CalledStationID_SetString(packet, req.CalledStationID)
	rfc2865.CallingStationID_SetString(packet, req.CallingStationID)

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
	rfc2865.UserName_SetString(packet, req.Username)
	rfc2866.AcctStatusType_Set(packet, req.AcctStatus)
	if req.PDPType == 0 {
		rfc2865.FramedIPAddress_Set(packet, req.Ipv4FramedIP)
	} else {
		rfc2865.FramedIPAddress_Set(packet, req.Ipv4FramedIP)
		rfc2865.FramedIPAddress_Set(packet, req.Ipv6FramedIP)
	}
	rfc2865.CalledStationID_SetString(packet, req.CalledStationID)
	rfc2866.AcctSessionID_Set(packet, []byte(req.AcctSessionID))

	switch req.AcctStatus {

	case rfc2866.AcctStatusType_Value_Start:
		rfc2866.AcctDelayTime_Set(packet, req.AcctDelayTime)

	case rfc2866.AcctStatusType_Value_InterimUpdate:
		rfc2866.AcctInputOctets_Set(packet, rfc2866.AcctInputOctets(req.UsedInputOctets))
		rfc2866.AcctOutputOctets_Set(packet, rfc2866.AcctOutputOctets(req.UsedOutputOctets))
		rfc2866.AcctInputPackets_Set(packet, 0)
		rfc2866.AcctOutputPackets_Set(packet, 0)
		rfc2866.AcctSessionTime_Set(packet, rfc2866.AcctSessionTime(req.Acctsessiontime))
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
