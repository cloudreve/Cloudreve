package thumb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gofrs/uuid"
)

func init() {
	RegisterGenerator(&FfmpegGenerator{})
}

type FfmpegGenerator struct {
	exts        []string
	lastRawExts string
}

func (f *FfmpegGenerator) Generate(ctx context.Context, file io.Reader, src, name string, options map[string]string) (*Result, error) {
	const (
		thumbFFMpegPath   = "thumb_ffmpeg_path"
		thumbFFMpegExts   = "thumb_ffmpeg_exts"
		thumbFFMpegSeek   = "thumb_ffmpeg_seek"
		thumbEncodeMethod = "thumb_encode_method"
		tempPath          = "temp_path"
	)
	ffmpegOpts := model.GetSettingByNames(thumbFFMpegPath, thumbFFMpegExts, thumbFFMpegSeek, thumbEncodeMethod, tempPath)

	if f.lastRawExts != ffmpegOpts[thumbFFMpegExts] {
		f.exts = strings.Split(ffmpegOpts[thumbFFMpegExts], ",")
		f.lastRawExts = ffmpegOpts[thumbFFMpegExts]
	}

	if !util.IsInExtensionList(f.exts, name) {
		return nil, fmt.Errorf("unsupported video format: %w", ErrPassThrough)
	}

	tempOutputPath := filepath.Join(
		util.RelativePath(ffmpegOpts[tempPath]),
		"thumb",
		fmt.Sprintf("thumb_%s.%s", uuid.Must(uuid.NewV4()).String(), ffmpegOpts[thumbEncodeMethod]),
	)

	tempInputPath := src
	if tempInputPath == "" {
		// If not local policy files, download to temp folder
		tempInputPath = filepath.Join(
			util.RelativePath(ffmpegOpts[tempPath]),
			"thumb",
			fmt.Sprintf("ffmpeg_%s%s", uuid.Must(uuid.NewV4()).String(), filepath.Ext(name)),
		)

		// Due to limitations of ffmpeg, we need to write the input file to disk first
		tempInputFile, err := util.CreatNestedFile(tempInputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}

		defer os.Remove(tempInputPath)
		defer tempInputFile.Close()

		if _, err = io.Copy(tempInputFile, file); err != nil {
			return nil, fmt.Errorf("failed to write input file: %w", err)
		}

		tempInputFile.Close()
	}

	// Invoke ffmpeg
	scaleOpt := fmt.Sprintf("scale=%s:%s:force_original_aspect_ratio=decrease", options["thumb_width"], options["thumb_height"])
	cmd := exec.CommandContext(ctx,
		ffmpegOpts[thumbFFMpegPath], "-ss", ffmpegOpts[thumbFFMpegSeek], "-i", tempInputPath,
		"-vf", scaleOpt, "-vframes", "1", tempOutputPath)

	// Redirect IO
	var stdErr bytes.Buffer
	cmd.Stdin = file
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		util.Log().Warning("Failed to invoke ffmpeg: %s", stdErr.String())
		return nil, fmt.Errorf("failed to invoke ffmpeg: %w", err)
	}

	return &Result{Path: tempOutputPath}, nil
}

func (f *FfmpegGenerator) Priority() int {
	return 200
}

func (f *FfmpegGenerator) EnableFlag() string {
	return "thumb_ffmpeg_enabled"
}
