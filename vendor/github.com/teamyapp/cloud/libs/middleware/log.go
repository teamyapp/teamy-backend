package middleware

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LogWebRequest(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return
		}

		if hasReadableBody(request.Header) {
			log.Printf("[Web][Begin] host=%v method=%v path=%v headers=%v bodySize=%v body=%v\n",
				request.URL.Host,
				request.Method,
				request.URL.Path,
				request.Header,
				len(buf),
				string(buf))
			request.Body = ioutil.NopCloser(bytes.NewReader(buf))
		} else {
			log.Printf("[Web][Begin] host=%v method=%v path=%v headers=%v bodySize=%v\n",
				request.URL.Host,
				request.Method,
				request.URL.Path,
				request.Header,
				len(buf))
		}

		loggableWriter := newLoggableResponseWriter(writer)
		handlerFunc(loggableWriter, request)
		if hasReadableBody(writer.Header()) {
			log.Printf("[Web][End] headers=%v bodySize=%v body=%v\n",
				writer.Header(),
				len(loggableWriter.responseBody),
				string(loggableWriter.responseBody))
		} else {
			log.Printf("[Web][End] headers=%v bodySize=%v\n",
				writer.Header(),
				len(loggableWriter.responseBody))
		}
	}
}

var LogGRPCRequest grpc.UnaryServerInterceptor = func(
	ct context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ct)
	if ok {
		log.Printf("[GRPC][Begin] method=%v metadata=%v request=%v\n",
			info.FullMethod,
			md,
			req)
	} else {
		log.Printf("[GRPC][Begin] method=%v request=%v\n",
			info.FullMethod,
			req)
	}

	res, err := handler(ct, req)
	log.Printf("[GRPC][End] response=%v\n", res)
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
