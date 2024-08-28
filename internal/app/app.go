package app

import (
	"diametertransfereagent/pkg/config"
	"diametertransfereagent/pkg/diameter"
	"diametertransfereagent/pkg/radius"
)

type App struct {
	diameterServer *diameter.Server
	radiusClient   *radius.Client
	requestChan    chan radius.Request
	responseChan   chan radius.Response
}

func NewApp(cfg *config.Config) *App {
	requestChan := make(chan radius.Request, 100) // Buffered channel
	responseChan := make(chan radius.Response, 100)
	radiusClient := radius.NewClient(cfg.RadiusConfig, requestChan, responseChan)
	diameterServer := diameter.NewServer(cfg.DiameterConfig, requestChan, responseChan)

	return &App{

		diameterServer: diameterServer,
		radiusClient:   radiusClient,
		requestChan:    requestChan,
		responseChan:   responseChan,
	}
}

func (a *App) Run() error {
	// Start handling messages
	go a.diameterServer.Start()
	go a.radiusClient.Start()

	select {}
}
