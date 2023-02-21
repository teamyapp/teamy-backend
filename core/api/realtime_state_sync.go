package api

import (
	"net/http"
	"path"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

const clientIDParam = "clientId"

type RealTimeStateSync struct {
	dataCollector       telemetry.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
}

var _ runner.Service = (*RealTimeStateSync)(nil)

func (r RealTimeStateSync) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(realTimeStateSyncPrefix, "clients", "connect"),
			HandlerFunc: r.connect,
		},
		{
			Method:      http.MethodPut,
			Pattern:     path.Join(realTimeStateSyncPrefix, "clients", runner.Param(clientIDParam), "initial-state-ready"),
			HandlerFunc: r.clientInitialStateReady,
		},
	})
	return nil
}

func (r RealTimeStateSync) clientInitialStateReady(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "userID not found",
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	clientIDRaw := chi.URLParam(request, clientIDParam)
	clientID, err := strconv.ParseUint(clientIDRaw, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "must provide teamId",
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	internalErr := r.realTimeStateSyncer.OnInitialStateReady(userID, clientID)
	if internalErr != nil {
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (r RealTimeStateSync) connect(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "userID not found",
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	conn, err := connection.WebSocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		internalErr := &errs.Error{
			Code: connection.ConnErr,
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	webSocketConn := connection.NewWebSocket(r.dataCollector, conn)
	internalErr := r.realTimeStateSyncer.OnClientConnect(userID, webSocketConn)
	if internalErr != nil {
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}
}

func NewRealTimeStateSync(
	dataCollector telemetry.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
) RealTimeStateSync {
	return RealTimeStateSync{
		dataCollector:       dataCollector,
		realTimeStateSyncer: realTimeStateSyncer,
	}
}
