package service

import (
	"context"
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"log"
	"time"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/app/lang"
	"github.com/teamyapp/cloud/app/storage"
)

const chunksBufferSize = 10

type File struct {
	mapBackend         storage.MapBackend
	uploadSessionDao   dao.UploadSession
	fileMetadataDao    dao.FileMetadata
	chunkDao           dao.ChunkMetadata
	uploadSessionIDGen *gen.UniqueNumber
	chunkIDGen         *gen.UniqueNumber
	fileIDGen          *gen.UniqueNumber
}

func (f File) GetUploadSession(ct context.Context, uploadSessionID uint64) (entity.UploadSession, error) {
	return f.uploadSessionDao.FindUploadSessionByID(uploadSessionID)
}

func (f File) CreateUploadSession(ct context.Context) (uint64, error) {
	chunkID, err := f.uploadSessionIDGen.GenerateUniqueNumber()
	if err != nil {
		log.Println(err)
		return 0, err
	}

	uploadSession := entity.UploadSession{
		ID:        chunkID,
		Status:    entity.CreatedUploadSessionStatus,
		CreatedAt: time.Now(),
	}
	return chunkID, f.uploadSessionDao.CreateUploadSession(uploadSession)
}

func (f File) InitUploadSession(
	ct context.Context,
	uploadSessionID uint64,
	fileName string,
	mimeType string,
	expectedContentHash string,
	totalSizeInBytes uint64,
	totalNumOfChunks int,
) (entity.UploadSession, error) {
	uploadSession, err := f.uploadSessionDao.FindUploadSessionByID(uploadSessionID)
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	switch uploadSession.Status {
	case entity.CompletedUploadSessionStatus:
		return entity.UploadSession{}, errors.New("upload session is already completed")
	case entity.InitializedUploadSessionStatus, entity.UploadingChunksUploadSessionStatus:
		return entity.UploadSession{}, errors.New("upload session is already initialized")
	}

	hashBuffer := sha256.New()
	hashState, err := hashBuffer.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	uploadSession.FileName = fileName
	uploadSession.MIMEType = mimeType
	uploadSession.ExpectedContentHash = expectedContentHash
	uploadSession.TotalSizeInBytes = totalSizeInBytes
	uploadSession.TotalNumOfChunks = totalNumOfChunks
	uploadSession.Status = entity.InitializedUploadSessionStatus
	now := time.Now()
	uploadSession.HashState = hashState
	uploadSession.UpdatedAt = &now
	err = f.uploadSessionDao.UpdateUploadSession(uploadSession)
	uploadSession.HashState = nil
	return uploadSession, err
}

func (f File) AddChunk(ct context.Context, uploadSessionID uint64, chunkData []byte) (entity.UploadSession, error) {
	uploadSession, err := f.uploadSessionDao.FindUploadSessionByID(uploadSessionID)
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	switch uploadSession.Status {
	case entity.CompletedUploadSessionStatus:
		return entity.UploadSession{}, errors.New("upload session is already completed")
	case entity.CreatedUploadSessionStatus:
		return entity.UploadSession{}, errors.New("upload session it not initialized")
	}

	chunkID, err := f.chunkIDGen.GenerateUniqueNumber()
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	err = saveChunk(f.mapBackend, chunkID, chunkData)
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	now := time.Now()
	chunkMetadata := entity.ChunkMetadata{
		ID:          chunkID,
		SizeInBytes: uint64(len(chunkData)),
		CreatedAt:   now,
	}
	err = f.chunkDao.CreateChunkMetadata(chunkMetadata)
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	hashBuffer := sha256.New()
	err = hashBuffer.(encoding.BinaryUnmarshaler).UnmarshalBinary(uploadSession.HashState)
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	_, err = hashBuffer.Write(chunkData)
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	hashState, err := hashBuffer.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	uploadSession.HashState = hashState
	uploadSession.ChunkIDs = append(uploadSession.ChunkIDs, chunkID)
	uploadSession.NextChunkIndexToUpload++
	uploadSession.UploadedSizeInBytes += chunkMetadata.SizeInBytes
	uploadSession.UpdatedAt = &now
	if uploadSession.NextChunkIndexToUpload < uploadSession.TotalNumOfChunks {
		uploadSession.Status = entity.UploadingChunksUploadSessionStatus
	} else {
		uploadSession, err = f.FinishFileUpload(uploadSession, hashBuffer)
		if err != nil {
			log.Println(err)
			return entity.UploadSession{}, err
		}
	}

	err = f.uploadSessionDao.UpdateUploadSession(uploadSession)
	uploadSession.HashState = nil
	return uploadSession, err
}

func (f File) FinishFileUpload(uploadSession entity.UploadSession, hashBuffer hash.Hash) (entity.UploadSession, error) {
	uploadSession.Status = entity.CompletedUploadSessionStatus
	actualHash := hashBuffer.Sum(nil)
	actualHashString := hex.EncodeToString(actualHash)
	if actualHashString != uploadSession.ExpectedContentHash {
		err := fmt.Errorf("sha256 hash not match: actualHash=%v, expectedHash=%v", actualHashString, uploadSession.ExpectedContentHash)
		log.Println(err)
		return entity.UploadSession{}, err
	}

	uploadSession.ActualContentHash = actualHashString
	fileID, err := f.fileIDGen.GenerateUniqueNumber()
	if err != nil {
		log.Println(err)
		return entity.UploadSession{}, err
	}

	fileMetadata := entity.FileMetadata{
		ID:          fileID,
		Name:        uploadSession.FileName,
		SizeInBytes: uploadSession.TotalSizeInBytes,
		MIMEType:    uploadSession.MIMEType,
		ChunkIDs:    uploadSession.ChunkIDs,
		CreatedAt:   uploadSession.CreatedAt,
	}
	uploadSession.FileID = fileID
	return uploadSession, f.fileMetadataDao.CreateFileMetadata(fileMetadata)
}

func (f File) GetFileMetadata(ct context.Context, fileID uint64) (entity.FileMetadata, error) {
	return f.fileMetadataDao.FindMetadataByFileID(fileID)
}

func (f File) GetFile(ct context.Context, fileID uint64) (entity.File, error) {
	metadata, err := f.GetFileMetadata(ct, fileID)
	if err != nil {
		log.Println(err)
		return entity.File{}, err
	}

	chunksIterator := newChunksIterator(f.mapBackend, metadata.ChunkIDs)
	chunksBuffer := make(chan lang.Result[[]byte], chunksBufferSize)
	go func() {
		defer close(chunksBuffer)
		for {
			hasNext, err := chunksIterator.HasNext()
			if err != nil {
				log.Println(err)
				chunksBuffer <- lang.Result[[]byte]{
					Error: err,
				}
				return
			}

			if !hasNext {
				return
			}

			data, err := chunksIterator.Next()
			if err != nil {
				log.Println(err)
				chunksBuffer <- lang.Result[[]byte]{
					Error: err,
				}
				return
			}

			chunksBuffer <- lang.Result[[]byte]{
				Value: data,
				Error: nil,
			}

			// Required to prevent Wire gen from failure
			continue
		}
	}()
	return entity.File{
		Metadata:     metadata,
		ChunksBuffer: chunksBuffer,
	}, nil
}

func NewFile(
	mapBackend storage.MapBackend,
	uniqueNumberFactory gen.UniqueNumberFactory,
	uploadSessionDao dao.UploadSession,
	fileMetadataDao dao.FileMetadata,
	chunkDao dao.ChunkMetadata,
) (File, error) {
	uploadSessionIDGen, err := uniqueNumberFactory.MakeUniqueNumber("uploadSessionID")
	if err != nil {
		return File{}, err
	}

	chunkIDGen, err := uniqueNumberFactory.MakeUniqueNumber("chunkID")
	if err != nil {
		return File{}, err
	}

	fileIDGen, err := uniqueNumberFactory.MakeUniqueNumber("fileID")
	if err != nil {
		return File{}, err
	}

	return File{
		mapBackend:         mapBackend,
		uploadSessionDao:   uploadSessionDao,
		fileMetadataDao:    fileMetadataDao,
		chunkDao:           chunkDao,
		uploadSessionIDGen: uploadSessionIDGen,
		chunkIDGen:         chunkIDGen,
		fileIDGen:          fileIDGen,
	}, nil
}
