package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/teamyapp/cloud/libs/ctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const requestIDMetadataKey = "T-Request-Id"

func WebWithRequestID(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		requestID := request.Header.Get(requestIDMetadataKey)
		if len(requestID) == 0 {
			// it's okay to have conflicts for request ID
			randomID := uuid.New()
			requestID = randomID.String()
			log.Printf("[Web] Generate request ID: requestID=%v\n", requestID)
		}

		ct := request.Context()
		ct = ctx.NewContextWithRequestID(ct, requestID)
		request = request.WithContext(ct)
		handlerFunc(writer, request)
	}
}

var GRPCWithRequestID grpc.UnaryServerInterceptor = func(
	ct context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ct)
	if !ok {
		md = metadata.New(map[string]string{})
	}

	requestIDs := md.Get(requestIDMetadataKey)
	if len(requestIDs) == 0 {
		// it's okay to have conflicts for request ID
		randomID := uuid.New()
		requestID := randomID.String()
		log.Printf("[GRPC] Generate request ID: requestID=%v\n", requestID)
		md.Set(requestIDMetadataKey, requestID)
	}

	ct = metadata.NewIncomingContext(ct, md)
	res, err := handler(ct, req)
	return res, err
}
