package thumb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gofrs/uuid"
)

func init() {
	RegisterGenerator(&VipsGenerator{})
}

type VipsGenerator struct {
	exts        []string
	lastRawExts string
}

func (v *VipsGenerator) Generate(ctx context.Context, file io.Reader, src, name string, options map[string]string) (*Result, error) {
	const (
		thumbVipsPath      = "thumb_vips_path"
		thumbVipsExts      = "thumb_vips_exts"
		thumbEncodeQuality = "thumb_encode_quality"
		thumbEncodeMethod  = "thumb_encode_method"
		tempPath           = "temp_path"
	)
	vipsOpts := model.GetSettingByNames(thumbVipsPath, thumbVipsExts, thumbEncodeQuality, thumbEncodeMethod, tempPath)

	if v.lastRawExts != vipsOpts[thumbVipsExts] {
		v.exts = strings.Split(vipsOpts[thumbVipsExts], ",")
		v.lastRawExts = vipsOpts[thumbVipsExts]
	}

	if !util.IsInExtensionList(v.exts, name) {
		return nil, fmt.Errorf("unsupported image format: %w", ErrPassThrough)
	}

	outputOpt := ".png"
	if vipsOpts[thumbEncodeMethod] == "jpg" {
		outputOpt = fmt.Sprintf(".jpg[Q=%s]", vipsOpts[thumbEncodeQuality])
	}

	cmd := exec.CommandContext(ctx,
		vipsOpts[thumbVipsPath], "thumbnail_source", "[descriptor=0]", outputOpt, options["thumb_width"],
		"--height", options["thumb_height"])

	outTempPath := filepath.Join(
		util.RelativePath(vipsOpts[tempPath]),
		"thumb",
		fmt.Sprintf("thumb_%s", uuid.Must(uuid.NewV4()).String()),
	)

	thumbFile, err := util.CreatNestedFile(outTempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	defer thumbFile.Close()

	// Redirect IO
	var vipsErr bytes.Buffer
	cmd.Stdin = file
	cmd.Stdout = thumbFile
	cmd.Stderr = &vipsErr

	if err := cmd.Run(); err != nil {
		util.Log().Warning("Failed to invoke vips: %s", vipsErr.String())
		return nil, fmt.Errorf("failed to invoke vips: %w", err)
	}

	return &Result{Path: outTempPath}, nil
}

func (v *VipsGenerator) Priority() int {
	return 100
}

func (v *VipsGenerator) EnableFlag() string {
	return "thumb_vips_enabled"
}
