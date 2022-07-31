package api

import (
	"log"
	"net/http"
	"path"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type RealTimeStateSync struct {
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
	realTimeStateSyncer *realtime.StateSyncer,
) RealTimeStateSync {
	return RealTimeStateSync{
		realTimeStateSyncer: realTimeStateSyncer,
	}
}
