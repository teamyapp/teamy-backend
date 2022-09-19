package api

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/teamy-backend/core/realtime"
)


type RealTimeStateSync struct {
	dataCollector       obs.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
}

var _ runner.Service = (*RealTimeStateSync)(nil)

func (r RealTimeStateSync) Start(rn *runner.ServiceRunner) error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(realTimeStateSyncPrefix, "clients", "connect"),
			Method:      http.MethodGet,
			HandlerFunc: r.connect,
		},
		{
			Path:        path.Join(realTimeStateSyncPrefix, "clients", "{clientID}","initial-state-ready"),
			Method:      http.MethodPut,
			HandlerFunc: r.clientInitialStateReady,
		},
	})
	return nil
}

func (r RealTimeStateSync) ready(writer http.ResponseWriter, request *http.Request) {
	userID, err := ctx.UserIDFromContext(r.dataCollector, request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(request.Body)
	readyParams := ReadyParams{}
	err = decoder.Decode(&readyParams)
	if err != nil {
		r.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp:   err,
			obs.MessageProp: "must provide clientId",
		})
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	clientID := readyParams.ClientID
	err = r.realTimeStateSyncer.OnInitialStateReady(userID, clientID)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (r RealTimeStateSync) connect(writer http.ResponseWriter, request *http.Request) {
	userID, err := ctx.UserIDFromContext(r.dataCollector, request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := connection.WebSocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		r.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	webSocketConn := connection.NewWebSocket(r.dataCollector, conn)
	err = r.realTimeStateSyncer.OnClientConnect(userID, webSocketConn)
	if err != nil {
		r.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func NewRealTimeStateSync(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
) RealTimeStateSync {
	return RealTimeStateSync{
		dataCollector:       dataCollector,
		realTimeStateSyncer: realTimeStateSyncer,
	}
}
