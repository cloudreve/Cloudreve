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
			if _, seekErr := c.file.Seek(c.Start(), io.SeekStart); err != nil {
				return fmt.Errorf("failed to seek back to chunk start: %w, last error: %w", seekErr, err)
			}

			util.Log().Debug("Retrying chunk %d, last error: %s", c.currentIndex, err)
			return c.Process(processor)
		}

		return err
	}

	return nil
}

// Start returns the byte index of current chunk
func (c *ChunkGroup) Start() int64 {
	return int64(uint64(c.Index()) * c.chunkSize)
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
