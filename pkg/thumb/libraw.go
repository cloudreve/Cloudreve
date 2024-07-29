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
	RegisterGenerator(&LibRawGenerator{})
}

type LibRawGenerator struct {
	exts        []string
	lastRawExts string
}

func (f *LibRawGenerator) Generate(ctx context.Context, file io.Reader, _ string, name string, options map[string]string) (*Result, error) {
	const (
		thumbLibRawPath = "thumb_libraw_path"
		thumbLibRawExt  = "thumb_libraw_exts"
		thumbTempPath   = "temp_path"
	)

	opts := model.GetSettingByNames(thumbLibRawPath, thumbLibRawExt, thumbTempPath)

	if f.lastRawExts != opts[thumbLibRawExt] {
		f.exts = strings.Split(opts[thumbLibRawExt], ",")
		f.lastRawExts = opts[thumbLibRawExt]
	}

	if !util.IsInExtensionList(f.exts, name) {
		return nil, fmt.Errorf("unsupported image format: %w", ErrPassThrough)
	}

	inputFilePath := filepath.Join(
		util.RelativePath(opts[thumbTempPath]),
		"thumb",
		fmt.Sprintf("thumb_%s", uuid.Must(uuid.NewV4()).String()),
	)
	defer func() { _ = os.Remove(inputFilePath) }()

	inputFile, err := util.CreatNestedFile(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err = io.Copy(inputFile, file); err != nil {
		_ = inputFile.Close()
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}
	_ = inputFile.Close()

	cmd := exec.CommandContext(ctx, opts[thumbLibRawPath], "-e", inputFilePath)

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	if err = cmd.Run(); err != nil {
		util.Log().Warning("Failed to invoke LibRaw: %s", stdErr.String())
		return nil, fmt.Errorf("failed to invoke LibRaw: %w", err)
	}

	outputFilePath := inputFilePath + ".thumb.jpg"
	defer func() { _ = os.Remove(outputFilePath) }()

	// use builtin function
	ff, err := os.OpenFile(outputFilePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer func() { _ = ff.Close() }()

	return new(Builtin).Generate(ctx, ff, outputFilePath, filepath.Base(outputFilePath), options)
}

func (f *LibRawGenerator) Priority() int {
	return 250
}

func (f *LibRawGenerator) EnableFlag() string {
	return "thumb_libraw_enabled"
}

var _ Generator = (*LibRawGenerator)(nil)
