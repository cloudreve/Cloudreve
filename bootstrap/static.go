package bootstrap

import (
	"encoding/json"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/util"
	_ "github.com/HFO4/cloudreve/statik"
	"github.com/gin-contrib/static"
	"github.com/rakyll/statik/fs"
	"io/ioutil"
	"net/http"
)

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
func InitStatic() {
	var err error

	if util.Exists(util.RelativePath("statics")) {
		util.Log().Info("检测到 statics 目录存在，将使用此目录下的静态资源文件")
		StaticFS = static.LocalFile(util.RelativePath("statics"), false)

		// 检查静态资源的版本
		f, err := StaticFS.Open("version.json")
		if err != nil {
			util.Log().Warning("静态资源版本标识文件不存在，请重新构建或删除 statics 目录")
			return
		}

		b, err := ioutil.ReadAll(f)
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

	} else {
		StaticFS = &GinFS{}
		StaticFS.(*GinFS).FS, err = fs.New()
		if err != nil {
			util.Log().Panic("无法初始化静态资源, %s", err)
		}
	}

}
