package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type UploadSession struct {
	db *sql.DB
}

var _ dao.UploadSession = (*UploadSession)(nil)

func (u UploadSession) FindUploadSessionByID(uploadSessionID uint64) (entity.UploadSession, error) {
	uploadSession := entity.UploadSession{}
	var chunkIDsString string
	err := u.db.QueryRow(`
	SELECT
	    id,
	    status,
	    file_id,
	    file_name,
	    mime_type,
	    chunk_ids,
	    uploaded_size_in_bytes,
	    total_size_in_bytes,
	    total_num_of_chunks,
	    next_chunk_index_to_upload,
	    hash_state,
	    actual_content_hash,
	    expected_content_hash,
	    created_at,
	    updated_at
	FROM file_upload_session
	WHERE id = $1;`,
		uploadSessionID).
		Scan(
			&uploadSession.ID,
			&uploadSession.Status,
			&uploadSession.FileID,
			&uploadSession.FileName,
			&uploadSession.MIMEType,
			&chunkIDsString,
			&uploadSession.UploadedSizeInBytes,
			&uploadSession.TotalSizeInBytes,
			&uploadSession.TotalNumOfChunks,
			&uploadSession.NextChunkIndexToUpload,
			&uploadSession.HashState,
			&uploadSession.ActualContentHash,
			&uploadSession.ExpectedContentHash,
			&uploadSession.CreatedAt,
			&uploadSession.UpdatedAt,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.UploadSession{}, dao.ErrNotFound(fmt.Sprintf(
			"upload session not found: id=%v", uploadSessionID))
	}

	if err != nil {
		return entity.UploadSession{}, err
	}

	chunkIDs, err := parseIDs(chunkIDsString)
	if err != nil {
		return entity.UploadSession{}, err
	}

	uploadSession.ChunkIDs = chunkIDs
	return uploadSession, err
}

func (u UploadSession) CreateUploadSession(uploadSession entity.UploadSession) error {
	_, err := u.db.Exec(`
	INSERT INTO file_upload_session
	(
	 	id,
	 	status,
	 	file_id,
	 	file_name,
	 	mime_type,
	 	chunk_ids,
	 	uploaded_size_in_bytes,
	 	total_size_in_bytes,
	 	total_num_of_chunks,
	 	next_chunk_index_to_upload,
	 	hash_state,
	 	actual_content_hash,
	 	expected_content_hash,
	 	created_at,
	 	updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);`,
		uploadSession.ID,
		uploadSession.Status,
		uploadSession.FileID,
		uploadSession.FileName,
		uploadSession.MIMEType,
		formatIDs(uploadSession.ChunkIDs),
		uploadSession.UploadedSizeInBytes,
		uploadSession.TotalSizeInBytes,
		uploadSession.TotalNumOfChunks,
		uploadSession.NextChunkIndexToUpload,
		uploadSession.HashState,
		uploadSession.ActualContentHash,
		uploadSession.ExpectedContentHash,
		uploadSession.CreatedAt,
		uploadSession.UpdatedAt,
	)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (u UploadSession) UpdateUploadSession(uploadSession entity.UploadSession) error {
	_, err := u.db.Exec(`
	UPDATE file_upload_session
	SET
	    id = $1,
	    status = $2,
	    file_id = $3,
	    file_name = $4,
	    mime_type = $5,
	    chunk_ids = $6,
	    uploaded_size_in_bytes = $7,
	    total_size_in_bytes = $8,
	    total_num_of_chunks = $9,
	    next_chunk_index_to_upload = $10,
	    hash_state = $11,
	    actual_content_hash = $12,
	    expected_content_hash = $13,
	    created_at = $14,
	    updated_at = $15
	WHERE id = $16;
	`,
		uploadSession.ID,
		uploadSession.Status,
		uploadSession.FileID,
		uploadSession.FileName,
		uploadSession.MIMEType,
		formatIDs(uploadSession.ChunkIDs),
		uploadSession.UploadedSizeInBytes,
		uploadSession.TotalSizeInBytes,
		uploadSession.TotalNumOfChunks,
		uploadSession.NextChunkIndexToUpload,
		uploadSession.HashState,
		uploadSession.ActualContentHash,
		uploadSession.ExpectedContentHash,
		uploadSession.CreatedAt,
		uploadSession.UpdatedAt,
		uploadSession.ID,
	)
	return err
}

func NewUploadSession(sqlDB *sql.DB) UploadSession {
	return UploadSession{
		db: sqlDB,
	}
}
