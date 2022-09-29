package main

import (
	"context"
	_ "embed"
	"flag"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cloudreve/Cloudreve/v3/bootstrap"
	model "github.com/cloudreve/Cloudreve/v3/models"
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

	staticFS = archiver.ArchiveFS{
		Stream: io.NewSectionReader(strings.NewReader(staticZip), 0, int64(len(staticZip))),
		Format: archiver.Zip{},
	}
	bootstrap.Init(confPath, staticFS)
}

func main() {
	// 关闭数据库连接
	defer model.DB.Close()

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
	server := &http.Server{Handler: api}

	// 收到信号后关闭服务器
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		sig := <-sigChan
		util.Log().Info("收到信号 %s，开始关闭 server", sig)
		ctx := context.Background()
		if conf.SystemConfig.GracePeriod != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(conf.SystemConfig.GracePeriod)*time.Second)
			defer cancel()
		}

		err := server.Shutdown(ctx)
		if err != nil {
			util.Log().Error("关闭 server 错误, %s", err)
		}
	}()

	// 如果启用了SSL
	if conf.SSLConfig.CertPath != "" {
		util.Log().Info("开始监听 %s", conf.SSLConfig.Listen)
		server.Addr = conf.SSLConfig.Listen
		if err := server.ListenAndServeTLS(conf.SSLConfig.CertPath, conf.SSLConfig.KeyPath); err != nil {
			util.Log().Error("无法监听[%s]，%s", conf.SSLConfig.Listen, err)
			return
		}
	}

	// 如果启用了Unix
	if conf.UnixConfig.Listen != "" {
		// delete socket file before listening
		if _, err := os.Stat(conf.UnixConfig.Listen); err == nil {
			if err = os.Remove(conf.UnixConfig.Listen); err != nil {
				util.Log().Error("删除 socket 文件错误, %s", err)
				return
			}
		}

		api.TrustedPlatform = conf.UnixConfig.ProxyHeader
		util.Log().Info("开始监听 %s", conf.UnixConfig.Listen)
		if err := RunUnix(server); err != nil {
			util.Log().Error("无法监听[%s]，%s", conf.UnixConfig.Listen, err)
		}
		return
	}

	util.Log().Info("开始监听 %s", conf.SystemConfig.Listen)
	server.Addr = conf.SystemConfig.Listen
	if err := server.ListenAndServe(); err != nil {
		util.Log().Error("无法监听[%s]，%s", conf.SystemConfig.Listen, err)
	}
}

func RunUnix(server *http.Server) error {
	listener, err := net.Listen("unix", conf.UnixConfig.Listen)
	if err != nil {
		return err
	}
	defer listener.Close()
	defer os.Remove(conf.UnixConfig.Listen)

	return server.Serve(listener)
}
