package service

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/teamyapp/teamy-backend/infras/connection"
	"github.com/teamyapp/teamy-backend/infras/runner"
	"github.com/teamyapp/teamy-backend/infras/storage"
	"github.com/teamyapp/teamy-backend/infras/web"
)

const realTimeCollectionsPrefix = "/real-time-collections"

type Message[Payload any] struct {
	Type    string
	Payload Payload
}

type ClientIDPayload struct {
	ClientID uint64
}

func newClientIDMessage(clientID uint64) Message[ClientIDPayload] {
	return Message[ClientIDPayload]{
		Type:    "CLIENT_ID",
		Payload: ClientIDPayload{clientID},
	}
}

type RealTimeCollectionsConfig struct {
	WebAPIPort int
}

type RealTimeCollections struct {
	webSocketUpgrader   websocket.Upgrader
	realTimeCollections *storage.RealTimeCollections
}

var _ runner.Service = (*RealTimeCollections)(nil)

func (r RealTimeCollections) Start(runner *runner.ServiceRunner) error {
	runner.RegisterWebRoutes([]web.Route{
		{
			Path:        path.Join(realTimeCollectionsPrefix, "clients", "connect"),
			Method:      http.MethodGet,
			HandlerFunc: r.connect,
		},
		{
			Path:        path.Join(realTimeCollectionsPrefix, "clients", "{clientId}", "subscriptions", "subscribe"),
			Method:      http.MethodPost,
			HandlerFunc: r.subscribe,
		},
		{
			Path:        path.Join(realTimeCollectionsPrefix, "subscriptions"),
			Method:      http.MethodGet,
			HandlerFunc: r.listSubscriptions,
		},
		{
			Path:        path.Join(realTimeCollectionsPrefix, "subscriptions", "{subscriptionId}", "unsubscribe"),
			Method:      http.MethodPost,
			HandlerFunc: r.unsubscribe,
		},
		{
			Path:        path.Join(realTimeCollectionsPrefix, "collection-types/register"),
			Method:      http.MethodPost,
			HandlerFunc: r.registerResourceType,
		},
		{
			Path:        path.Join(realTimeCollectionsPrefix, "collection-types"),
			Method:      http.MethodGet,
			HandlerFunc: r.listResourceTypes,
		},
		{
			Path:        path.Join(realTimeCollectionsPrefix, "mutate"),
			Method:      http.MethodPost,
			HandlerFunc: r.mutate,
		},
	})

	return nil
}

func (r RealTimeCollections) connect(writer http.ResponseWriter, request *http.Request) {
	conn, err := r.webSocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	webSocketConn := connection.NewWebSocket(conn)
	clientID := r.realTimeCollections.OnClientConnect(webSocketConn)
	log.Printf("client connected: %v\n", clientID)
	message, err := json.MarshalIndent(newClientIDMessage(clientID), "", "  ")
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	webSocketConn.SendMessage(message)
	writer.WriteHeader(http.StatusNoContent)
}

func (r RealTimeCollections) registerResourceType(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	r.realTimeCollections.RegisterCollectionType(string(body))
	writer.WriteHeader(http.StatusNoContent)
}

func (r RealTimeCollections) listResourceTypes(writer http.ResponseWriter, request *http.Request) {
	collectionTypes := r.realTimeCollections.ListCollectionTypes()
	buf, err := json.MarshalIndent(collectionTypes, "", "  ")
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	web.WriteJSON(writer, buf)
}

func (r RealTimeCollections) subscribe(writer http.ResponseWriter, request *http.Request) {
	// TODO: verify user access token
	clientIDParam := mux.Vars(request)["clientId"]
	clientID, err := strconv.ParseUint(clientIDParam, 10, 64)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	var subscribeRequest struct {
		CollectionType    string
		MutationType      storage.MutationType
		AttributeMatchers []storage.AttributeMatcher
	}

	err = json.Unmarshal(body, &subscribeRequest)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	subscription := storage.Subscription{
		CollectionType:    subscribeRequest.CollectionType,
		MutationType:      subscribeRequest.MutationType,
		AttributeMatchers: subscribeRequest.AttributeMatchers,
		ClientID:          clientID,
	}

	if err = subscription.Validate(); err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}

	subscriptionID, err := r.realTimeCollections.Subscribe(subscription)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.Write([]byte(strconv.FormatUint(subscriptionID, 10)))
}

func (r RealTimeCollections) listSubscriptions(writer http.ResponseWriter, request *http.Request) {
	buf, err := json.MarshalIndent(r.realTimeCollections.ListSubscriptions(), "", "  ")
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	web.WriteJSON(writer, buf)
}

func (r RealTimeCollections) unsubscribe(writer http.ResponseWriter, request *http.Request) {
	subscriptionIDParam := mux.Vars(request)["subscriptionId"]
	subscriptionID, err := strconv.ParseUint(subscriptionIDParam, 10, 64)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	err = r.realTimeCollections.Unsubscribe(subscriptionID)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (r RealTimeCollections) mutate(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	var mutation storage.Mutation
	err = json.Unmarshal(body, &mutation)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	if err = mutation.Validate(); err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}

	err = r.realTimeCollections.Mutate(mutation)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (r RealTimeCollections) InMemoryRealTimeCollections() *storage.RealTimeCollections {
	return r.realTimeCollections
}

func NewRealTimeCollections() RealTimeCollections {
	return RealTimeCollections{
		realTimeCollections: storage.NewRealTimeCollections(),
	}
}
