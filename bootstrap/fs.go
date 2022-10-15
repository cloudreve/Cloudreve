package bootstrap

import (
	"archive/zip"
	"crypto/sha256"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/pkg/errors"
	"io"
	"io/fs"
	"sort"
	"strings"
)

func NewFS(zipContent string) fs.FS {
	zipReader, err := zip.NewReader(strings.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		util.Log().Panic("Static resource is not a valid zip file: %s", err)
	}

	var files []file
	err = fs.WalkDir(zipReader, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Errorf("无法获取[%s]的信息, %s, 跳过...", path, err)
		}

		if path == "." {
			return nil
		}

		var f file
		if d.IsDir() {
			f.name = path + "/"
		} else {
			f.name = path

			rc, err := zipReader.Open(path)
			if err != nil {
				return errors.Errorf("无法打开文件[%s], %s, 跳过...", path, err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return errors.Errorf("无法读取文件[%s], %s, 跳过...", path, err)
			}

			f.data = string(data)

			hash := sha256.Sum256(data)
			for i := range f.hash {
				f.hash[i] = ^hash[i]
			}
		}
		files = append(files, f)
		return nil
	})
	if err != nil {
		util.Log().Panic("初始化静态资源失败: %s", err)
	}

	sort.Slice(files, func(i, j int) bool {
		fi, fj := files[i], files[j]
		di, ei, _ := split(fi.name)
		dj, ej, _ := split(fj.name)

		if di != dj {
			return di < dj
		}
		return ei < ej
	})

	var embedFS FS
	embedFS.files = &files
	return embedFS
}
