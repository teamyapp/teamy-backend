package middleware

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/libs/ctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var webRequestLogFieldOrder = []string{
	"requestId",
	"host",
	"method",
	"path",
	"headers",
	"bodySize",
	"body",
}

var gRPCRequestLogFieldOrder = []string{
	"requestId",
	"method",
	"request",
	"response",
}

func LogWebRequest(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return
		}

		requestLogFields := map[string]string{
			"host":     request.URL.Host,
			"method":   request.Method,
			"path":     request.URL.Path,
			"headers":  fmt.Sprintf("%v", request.Header),
			"bodySize": strconv.FormatInt(int64(len(buf)), 10),
		}
		responseLogFields := map[string]string{}

		requestID, err := ctx.RequestIDFromContext(request.Context())
		if err == nil {
			requestLogFields["requestId"] = requestID
			responseLogFields["requestId"] = requestID
		}

		if hasReadableBody(request.Header) {
			requestLogFields["body"] = string(buf)
			request.Body = ioutil.NopCloser(bytes.NewReader(buf))
		}

		log.Printf("[Web][Begin] %v\n", mapToString(requestLogFields, webRequestLogFieldOrder))
		loggableWriter := newLoggableResponseWriter(writer)

		// Process request
		handlerFunc(loggableWriter, request)

		responseLogFields["headers"] = fmt.Sprintf("%v", writer.Header())
		responseLogFields["bodySize"] = strconv.FormatInt(int64(len(loggableWriter.responseBody)), 10)
		if hasReadableBody(writer.Header()) {
			responseLogFields["body"] = string(loggableWriter.responseBody)
		}

		log.Printf("[Web][End] %v\n", mapToString(responseLogFields, webRequestLogFieldOrder))
	}
}

var LogGRPCRequest grpc.UnaryServerInterceptor = func(
	ct context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	requestLogFields := map[string]string{
		"method":  info.FullMethod,
		"request": fmt.Sprintf("%v", req),
	}
	responseLogFields := map[string]string{}

	md, ok := metadata.FromIncomingContext(ct)
	if ok {
		requestLogFields["metadata"] = fmt.Sprintf("%v", md)
		requestIDs := md.Get(requestIDMetadataKey)
		if len(requestIDs) > 0 {
			requestLogFields["requestId"] = requestIDs[0]
			responseLogFields["requestId"] = requestIDs[0]
		}
	}

	log.Printf("[GRPC][Begin] %v\n", mapToString(requestLogFields, gRPCRequestLogFieldOrder))

	// Process request
	res, err := handler(ct, req)

	responseLogFields["response"] = fmt.Sprintf("%v", res)
	log.Printf("[GRPC][End] %v\n", mapToString(responseLogFields, gRPCRequestLogFieldOrder))
	return res, err
}

type LoggableResponseWriter struct {
	http.ResponseWriter
	responseBody []byte
}

var _ http.ResponseWriter = (*LoggableResponseWriter)(nil)
var _ http.Hijacker = (*LoggableResponseWriter)(nil)
var _ http.Flusher = (*LoggableResponseWriter)(nil)

func (l *LoggableResponseWriter) Write(i []byte) (int, error) {
	l.responseBody = i
	return l.ResponseWriter.Write(i)
}

func (l *LoggableResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := l.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response does not implement http.Hijacker")
	}

	return hijacker.Hijack()
}

func (l *LoggableResponseWriter) Flush() {
	flusher, ok := l.ResponseWriter.(http.Flusher)
	if !ok {
		log.Println("response does not implement http.Flusher")
		return
	}

	flusher.Flush()
}

func newLoggableResponseWriter(writer http.ResponseWriter) *LoggableResponseWriter {
	return &LoggableResponseWriter{
		ResponseWriter: writer,
	}
}

func hasReadableBody(headers http.Header) bool {
	contentType := headers.Get("Content-Type")
	if len(contentType) == 0 {
		return false
	}

	if strings.HasPrefix(contentType, "text/") {
		return true
	}

	if strings.HasSuffix(contentType, "json") ||
		strings.HasSuffix(contentType, "xml") {
		return true
	}

	return false
}

func mapToString(mp map[string]string, fieldOrder []string) string {
	pairs := make([]string, 0)
	for _, field := range fieldOrder {
		value, ok := mp[field]
		if ok {
			pairs = append(pairs, field+"="+value)
		}
	}

	return strings.Join(pairs, " ")
}
