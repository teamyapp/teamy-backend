package service

import (
	"fmt"
	"log"
	"path"
	"strconv"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/storage"
)

const chunkKeyPrefix = "chunks"

type ChunksIterator struct {
	mapBackend     storage.MapBackend
	chunkIDs       []uint64
	nextChunkIndex int
}

var _ entity.Iterator[[]byte] = (*ChunksIterator)(nil)

func (c ChunksIterator) HasNext() (bool, error) {
	return c.nextChunkIndex < len(c.chunkIDs), nil
}

func (c *ChunksIterator) Next() ([]byte, error) {
	hasNext, err := c.HasNext()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !hasNext {
		return nil, fmt.Errorf("no next chunk: nextChunkIndex=%v. numOfChunks=%v", c.nextChunkIndex, c.chunkIDs)
	}

	chunkIDPath := strconv.FormatUint(c.chunkIDs[c.nextChunkIndex], 10)
	fullPath := path.Join(chunkKeyPrefix, chunkIDPath)
	data, err := c.mapBackend.Get(fullPath)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	c.nextChunkIndex++
	return data, nil
}

func newChunksIterator(mapBackend storage.MapBackend, chunkIDs []uint64) *ChunksIterator {
	return &ChunksIterator{
		mapBackend:     mapBackend,
		chunkIDs:       chunkIDs,
		nextChunkIndex: 0,
	}
}

func saveChunk(mapBackend storage.MapBackend, chunkID uint64, data []byte) error {
	chunkIDPath := strconv.FormatUint(chunkID, 10)
	fullPath := path.Join(chunkKeyPrefix, chunkIDPath)
	return mapBackend.Put(fullPath, data)
}
