package diameter

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"diametertransfereagent/pkg/config"
	"diametertransfereagent/pkg/radius"

	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"github.com/fiorix/go-diameter/v4/diam/dict"
	"github.com/fiorix/go-diameter/v4/diam/sm"
)

const (
	VENDOR_3GPP           = 10415
	S6B_APP_ID            = 16777272
	defaultDictionaryPath = "./dictionary/"
)

type Server struct {
	cfg          *config.DiameterConfig
	requestChan  chan radius.Request
	responseChan chan radius.Response
}

func NewServer(cfg config.DiameterConfig, requestChan chan radius.Request, responseChan chan radius.Response) *Server {
	return &Server{cfg: &cfg, requestChan: requestChan, responseChan: responseChan}
}

func (s *Server) Start() {
	addr := flag.String("addr", s.cfg.Addr, "address in the form of ip:port to listen on")
	ppaddr := flag.String("pprof_addr", ":9000", "address in form of ip:port for the pprof server")
	host := flag.String("diam_host", s.cfg.DiamHost, "diameter identity host")
	realm := flag.String("diam_realm", s.cfg.DiamRealm, "diameter identity realm")
	certFile := flag.String("cert_file", s.cfg.CertFile, "tls certificate file (optional)")
	keyFile := flag.String("key_file", s.cfg.KeyFile, "tls key file (optional)")
	networkType := flag.String("network_type", s.cfg.NetworkType, "protocol type tcp/sctp")
	flag.Parse()

	settings := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(*host),
		OriginRealm:      datatype.DiameterIdentity(*realm),
		VendorID:         13,
		ProductName:      "go-diameter",
		FirmwareRevision: 1,
	}

	customDict, err := loadCustomDictionaries()
	if err != nil {
		log.Fatalf("Failed to load custom dictionaries: %v", err)
	}
	//changing default dictonary global variable
	dict.Default = customDict
	mux := sm.New(settings)

	mux.Handle("AIR", HandleAuthenticationInformation(*settings, s.requestChan, s.responseChan))
	mux.Handle("AAR", HandleAuthorizationAuthenticationRequest(*settings, s.requestChan, s.responseChan))
	mux.Handle("CCR", HandleCreditControlRequest(*settings, s.requestChan, s.responseChan))
	mux.Handle("DPR", HandleDisconnectPeerRequest(*settings))
	mux.HandleFunc("ALL", HandleALL)

	go PrintErrors(mux.ErrorReports())

	if len(*ppaddr) > 0 {
		go func() {
			srv := &http.Server{
				Addr:         *ppaddr,
				Handler:      nil,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  15 * time.Second,
			}
			log.Fatal(srv.ListenAndServe())
		}()
	}
	// Start listening for incoming connections
	go func() {
		if err := listen(*networkType, *addr, *certFile, *keyFile, mux); err != nil {
			log.Fatal(err)
		}
	}()

}

func listen(networkType, addr, cert, key string, handler diam.Handler) error {
	if len(cert) > 0 && len(key) > 0 {
		log.Println("Starting secure diameter server on", addr)
		return diam.ListenAndServeNetworkTLS(networkType, addr, cert, key, logHandler(handler), nil)
	}
	log.Println("Starting diameter server on", addr)
	return diam.ListenAndServeNetwork(networkType, addr, logHandler(handler), nil)
}

func logHandler(handler diam.Handler) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		if m.Header.CommandCode == diam.CapabilitiesExchange {
			log.Printf("New connection from %s", c.RemoteAddr())
		}
		handler.ServeDIAM(c, m)
	}
}

func loadCustomDictionaries() (*dict.Parser, error) {
	// List of dictionary files to load
	dictionaries := []struct{ name, file string }{
		{"Base", "base.xml"},
		{"TGPP_S6a", "s6a.xml"},
		{"TGPP_Swx", "swx.xml"},
		{"TGPP_S6b", "s6b.xml"},
		{"TGPP_GY", "gy.xml"},
		{"TGPP_3GPP", "3gpp.xml"},
	}

	// Create a new parser
	parser, err := dict.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create dictionary parser: %w", err)
	}

	// Load each dictionary file
	for _, dictEntry := range dictionaries {
		filePath := defaultDictionaryPath + dictEntry.file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read dictionary file %s: %w", filePath, err)
		}

		if err := parser.Load(bytes.NewReader(data)); err != nil {
			return nil, fmt.Errorf("failed to load dictionary from file %s: %w", filePath, err)
		}

		log.Printf("Successfully loaded dictionary from file %s", filePath)
	}

	return parser, nil
}
