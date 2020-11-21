// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdav

import (
	"context"
	"net/http"
	"path"
	"path/filepath"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
)

// slashClean is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

// moveFiles moves files and/or directories from src to dst.
//
// See section 9.9.4 for when various HTTP status codes apply.
func moveFiles(ctx context.Context, fs *filesystem.FileSystem, src FileInfo, dst string, overwrite bool) (status int, err error) {

	var (
		fileIDs   []uint
		folderIDs []uint
	)
	if src.IsDir() {
		folderIDs = []uint{src.(*model.Folder).ID}
	} else {
		fileIDs = []uint{src.(*model.File).ID}
	}

	// 判断是否为重命名
	if src.GetPosition() == path.Dir(dst) {
		err = fs.Rename(
			ctx,
			folderIDs,
			fileIDs,
			path.Base(dst),
		)
	} else {
		err = fs.Move(
			ctx,
			folderIDs,
			fileIDs,
			src.GetPosition(),
			path.Dir(dst),
		)
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusNoContent, nil
}

// copyFiles copies files and/or directories from src to dst.
//
// See section 9.8.5 for when various HTTP status codes apply.
func copyFiles(ctx context.Context, fs *filesystem.FileSystem, src FileInfo, dst string, overwrite bool, depth int, recursion int) (status int, err error) {
	if recursion == 1000 {
		return http.StatusInternalServerError, errRecursionTooDeep
	}
	recursion++

	if src.IsDir() {
		err := fs.Copy(
			ctx,
			[]uint{src.(*model.Folder).ID},
			[]uint{}, src.(*model.Folder).Position,
			path.Dir(dst),
		)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	} else {
		err := fs.Copy(ctx, []uint{}, []uint{src.(*model.File).ID}, src.(*model.File).Position, path.Dir(dst))
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusNoContent, nil
}

// walkFS traverses filesystem fs starting at name up to depth levels.
//
// Allowed values for depth are 0, 1 or infiniteDepth. For each visited node,
// walkFS calls walkFn. If a visited file system node is a directory and
// walkFn returns filepath.SkipDir, walkFS will skip traversal of this node.
func walkFS(
	ctx context.Context,
	fs *filesystem.FileSystem,
	depth int,
	name string,
	info FileInfo,
	walkFn func(reqPath string, info FileInfo, err error) error) error {
	// This implementation is based on Walk's code in the standard path/filepath package.
	err := walkFn(name, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}
	if !info.IsDir() || depth == 0 {
		return nil
	}
	if depth == 1 {
		depth = 0
	}

	dirs, _ := info.(*model.Folder).GetChildFolder()
	files, _ := info.(*model.Folder).GetChildFiles()

	for _, fileInfo := range files {
		filename := path.Join(name, fileInfo.Name)
		err = walkFS(ctx, fs, depth, filename, &fileInfo, walkFn)
		if err != nil {
			if !fileInfo.IsDir() || err != filepath.SkipDir {
				return err
			}
		}
	}

	for _, fileInfo := range dirs {
		filename := path.Join(name, fileInfo.Name)
		err = walkFS(ctx, fs, depth, filename, &fileInfo, walkFn)
		if err != nil {
			if !fileInfo.IsDir() || err != filepath.SkipDir {
				return err
			}
		}
	}
	return nil
}
