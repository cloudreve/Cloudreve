package manager

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs/dbfs"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/manager/entitysource"
	"golang.org/x/tools/container/intsets"
)

func (m *manager) CreateArchive(ctx context.Context, uris []*fs.URI, writer io.Writer, opts ...fs.Option) (int, error) {
	o := newOption()
	for _, opt := range opts {
		opt.Apply(o)
	}

	failed := 0

	// List all top level files
	files := make([]fs.File, 0, len(uris))
	for _, uri := range uris {
		file, err := m.Get(ctx, uri, dbfs.WithFileEntities(), dbfs.WithRequiredCapabilities(dbfs.NavigatorCapabilityDownloadFile))
		if err != nil {
			return 0, fmt.Errorf("failed to get file %s: %w", uri, err)
		}

		files = append(files, file)
	}

	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	var compressed int64
	for _, file := range files {
		if file.Type() == types.FileTypeFile {
			if err := m.compressFileToArchive(ctx, "/", file, zipWriter, o.ArchiveCompression, o.DryRun); err != nil {
				failed++
				m.l.Warning("Failed to compress file %s: %s, skipping it...", file.Uri(false), err)
			}

			compressed += file.Size()
			if o.ProgressFunc != nil {
				o.ProgressFunc(compressed, file.Size(), 0)
			}

			if o.MaxArchiveSize > 0 && compressed > o.MaxArchiveSize {
				return 0, fs.ErrArchiveSrcSizeTooBig
			}

		} else {
			if err := m.Walk(ctx, file.Uri(false), intsets.MaxInt, func(f fs.File, level int) error {
				if f.Type() == types.FileTypeFolder || f.IsSymbolic() {
					return nil
				}
				if err := m.compressFileToArchive(ctx, strings.TrimPrefix(f.Uri(false).Dir(),
					file.Uri(false).Dir()), f, zipWriter, o.ArchiveCompression, o.DryRun); err != nil {
					failed++
					m.l.Warning("Failed to compress file %s: %s, skipping it...", f.Uri(false), err)
				}

				compressed += f.Size()
				if o.ProgressFunc != nil {
					o.ProgressFunc(compressed, f.Size(), 0)
				}

				if o.MaxArchiveSize > 0 && compressed > o.MaxArchiveSize {
					return fs.ErrArchiveSrcSizeTooBig
				}

				return nil
			}); err != nil {
				m.l.Warning("Failed to walk folder %s: %s, skipping it...", file.Uri(false), err)
				failed++
			}
		}
	}

	return failed, nil
}

func (m *manager) compressFileToArchive(ctx context.Context, parent string, file fs.File, zipWriter *zip.Writer,
	compression bool, dryrun fs.CreateArchiveDryRunFunc) error {
	es, err := m.GetEntitySource(ctx, file.PrimaryEntityID())
	if err != nil {
		return fmt.Errorf("failed to get entity source for file %s: %w", file.Uri(false), err)
	}

	zipName := filepath.FromSlash(path.Join(parent, file.DisplayName()))
	if dryrun != nil {
		dryrun(zipName, es.Entity())
		return nil
	}

	m.l.Debug("Compressing %s to archive...", file.Uri(false))
	header := &zip.FileHeader{
		Name:               zipName,
		Modified:           file.UpdatedAt(),
		UncompressedSize64: uint64(file.Size()),
	}

	if !compression {
		header.Method = zip.Store
	} else {
		header.Method = zip.Deflate
	}

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip header for %s: %w", file.Uri(false), err)
	}

	es.Apply(entitysource.WithContext(ctx))
	_, err = io.Copy(writer, es)
	return err

}
