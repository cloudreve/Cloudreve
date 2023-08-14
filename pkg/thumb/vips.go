package thumb

import (
	"bytes"
	"context"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gofrs/uuid"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	RegisterGenerator(&VipsGenerator{})
}

type VipsGenerator struct {
	exts        []string
	lastRawExts string
}

func (v *VipsGenerator) Generate(ctx context.Context, file io.Reader, src, url, name string, options map[string]string) (*Result, error) {
	vipsOpts := model.GetSettingByNames("thumb_vips_path", "thumb_vips_exts", "thumb_encode_quality", "thumb_encode_method", "temp_path")

	if v.lastRawExts != vipsOpts["thumb_vips_exts"] {
		v.exts = strings.Split(vipsOpts["thumb_vips_exts"], ",")
	}

	if !util.IsInExtensionList(v.exts, name) {
		return nil, fmt.Errorf("unsupported image format: %w", ErrPassThrough)
	}

	outputOpt := ".png"
	if vipsOpts["thumb_encode_method"] == "jpg" {
		outputOpt = fmt.Sprintf(".jpg[Q=%s]", vipsOpts["thumb_encode_quality"])
	}

	cmd := exec.CommandContext(ctx,
		vipsOpts["thumb_vips_path"], "thumbnail_source", "[descriptor=0]", outputOpt, options["thumb_width"],
		"--height", options["thumb_height"])

	tempPath := filepath.Join(
		util.RelativePath(vipsOpts["temp_path"]),
		"thumb",
		fmt.Sprintf("thumb_%s", uuid.Must(uuid.NewV4()).String()),
	)

	thumbFile, err := util.CreatNestedFile(tempPath)
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

	return &Result{Path: tempPath}, nil
}

func (v *VipsGenerator) Priority() int {
	return 100
}

func (v *VipsGenerator) EnableFlag() string {
	return "thumb_vips_enabled"
}
