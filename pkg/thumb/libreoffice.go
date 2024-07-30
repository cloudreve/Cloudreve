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
	RegisterGenerator(&LibreOfficeGenerator{})
}

type LibreOfficeGenerator struct {
	exts        []string
	lastRawExts string
}

func (l *LibreOfficeGenerator) Generate(ctx context.Context, file io.Reader, src string, name string, options map[string]string) (*Result, error) {
	const (
		thumbLibreOfficePath = "thumb_libreoffice_path"
		thumbLibreOfficeExts = "thumb_libreoffice_exts"
		thumbEncodeMethod    = "thumb_encode_method"
		tempPath             = "temp_path"
	)
	sofficeOpts := model.GetSettingByNames(thumbLibreOfficePath, thumbLibreOfficeExts, thumbEncodeMethod, tempPath)

	if l.lastRawExts != sofficeOpts[thumbLibreOfficeExts] {
		l.exts = strings.Split(sofficeOpts[thumbLibreOfficeExts], ",")
		l.lastRawExts = sofficeOpts[thumbLibreOfficeExts]
	}

	if !util.IsInExtensionList(l.exts, name) {
		return nil, fmt.Errorf("unsupported document format: %w", ErrPassThrough)
	}

	tempOutputPath := filepath.Join(
		util.RelativePath(sofficeOpts[tempPath]),
		"thumb",
		fmt.Sprintf("soffice_%s", uuid.Must(uuid.NewV4()).String()),
	)

	tempInputPath := src
	if tempInputPath == "" {
		// If not local policy files, download to temp folder
		tempInputPath = filepath.Join(
			util.RelativePath(sofficeOpts[tempPath]),
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
	cmd := exec.CommandContext(ctx, sofficeOpts[thumbLibreOfficePath], "--headless",
		"-nologo", "--nofirststartwizard", "--invisible", "--norestore", "--convert-to",
		sofficeOpts[thumbEncodeMethod], "--outdir", tempOutputPath, tempInputPath)

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
			strings.TrimSuffix(filepath.Base(tempInputPath), filepath.Ext(tempInputPath))+"."+sofficeOpts[thumbEncodeMethod],
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
