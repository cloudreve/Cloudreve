package chunk

import (
	"errors"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk/backoff"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNewChunkGroup(t *testing.T) {
	a := assert.New(t)

	testCases := []struct {
		fileSize               uint64
		chunkSize              uint64
		expectedInnerChunkSize uint64
		expectedChunkNum       uint64
		expectedInfo           [][2]int //Start, Index,Length
	}{
		{10, 0, 10, 1, [][2]int{{0, 10}}},
		{0, 0, 0, 1, [][2]int{{0, 0}}},
		{0, 10, 10, 1, [][2]int{{0, 0}}},
		{50, 10, 10, 5, [][2]int{
			{0, 10},
			{10, 10},
			{20, 10},
			{30, 10},
			{40, 10},
		}},
		{50, 50, 50, 1, [][2]int{
			{0, 50},
		}},

		{50, 15, 15, 4, [][2]int{
			{0, 15},
			{15, 15},
			{30, 15},
			{45, 5},
		}},
	}

	for index, testCase := range testCases {
		file := &fsctx.FileStream{Size: testCase.fileSize}
		chunkGroup := NewChunkGroup(file, testCase.chunkSize, &backoff.ConstantBackoff{}, true)
		a.EqualValues(testCase.expectedChunkNum, chunkGroup.Num(),
			"TestCase:%d,ChunkNum()", index)
		a.EqualValues(testCase.expectedInnerChunkSize, chunkGroup.chunkSize,
			"TestCase:%d,InnerChunkSize()", index)
		a.EqualValues(testCase.expectedChunkNum, chunkGroup.Num(),
			"TestCase:%d,len(Chunks)", index)
		a.EqualValues(testCase.fileSize, chunkGroup.Total())

		for cIndex, info := range testCase.expectedInfo {
			a.True(chunkGroup.Next())
			a.EqualValues(info[1], chunkGroup.Length(),
				"TestCase:%d,Chunks[%d].Length()", index, cIndex)
			a.EqualValues(info[0], chunkGroup.Start(),
				"TestCase:%d,Chunks[%d].Start()", index, cIndex)

			a.Equal(cIndex == len(testCase.expectedInfo)-1, chunkGroup.IsLast(),
				"TestCase:%d,Chunks[%d].IsLast()", index, cIndex)

			a.NotEmpty(chunkGroup.RangeHeader())
		}
		a.False(chunkGroup.Next())
	}
}

func TestChunkGroup_TempAvailablet(t *testing.T) {
	a := assert.New(t)

	file := &fsctx.FileStream{Size: 1}
	c := NewChunkGroup(file, 0, &backoff.ConstantBackoff{}, true)
	a.False(c.TempAvailable())

	f, err := os.CreateTemp("", "TestChunkGroup_TempAvailablet.*")
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	a.NoError(err)
	c.bufferTemp = f

	a.False(c.TempAvailable())
	f.Write([]byte("1"))
	a.True(c.TempAvailable())

}

func TestChunkGroup_Process(t *testing.T) {
	a := assert.New(t)
	file := &fsctx.FileStream{Size: 10}

	// success
	{
		file.File = io.NopCloser(strings.NewReader("1234567890"))
		c := NewChunkGroup(file, 5, &backoff.ConstantBackoff{}, true)
		count := 0
		a.True(c.Next())
		a.NoError(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			res, err := io.ReadAll(chunk)
			a.NoError(err)
			a.EqualValues("12345", string(res))
			return nil
		}))
		a.True(c.Next())
		a.NoError(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			res, err := io.ReadAll(chunk)
			a.NoError(err)
			a.EqualValues("67890", string(res))
			return nil
		}))
		a.False(c.Next())
		a.Equal(2, count)
	}

	// retry, read from buffer file
	{
		file.File = io.NopCloser(strings.NewReader("1234567890"))
		c := NewChunkGroup(file, 5, &backoff.ConstantBackoff{Max: 2}, true)
		count := 0
		a.True(c.Next())
		a.NoError(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			res, err := io.ReadAll(chunk)
			a.NoError(err)
			a.EqualValues("12345", string(res))
			return nil
		}))
		a.True(c.Next())
		a.NoError(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			res, err := io.ReadAll(chunk)
			a.NoError(err)
			a.EqualValues("67890", string(res))
			if count == 2 {
				return errors.New("error")
			}
			return nil
		}))
		a.False(c.Next())
		a.Equal(3, count)
	}

	// retry, read from seeker
	{
		f, _ := os.CreateTemp("", "TestChunkGroup_Process.*")
		f.Write([]byte("1234567890"))
		f.Seek(0, 0)
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()
		file.File = f
		file.Seeker = f
		c := NewChunkGroup(file, 5, &backoff.ConstantBackoff{Max: 2}, false)
		count := 0
		a.True(c.Next())
		a.NoError(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			res, err := io.ReadAll(chunk)
			a.NoError(err)
			a.EqualValues("12345", string(res))
			return nil
		}))
		a.True(c.Next())
		a.NoError(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			res, err := io.ReadAll(chunk)
			a.NoError(err)
			a.EqualValues("67890", string(res))
			if count == 2 {
				return errors.New("error")
			}
			return nil
		}))
		a.False(c.Next())
		a.Equal(3, count)
	}

	// retry, seek error
	{
		f, _ := os.CreateTemp("", "TestChunkGroup_Process.*")
		f.Write([]byte("1234567890"))
		f.Seek(0, 0)
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()
		file.File = f
		file.Seeker = f
		c := NewChunkGroup(file, 5, &backoff.ConstantBackoff{Max: 2}, false)
		count := 0
		a.True(c.Next())
		a.NoError(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			res, err := io.ReadAll(chunk)
			a.NoError(err)
			a.EqualValues("12345", string(res))
			return nil
		}))
		a.True(c.Next())
		f.Close()
		a.Error(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			if count == 2 {
				return errors.New("error")
			}
			return nil
		}))
		a.False(c.Next())
		a.Equal(2, count)
	}

	// retry, finally error
	{
		f, _ := os.CreateTemp("", "TestChunkGroup_Process.*")
		f.Write([]byte("1234567890"))
		f.Seek(0, 0)
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()
		file.File = f
		file.Seeker = f
		c := NewChunkGroup(file, 5, &backoff.ConstantBackoff{Max: 2}, false)
		count := 0
		a.True(c.Next())
		a.NoError(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			res, err := io.ReadAll(chunk)
			a.NoError(err)
			a.EqualValues("12345", string(res))
			return nil
		}))
		a.True(c.Next())
		a.Error(c.Process(func(c *ChunkGroup, chunk io.Reader) error {
			count++
			return errors.New("error")
		}))
		a.False(c.Next())
		a.Equal(4, count)
	}
}
