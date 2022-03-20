package chunk

import (
	"context"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk/backoff"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"io"
)

// ChunkProcessFunc callback function for processing a chunk
type ChunkProcessFunc func(c *ChunkGroup, chunk io.Reader) error

// ChunkGroup manage groups of chunks
type ChunkGroup struct {
	file      fsctx.FileHeader
	chunkSize uint64
	backoff   backoff.Backoff

	fileInfo     *fsctx.UploadTaskInfo
	currentIndex int
	chunkNum     uint64
}

func NewChunkGroup(file fsctx.FileHeader, chunkSize uint64, backoff backoff.Backoff) *ChunkGroup {
	c := &ChunkGroup{
		file:         file,
		chunkSize:    chunkSize,
		backoff:      backoff,
		fileInfo:     file.Info(),
		currentIndex: -1,
	}

	if c.chunkSize == 0 {
		c.chunkSize = c.fileInfo.Size
	}

	c.chunkNum = c.fileInfo.Size / c.chunkSize
	if c.fileInfo.Size%c.chunkSize != 0 || c.fileInfo.Size == 0 {
		c.chunkNum++
	}

	return c
}

// Process a chunk with retry logic
func (c *ChunkGroup) Process(processor ChunkProcessFunc) error {
	err := processor(c, io.LimitReader(c.file, int64(c.chunkSize)))
	if err != nil {
		if err != context.Canceled && c.file.Seekable() && c.backoff.Next() {
			if _, seekErr := c.file.Seek(c.Start(), io.SeekStart); seekErr != nil {
				return fmt.Errorf("failed to seek back to chunk start: %w, last error: %w", seekErr, err)
			}

			util.Log().Debug("Retrying chunk %d, last error: %s", c.currentIndex, err)
			return c.Process(processor)
		}

		return err
	}

	util.Log().Debug("Chunk %d processed", c.currentIndex)
	return nil
}

// Start returns the byte index of current chunk
func (c *ChunkGroup) Start() int64 {
	return int64(uint64(c.Index()) * c.chunkSize)
}

// Total returns the total length current chunk
func (c *ChunkGroup) Total() int64 {
	return int64(c.fileInfo.Size)
}

// Num returns the total chunk number
func (c *ChunkGroup) Num() int {
	return int(c.chunkNum)
}

// RangeHeader returns header value of Content-Range
func (c *ChunkGroup) RangeHeader() string {
	return fmt.Sprintf("bytes %d-%d/%d", c.Start(), c.Start()+c.Length()-1, c.Total())
}

// Index returns current chunk index, starts from 0
func (c *ChunkGroup) Index() int {
	return c.currentIndex
}

// Next switch to next chunk, returns whether all chunks are processed
func (c *ChunkGroup) Next() bool {
	c.currentIndex++
	c.backoff.Reset()
	return c.currentIndex < int(c.chunkNum)
}

// Length returns the length of current chunk
func (c *ChunkGroup) Length() int64 {
	contentLength := c.chunkSize
	if c.Index() == int(c.chunkNum-1) {
		contentLength = c.fileInfo.Size - c.chunkSize*(c.chunkNum-1)
	}

	return int64(contentLength)
}

// IsLast returns if current chunk is the last one
func (c *ChunkGroup) IsLast() bool {
	return c.Index() == int(c.chunkNum-1)
}
