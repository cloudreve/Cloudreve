package chunk

import (
	"context"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk/backoff"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"io"
	"os"
)

const bufferTempPattern = "cdChunk.*.tmp"

// ChunkProcessFunc callback function for processing a chunk
type ChunkProcessFunc func(c *ChunkGroup, chunk io.Reader) error

// ChunkGroup manage groups of chunks
type ChunkGroup struct {
	file              fsctx.FileHeader
	chunkSize         uint64
	backoff           backoff.Backoff
	enableRetryBuffer bool

	fileInfo     *fsctx.UploadTaskInfo
	currentIndex int
	chunkNum     uint64
	bufferTemp   *os.File
}

func NewChunkGroup(file fsctx.FileHeader, chunkSize uint64, backoff backoff.Backoff, useBuffer bool) *ChunkGroup {
	c := &ChunkGroup{
		file:              file,
		chunkSize:         chunkSize,
		backoff:           backoff,
		fileInfo:          file.Info(),
		currentIndex:      -1,
		enableRetryBuffer: useBuffer,
	}

	if c.chunkSize == 0 {
		c.chunkSize = c.fileInfo.Size
	}

	if c.fileInfo.Size == 0 {
		c.chunkNum = 1
	} else {
		c.chunkNum = c.fileInfo.Size / c.chunkSize
		if c.fileInfo.Size%c.chunkSize != 0 {
			c.chunkNum++
		}
	}

	return c
}

// TempAvailable returns if current chunk temp file is available to be read
func (c *ChunkGroup) TempAvailable() bool {
	if c.bufferTemp != nil {
		state, _ := c.bufferTemp.Stat()
		return state != nil && state.Size() == c.Length()
	}

	return false
}

// Process a chunk with retry logic
func (c *ChunkGroup) Process(processor ChunkProcessFunc) error {
	reader := io.LimitReader(c.file, c.Length())

	// If useBuffer is enabled, tee the reader to a temp file
	if c.enableRetryBuffer && c.bufferTemp == nil && !c.file.Seekable() {
		c.bufferTemp, _ = os.CreateTemp("", bufferTempPattern)
		reader = io.TeeReader(reader, c.bufferTemp)
	}

	if c.bufferTemp != nil {
		defer func() {
			if c.bufferTemp != nil {
				c.bufferTemp.Close()
				os.Remove(c.bufferTemp.Name())
				c.bufferTemp = nil
			}
		}()

		// if temp buffer file is available, use it
		if c.TempAvailable() {
			if _, err := c.bufferTemp.Seek(0, io.SeekStart); err != nil {
				return fmt.Errorf("failed to seek temp file back to chunk start: %w", err)
			}

			util.Log().Debug("Chunk %d will be read from temp file %q.", c.Index(), c.bufferTemp.Name())
			reader = io.NopCloser(c.bufferTemp)
		}
	}

	err := processor(c, reader)
	if err != nil {
		if c.enableRetryBuffer {
			request.BlackHole(reader)
		}

		if err != context.Canceled && (c.file.Seekable() || c.TempAvailable()) && c.backoff.Next(err) {
			if c.file.Seekable() {
				if _, seekErr := c.file.Seek(c.Start(), io.SeekStart); seekErr != nil {
					return fmt.Errorf("failed to seek back to chunk start: %w, last error: %s", seekErr, err)
				}
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

// Total returns the total length
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
