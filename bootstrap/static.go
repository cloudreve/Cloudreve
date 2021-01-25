package bootstrap

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	_ "github.com/cloudreve/Cloudreve/v3/statik"
	"github.com/gin-contrib/static"
	"github.com/rakyll/statik/fs"
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
func InitStatic() {
	var err error

	if util.Exists(util.RelativePath(StaticFolder)) {
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

// Eject 抽离内置静态资源
func Eject() {
	staticFS, err := fs.New()
	if err != nil {
		util.Log().Panic("无法初始化静态资源, %s", err)
	}

	root, err := staticFS.Open("/")
	if err != nil {
		util.Log().Panic("根目录不存在, %s", err)
	}

	var walk func(relPath string, object http.File)
	walk = func(relPath string, object http.File) {
		stat, err := object.Stat()
		if err != nil {
			util.Log().Error("无法获取[%s]的信息, %s, 跳过...", relPath, err)
			return
		}

		if !stat.IsDir() {
			// 写入文件
			out, err := util.CreatNestedFile(util.RelativePath(StaticFolder + relPath))
			defer out.Close()

			if err != nil {
				util.Log().Error("无法创建文件[%s], %s, 跳过...", relPath, err)
				return
			}

			util.Log().Info("导出 [%s]...", relPath)
			if _, err := io.Copy(out, object); err != nil {
				util.Log().Error("无法写入文件[%s], %s, 跳过...", relPath, err)
				return
			}
		} else {
			// 列出目录
			objects, err := object.Readdir(0)
			if err != nil {
				util.Log().Error("无法步入子目录[%s], %s, 跳过...", relPath, err)
				return
			}

			// 递归遍历子目录
			for _, newObject := range objects {
				newPath := path.Join(relPath, newObject.Name())
				newRoot, err := staticFS.Open(newPath)
				if err != nil {
					util.Log().Error("无法打开对象[%s], %s, 跳过...", newPath, err)
					continue
				}
				walk(newPath, newRoot)
			}

		}
	}

	util.Log().Info("开始导出内置静态资源...")
	walk("/", root)
	util.Log().Info("内置静态资源导出完成")
}
