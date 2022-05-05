package main

import (
	_ "embed"
	"bytes"
	"flag"
	"io"
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/cloudreve/Cloudreve/v3/bootstrap"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/cloudreve/Cloudreve/v3/routers"

	"github.com/mholt/archiver/v4"
)

var (
	isEject    bool
	confPath   string
	scriptName string
)

//go:embed assets.zip
var staticZip string

var staticFS fs.FS

func init() {
	flag.StringVar(&confPath, "c", util.RelativePath("conf.ini"), "配置文件路径")
	flag.BoolVar(&isEject, "eject", false, "导出内置静态资源")
	flag.StringVar(&scriptName, "database-script", "", "运行内置数据库助手脚本")
	flag.Parse()

	staticFS = &seekableFS{archiver.ArchiveFS{
		Stream: io.NewSectionReader(strings.NewReader(staticZip), 0, int64(len(staticZip))),
		Format: archiver.Zip{},
	}}
	bootstrap.Init(confPath, staticFS)
}

func main() {
	if isEject {
		// 开始导出内置静态资源文件
		bootstrap.Eject(staticFS)
		return
	}

	if scriptName != "" {
		// 开始运行助手数据库脚本
		bootstrap.RunScript(scriptName)
		return
	}

	api := routers.InitRouter()

	// 如果启用了SSL
	if conf.SSLConfig.CertPath != "" {
		go func() {
			util.Log().Info("开始监听 %s", conf.SSLConfig.Listen)
			if err := api.RunTLS(conf.SSLConfig.Listen,
				conf.SSLConfig.CertPath, conf.SSLConfig.KeyPath); err != nil {
				util.Log().Error("无法监听[%s]，%s", conf.SSLConfig.Listen, err)
			}
		}()
	}

	// 如果启用了Unix
	if conf.UnixConfig.Listen != "" {
		util.Log().Info("开始监听 %s", conf.UnixConfig.Listen)
		if err := api.RunUnix(conf.UnixConfig.Listen); err != nil {
			util.Log().Error("无法监听[%s]，%s", conf.UnixConfig.Listen, err)
		}
		return
	}

	util.Log().Info("开始监听 %s", conf.SystemConfig.Listen)
	if err := api.Run(conf.SystemConfig.Listen); err != nil {
		util.Log().Error("无法监听[%s]，%s", conf.SystemConfig.Listen, err)
	}
}

// https://github.com/golang/go/issues/46809
// A seekableFS is an FS wrapper that makes every file seekable
// by reading it entirely into memory when it is opened and then
// serving read operations (including seek) from the memory copy.
type seekableFS struct {
	fs fs.FS
}

func (s *seekableFS) Open(name string) (fs.File, error) {
	f, err := s.fs.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if info.IsDir() {
		return f, nil
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	var sf seekableFile
	sf.File = f
	sf.Reset(data)
	return &sf, nil
}

// A seekableFile is a fs.File augmented by an in-memory copy of the file data to allow use of Seek.
type seekableFile struct {
	bytes.Reader
	fs.File
}

// Read calls f.Reader.Read.
// Both f.Reader and f.File have Read methods - a conflict - so f inherits neither.
// This method calls the one we want.
func (f *seekableFile) Read(b []byte) (int, error) {
	return f.Reader.Read(b)
}