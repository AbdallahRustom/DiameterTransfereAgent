package radius

import (
	"context"
	"log"
	"net"
	"time"

	"diametertransfereagent/pkg/config"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const FramedProtocolGPRSPDPContext uint32 = 7

type RequestType int

const (
	AccessRequest RequestType = iota
)

type Request struct {
	Type             RequestType
	Username         string
	Password         string
	NASIPAddress     string
	NASPortType      rfc2865.NASPortType
	ServiceType      rfc2865.ServiceType
	CalledStationID  string
	CallingStationID string
}

type Response struct {
	Code      radius.Code
	FramedIP  net.IP
	FramedMTU uint32
}

type Client struct {
	cfg          *config.RadiusConfig
	requestChan  chan Request
	responseChan chan Response
}

func NewClient(cfg config.RadiusConfig, requestChan chan Request, responseChan chan Response) *Client {
	return &Client{cfg: &cfg, requestChan: requestChan, responseChan: responseChan}
}

func (c *Client) SendAccessRequest(ctx context.Context, req Request) error {

	packet := radius.New(radius.CodeAccessRequest, []byte(c.cfg.Secret))
	rfc2865.UserName_SetString(packet, req.Username)
	rfc2865.UserPassword_SetString(packet, req.Password)
	rfc2865.NASIPAddress_Set(packet, net.ParseIP(req.NASIPAddress))
	rfc2865.NASPortType_Set(packet, req.NASPortType)
	rfc2865.ServiceType_Set(packet, req.ServiceType)
	rfc2865.CalledStationID_SetString(packet, req.CalledStationID)
	rfc2865.CallingStationID_SetString(packet, req.CallingStationID)

	packet.Attributes.Add(rfc2865.FramedProtocol_Type, radius.NewInteger(FramedProtocolGPRSPDPContext))

	response, err := radius.Exchange(ctx, packet, c.cfg.Addr)
	if err != nil {
		log.Printf("Respone error: %v", err)
		return err
	}
	framedIP := rfc2865.FramedIPAddress_Get(response)
	framedMTU := uint32(rfc2865.FramedMTU_Get(response))

	// Send the response back through the response channel
	c.responseChan <- Response{
		Code:      response.Code,
		FramedIP:  framedIP,
		FramedMTU: framedMTU}
	return nil
}

func (c *Client) Start() {
	log.Println("Radius client started")

	for req := range c.requestChan {
		go func(req Request) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var err error

			switch req.Type {
			case AccessRequest:
				err = c.SendAccessRequest(ctx, req)
			}

			if err != nil {
				log.Println("Failed to send Radius access request:", err)
			}
		}(req)
	}
}
