package bootstrap

import (
	"bufio"
	"embed"
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
func InitStatic(statics embed.FS) {
	if util.Exists(util.RelativePath(StaticFolder)) {
		util.Log().Info("检测到 statics 目录存在，将使用此目录下的静态资源文件")
		StaticFS = static.LocalFile(util.RelativePath("statics"), false)
	} else {
		// 初始化静态资源
		embedFS, err := fs.Sub(statics, "assets/build")
		if err != nil {
			util.Log().Panic("无法初始化静态资源, %s", err)
		}

		StaticFS = &GinFS{
			FS: http.FS(embedFS),
		}
	}
	// 检查静态资源的版本
	f, err := StaticFS.Open("version.json")
	if err != nil {
		util.Log().Warning("静态资源版本标识文件不存在，请重新构建或删除 statics 目录")
		return
	}

	b, err := io.ReadAll(f)
	if err != nil {
		util.Log().Warning("无法读取静态资源文件版本，请重新构建或删除 statics 目录")
		return
	}

	var v staticVersion
	if err := json.Unmarshal(b, &v); err != nil {
		util.Log().Warning("无法解析静态资源文件版本, %s", err)
		return
	}

	staticName := "cloudreve-frontend"
	if conf.IsPro == "true" {
		staticName += "-pro"
	}

	if v.Name != staticName {
		util.Log().Warning("静态资源版本不匹配，请重新构建或删除 statics 目录")
		return
	}

	if v.Version != conf.RequiredStaticVersion {
		util.Log().Warning("静态资源版本不匹配 [当前 %s, 需要: %s]，请重新构建或删除 statics 目录", v.Version, conf.RequiredStaticVersion)
		return
	}
}

// Eject 抽离内置静态资源
func Eject(statics embed.FS) {
	// 初始化静态资源
	embedFS, err := fs.Sub(statics, "assets/build")
	if err != nil {
		util.Log().Panic("无法初始化静态资源, %s", err)
	}

	var walk func(relPath string, d fs.DirEntry, err error) error
	walk = func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Errorf("无法获取[%s]的信息, %s, 跳过...", relPath, err)
		}

		if !d.IsDir() {
			// 写入文件
			out, err := util.CreatNestedFile(filepath.Join(util.RelativePath(""), StaticFolder, relPath))
			defer out.Close()

			if err != nil {
				return errors.Errorf("无法创建文件[%s], %s, 跳过...", relPath, err)
			}

			util.Log().Info("导出 [%s]...", relPath)
			obj, _ := embedFS.Open(relPath)
			if _, err := io.Copy(out, bufio.NewReader(obj)); err != nil {
				return errors.Errorf("无法写入文件[%s], %s, 跳过...", relPath, err)
			}
		}
		return nil
	}

	// util.Log().Info("开始导出内置静态资源...")
	err = fs.WalkDir(embedFS, ".", walk)
	if err != nil {
		util.Log().Error("导出内置静态资源遇到错误：%s", err)
		return
	}
	util.Log().Info("内置静态资源导出完成")
}
