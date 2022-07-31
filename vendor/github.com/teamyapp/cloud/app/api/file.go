package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type File struct {
	fileService service.File
	proto.UnimplementedFileServer
}

var _ runner.Service = (*File)(nil)
var _ proto.FileServer = (*File)(nil)

func (f File) Start(rn *runner.ServiceRunner) error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(filePathPrefix, "upload-sessions", "{uploadSessionId}"),
			Method:      http.MethodGet,
			HandlerFunc: f.webGetUploadSession,
		},
		{
			Path:        path.Join(filePathPrefix, "upload-sessions", "{uploadSessionId}", "init"),
			Method:      http.MethodPut,
			HandlerFunc: f.webInitUploadSession,
		},
		{
			Path:        path.Join(filePathPrefix, "upload-sessions", "{uploadSessionId}", "delete"),
			Method:      http.MethodDelete,
			HandlerFunc: f.webDeleteUploadSession,
		},
		{
			Path:        path.Join(filePathPrefix, "upload-sessions", "{uploadSessionId}", "chunks", "add"),
			Method:      http.MethodPost,
			HandlerFunc: f.webAddChunk,
		},
		{
			Path:        path.Join(filePathPrefix, "files", "{fileId}", "metadata"),
			Method:      http.MethodGet,
			HandlerFunc: f.webGetFileMetadata,
		},
		{
			Path:        path.Join(filePathPrefix, "files", "{fileId}"),
			Method:      http.MethodGet,
			HandlerFunc: f.webGetFile,
		},
	})
	rn.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterFileServer(server, f)
	})

	return nil
}

func (f File) CreateUploadSession(ct context.Context, empty *emptypb.Empty) (*proto.CreateUploadSessionResponse, error) {
	uploadSessionID, err := f.fileService.CreateUploadSession(ct)
	if err != nil {
		return nil, err
	}

	return &proto.CreateUploadSessionResponse{
		UploadSessionId: uploadSessionID,
	}, nil
}

func (f File) FindUploadSession(ct context.Context, req *proto.FindUploadSessionRequest) (*proto.UploadSession, error) {
	uploadSession, err := f.fileService.GetUploadSession(ct, req.UploadSessionId)
	if err != nil {
		return &proto.UploadSession{}, err
	}

	return &proto.UploadSession{
		Id:                     uploadSession.ID,
		Status:                 toProtoUploadSessionStatus[uploadSession.Status],
		UploadedSizeInBytes:    uploadSession.UploadedSizeInBytes,
		FileId:                 uploadSession.FileID,
		FileName:               uploadSession.FileName,
		MimeType:               uploadSession.MIMEType,
		TotalSizeInBytes:       uploadSession.TotalSizeInBytes,
		TotalNumOfChunks:       int32(uploadSession.TotalNumOfChunks),
		ChunkIDs:               uploadSession.ChunkIDs,
		NextChunkIndexToUpload: int32(uploadSession.NextChunkIndexToUpload),
		ActualContentHash:      uploadSession.ActualContentHash,
		ExpectedContentHash:    uploadSession.ExpectedContentHash,
		CreatedAt:              timestamppb.New(uploadSession.CreatedAt),
		UpdatedAt:              toProtoTimePtr(uploadSession.UpdatedAt),
	}, nil
}

func (f File) webGetUploadSession(writer http.ResponseWriter, request *http.Request) {
	uploadSessionIDParam := mux.Vars(request)["uploadSessionId"]
	uploadSessionID, err := strconv.ParseUint(uploadSessionIDParam, 10, 64)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	uploadSession, err := f.fileService.GetUploadSession(request.Context(), uploadSessionID)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	web.WriteJSON(writer, uploadSession)
}

func (f File) webInitUploadSession(writer http.ResponseWriter, request *http.Request) {
	uploadSessionIDParam := mux.Vars(request)["uploadSessionId"]
	uploadSessionID, err := strconv.ParseUint(uploadSessionIDParam, 10, 64)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	buf, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	var body struct {
		FileName            string `json:"fileName"`
		MIMEType            string `json:"mimeType"`
		ExpectedContentHash string `json:"expectedContentHash"`
		TotalSizeInBytes    uint64 `json:"totalSizeInBytes"`
		TotalNumOfChunks    int    `json:"totalNumOfChunks"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	uploadSession, err := f.fileService.InitUploadSession(
		request.Context(),
		uploadSessionID,
		body.FileName,
		body.MIMEType,
		body.ExpectedContentHash,
		body.TotalSizeInBytes,
		body.TotalNumOfChunks)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	web.WriteJSON(writer, uploadSession)
}

func (f File) webDeleteUploadSession(writer http.ResponseWriter, request *http.Request) {
	panic("not implemented")
}

func (f File) webAddChunk(writer http.ResponseWriter, request *http.Request) {
	uploadSessionIDParam := mux.Vars(request)["uploadSessionId"]
	uploadSessionID, err := strconv.ParseUint(uploadSessionIDParam, 10, 64)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	uploadSession, err := f.fileService.AddChunk(request.Context(), uploadSessionID, data)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	web.WriteJSON(writer, uploadSession)
}

func (f File) webGetFileMetadata(writer http.ResponseWriter, request *http.Request) {
	fileIDParam := mux.Vars(request)["fileId"]
	fileID, err := strconv.ParseUint(fileIDParam, 10, 64)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	fileMetadata, err := f.fileService.GetFileMetadata(request.Context(), fileID)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	web.WriteJSON(writer, fileMetadata)
}

func (f File) webGetFile(writer http.ResponseWriter, request *http.Request) {
	fileIDParam := mux.Vars(request)["fileId"]
	fileID, err := strconv.ParseUint(fileIDParam, 10, 64)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	file, err := f.fileService.GetFile(request.Context(), fileID)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", file.Metadata.MIMEType)
	writer.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.Metadata.Name))
	if file.Metadata.LastModifiedAt != nil {
		writer.Header().Set("Last-Modified", file.Metadata.LastModifiedAt.UTC().Format(http.TimeFormat))
	}

	flusher, ok := writer.(http.Flusher)
	if !ok {
		log.Println("writer must be http.Flusher")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	for chunkResult := range file.ChunksBuffer {
		if chunkResult.Error != nil {
			log.Println(chunkResult.Error)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = writer.Write(chunkResult.Value)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		flusher.Flush()
	}
}

func NewFile(fileService service.File) File {
	return File{fileService: fileService}
}
