package bootstrap

import (
	"bufio"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"

	"github.com/gin-contrib/static"
)

const StaticFolder = "statics"

type GinFS struct {
	FS http.FileSystem
}

type staticVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// StaticFS 内置静态文件资源
var StaticFS static.ServeFileSystem

// Open 打开文件
func (b *GinFS) Open(name string) (http.File, error) {
	return b.FS.Open(name)
}

// Exists 文件是否存在
func (b *GinFS) Exists(prefix string, filepath string) bool {
	if _, err := b.FS.Open(filepath); err != nil {
		return false
	}
	return true
}

// InitStatic 初始化静态资源文件
func InitStatic(statics fs.FS) {
	if util.Exists(util.RelativePath(StaticFolder)) {
		util.Log().Info("Folder with name \"statics\" already exists, it will be used to serve static files.")
		StaticFS = static.LocalFile(util.RelativePath("statics"), false)
	} else {
		// 初始化静态资源
		embedFS, err := fs.Sub(statics, "assets/build")
		if err != nil {
			util.Log().Panic("Failed to initialize static resources: %s", err)
		}

		StaticFS = &GinFS{
			FS: http.FS(embedFS),
		}
	}
	// 检查静态资源的版本
	f, err := StaticFS.Open("version.json")
	if err != nil {
		util.Log().Warning("Missing version identifier file in static resources, please delete \"statics\" folder and rebuild it.")
		return
	}

	b, err := io.ReadAll(f)
	if err != nil {
		util.Log().Warning("Failed to read version identifier file in static resources, please delete \"statics\" folder and rebuild it.")
		return
	}

	var v staticVersion
	if err := json.Unmarshal(b, &v); err != nil {
		util.Log().Warning("Failed to parse version identifier file in static resources: %s", err)
		return
	}

	staticName := "cloudreve-frontend"
	if conf.IsPlus == "true" {
		staticName += "-plus"
	}

	if v.Name != staticName {
		util.Log().Warning("Static resource version mismatch, please delete \"statics\" folder and rebuild it.")
		return
	}

	if v.Version != conf.RequiredStaticVersion {
		util.Log().Warning("Static resource version mismatch [Current %s, Desired: %s]，please delete \"statics\" folder and rebuild it.", v.Version, conf.RequiredStaticVersion)
		return
	}
}

// Eject 抽离内置静态资源
func Eject(statics fs.FS) {
	// 初始化静态资源
	embedFS, err := fs.Sub(statics, "assets/build")
	if err != nil {
		util.Log().Panic("Failed to initialize static resources: %s", err)
	}

	// var walk func(relPath string, d fs.DirEntry, err error) error
	walk := func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Errorf("Failed to read info of %q: %s, skipping...", relPath, err)
		}

		if !d.IsDir() {
			// 写入文件
			out, err := util.CreatNestedFile(filepath.Join(util.RelativePath(""), StaticFolder, relPath))

			if err != nil {
				return errors.Errorf("Failed to create file %q: %s, skipping...", relPath, err)
			}
			defer out.Close()

			util.Log().Info("Ejecting %q...", relPath)
			obj, _ := embedFS.Open(relPath)
			if _, err := io.Copy(out, bufio.NewReader(obj)); err != nil {
				return errors.Errorf("Cannot write file %q: %s, skipping...", relPath, err)
			}
		}
		return nil
	}

	// util.Log().Info("开始导出内置静态资源...")
	err = fs.WalkDir(embedFS, ".", walk)
	if err != nil {
		util.Log().Error("Error occurs while ejecting static resources: %s", err)
		return
	}
	util.Log().Info("Finish ejecting static resources.")
}
