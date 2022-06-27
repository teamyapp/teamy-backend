package api

import (
	"log"
	"net/http"
	"path"

	"github.com/teamyapp/cloud/app/ctx"
	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/web"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

const realTimeStateSyncPrefix = "/real-time-state-sync"

type RealTimeStateSync struct {
	identityAPIEndpoint string
	realTimeStateSyncer *realtime.StateSyncer
}

var _ runner.Service = (*RealTimeStateSync)(nil)

func (r RealTimeStateSync) Start(runner *runner.ServiceRunner) error {
	runner.RegisterWebRoutes([]web.Route{
		{
			Path:        path.Join(realTimeStateSyncPrefix, "clients", "connect"),
			Method:      http.MethodGet,
			HandlerFunc: middleware.WithWebSocketIdentity(r.identityAPIEndpoint, r.connect).ServeHTTP,
		},
	})
	return nil
}

func (r RealTimeStateSync) connect(writer http.ResponseWriter, request *http.Request) {
	userID, err := ctx.UserIDFromContext(request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := connection.WebSocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	webSocketConn := connection.NewWebSocket(conn)
	err = r.realTimeStateSyncer.OnClientConnect(userID, webSocketConn)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func NewRealTimeStateSync(
	identityAPIEndpoint string,
	realTimeStateSyncer *realtime.StateSyncer,
) RealTimeStateSync {
	return RealTimeStateSync{
		identityAPIEndpoint: identityAPIEndpoint,
		realTimeStateSyncer: realTimeStateSyncer,
	}
}
