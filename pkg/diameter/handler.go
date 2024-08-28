package diameter

import (
	"context"
	"diametertransfereagent/pkg/radius"
	"io"
	"log"
	"net"
	"time"

	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/sm"
	radiusres "layeh.com/radius"
)

func PrintErrors(ec <-chan *diam.ErrorReport) {
	for err := range ec {
		log.Println(err)
	}
}

func handleDiameterRequest(settings sm.Settings, requestChan chan radius.Request, responseChan chan radius.Response, messageType string, c diam.Conn, m *diam.Message) {
	log.Printf("Handling %s Request from %s", messageType, c.RemoteAddr())

	radiusMessageparams, req := ConvertToRadius(messageType, m, c)

	if radiusMessageparams == nil {
		switch messageType {
		case diam.AIR:
			a := BuildDiameterResponse(settings, req.(AuthenticationInformationRequest), diam.UnableToComply, nil, 0, m)
			_, _ = sendReply(c, a)
		case diam.AAR:
			a := BuildDiameterResponse(settings, req.(AuthenticationAuthorizationRequest), diam.UnableToComply, nil, 0, m)
			_, _ = sendReply(c, a)
		}
		return
	}

	// Send a request to the Radius client
	requestChan <- *radiusMessageparams

	// Wait for the response from the Radius client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case response := <-responseChan:
		var resultCode uint32
		var radiusIp net.IP
		var radiusMtu uint32
		if response.Code == radiusres.CodeAccessAccept && response.FramedIP != nil && response.FramedMTU != 0 {
			log.Printf("Received a successful response from Radius client: %v", response)
			resultCode = diam.Success
			radiusIp = response.FramedIP
			radiusMtu = response.FramedMTU
		} else {
			log.Printf("Received an unsuccessful response from Radius client: %v", response)
			resultCode = diam.AuthorizationRejected
		}
		switch messageType {
		case diam.AIR:
			a := BuildDiameterResponse(settings, req.(AuthenticationInformationRequest), resultCode, radiusIp, radiusMtu, m)
			_, _ = sendReply(c, a)
		case diam.AAR:
			a := BuildDiameterResponse(settings, req.(AuthenticationAuthorizationRequest), resultCode, radiusIp, radiusMtu, m)
			_, _ = sendReply(c, a)
		}

	case <-ctx.Done():
		log.Println("Timeout waiting for Radius response")
		// Send a reject response back to the Diameter client
		switch messageType {
		case diam.AIR:
			a := BuildDiameterResponse(settings, req.(AuthenticationInformationRequest), diam.AuthorizationRejected, nil, 0, m)
			_, _ = sendReply(c, a)
		case diam.AAR:
			a := BuildDiameterResponse(settings, req.(AuthenticationAuthorizationRequest), diam.AuthorizationRejected, nil, 0, m)
			_, _ = sendReply(c, a)
		}
	}
}

func HandleAuthenticationInformation(settings sm.Settings, requestChan chan radius.Request, responseChan chan radius.Response) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		go handleDiameterRequest(settings, requestChan, responseChan, diam.AIR, c, m)
	}
}

func HandleAuthorizationAuthenticationRequest(settings sm.Settings, requestChan chan radius.Request, responseChan chan radius.Response) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		go handleDiameterRequest(settings, requestChan, responseChan, diam.AAR, c, m)
	}
}

func sendReply(w io.Writer, m *diam.Message) (n int64, err error) {
	return m.WriteTo(w)
}

func HandleDisconnectPeerRequest(settings sm.Settings) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		go func() {
			log.Printf("Handling Disconnect-Peer-Request from %s", c.RemoteAddr())
			_, req := ConvertToRadius(diam.DPR, m, c)
			a := BuildDiameterResponse(settings, req.(DisconnectPeerRequest), diam.Success, nil, 0, m)
			_, _ = sendReply(c, a)
			c.Close()
		}()

	}
}

func HandleALL(c diam.Conn, m *diam.Message) {
	go func() {
		// Handle all other messages here
		log.Printf("Received unexpected message from %s:\n%s", c.RemoteAddr(), m)
	}()
}
