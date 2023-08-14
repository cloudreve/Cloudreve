package thumb

import (
	"bytes"
	"context"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gofrs/uuid"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	RegisterGenerator(&FfmpegGenerator{})
}

type FfmpegGenerator struct {
	exts        []string
	lastRawExts string
}

func (f *FfmpegGenerator) Generate(ctx context.Context, file io.Reader, src, url, name string, options map[string]string) (*Result, error) {
	ffmpegOpts := model.GetSettingByNames("thumb_ffmpeg_path", "thumb_ffmpeg_exts", "thumb_ffmpeg_seek", "thumb_encode_method", "temp_path")

	if f.lastRawExts != ffmpegOpts["thumb_ffmpeg_exts"] {
		f.exts = strings.Split(ffmpegOpts["thumb_ffmpeg_exts"], ",")
	}

	if !util.IsInExtensionList(f.exts, name) {
		return nil, fmt.Errorf("unsupported video format: %w", ErrPassThrough)
	}

	tempOutputPath := filepath.Join(
		util.RelativePath(ffmpegOpts["temp_path"]),
		"thumb",
		fmt.Sprintf("thumb_%s.%s", uuid.Must(uuid.NewV4()).String(), ffmpegOpts["thumb_encode_method"]),
	)

	tempInputPath := src
	if url != "" {
		tempInputPath = url
	}
	if tempInputPath == "" {
		// If not local policy files, download to temp folder
		tempInputPath = filepath.Join(
			util.RelativePath(ffmpegOpts["temp_path"]),
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
		ffmpegOpts["thumb_ffmpeg_path"], "-ss", ffmpegOpts["thumb_ffmpeg_seek"], "-i", tempInputPath,
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
