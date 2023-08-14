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
	RegisterGenerator(&LibreOfficeGenerator{})
}

type LibreOfficeGenerator struct {
	exts        []string
	lastRawExts string
}

func (l *LibreOfficeGenerator) Generate(ctx context.Context, file io.Reader, src, url, name string, options map[string]string) (*Result, error) {
	sofficeOpts := model.GetSettingByNames("thumb_libreoffice_path", "thumb_libreoffice_exts", "thumb_encode_method", "temp_path")

	if l.lastRawExts != sofficeOpts["thumb_libreoffice_exts"] {
		l.exts = strings.Split(sofficeOpts["thumb_libreoffice_exts"], ",")
	}

	if !util.IsInExtensionList(l.exts, name) {
		return nil, fmt.Errorf("unsupported document format: %w", ErrPassThrough)
	}

	tempOutputPath := filepath.Join(
		util.RelativePath(sofficeOpts["temp_path"]),
		"thumb",
		fmt.Sprintf("soffice_%s", uuid.Must(uuid.NewV4()).String()),
	)

	tempInputPath := src
	if tempInputPath == "" {
		// If not local policy files, download to temp folder
		tempInputPath = filepath.Join(
			util.RelativePath(sofficeOpts["temp_path"]),
			"thumb",
			fmt.Sprintf("soffice_%s%s", uuid.Must(uuid.NewV4()).String(), filepath.Ext(name)),
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

	// Convert the document to an image
	cmd := exec.CommandContext(ctx, sofficeOpts["thumb_libreoffice_path"], "--headless",
		"-nologo", "--nofirststartwizard", "--invisible", "--norestore", "--convert-to",
		sofficeOpts["thumb_encode_method"], "--outdir", tempOutputPath, tempInputPath)

	// Redirect IO
	var stdErr bytes.Buffer
	cmd.Stdin = file
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		util.Log().Warning("Failed to invoke LibreOffice: %s", stdErr.String())
		return nil, fmt.Errorf("failed to invoke LibreOffice: %w", err)
	}

	return &Result{
		Path: filepath.Join(
			tempOutputPath,
			strings.TrimSuffix(filepath.Base(tempInputPath), filepath.Ext(tempInputPath))+"."+sofficeOpts["thumb_encode_method"],
		),
		Continue: true,
		Cleanup:  []func(){func() { _ = os.RemoveAll(tempOutputPath) }},
	}, nil
}

func (l *LibreOfficeGenerator) Priority() int {
	return 50
}

func (l *LibreOfficeGenerator) EnableFlag() string {
	return "thumb_libreoffice_enabled"
}
