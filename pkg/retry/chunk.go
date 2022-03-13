package retry

import (
	"context"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"io"
)

type ChunkProcessFunc func(index int, chunk io.Reader) error

func Chunk(index int, chunkSize uint64, file fsctx.FileHeader, processor ChunkProcessFunc, backoff Backoff) error {
	err := processor(index, file)
	if err != nil {
		if err != context.Canceled && file.Seekable() && backoff.Next() {
			if _, seekErr := file.Seek(int64(uint64(index)*chunkSize), io.SeekStart); err != nil {
				return fmt.Errorf("failed to seek back to chunk start: %w, last error: %w", seekErr, err)
			}

			util.Log().Debug("Retrying chunk %d, last error: %s", index, err)
			return Chunk(index, chunkSize, file, processor, backoff)
		}

		return err
	}

	return nil
}
