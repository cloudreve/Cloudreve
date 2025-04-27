package statics

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"debug/buildinfo"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/cloudreve/Cloudreve/v4/application/constants"
	"github.com/cloudreve/Cloudreve/v4/pkg/logging"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
	"github.com/gin-contrib/static"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const StaticFolder = "statics"

//go:embed assets.zip
var zipContent string

type GinFS struct {
	FS http.FileSystem
}

type version struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

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

// NewServerStaticFS 初始化静态资源文件
func NewServerStaticFS(l logging.Logger, statics fs.FS, isPro bool) (static.ServeFileSystem, error) {
	var staticFS static.ServeFileSystem
	if util.Exists(util.DataPath(StaticFolder)) {
		l.Info("Folder with %q already exists, it will be used to serve static files.", util.DataPath(StaticFolder))
		staticFS = static.LocalFile(util.DataPath(StaticFolder), false)
	} else {
		// 初始化静态资源
		embedFS, err := fs.Sub(statics, "assets/build")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize static resources: %w", err)
		}

		staticFS = &GinFS{
			FS: http.FS(embedFS),
		}
	}
	// 检查静态资源的版本
	f, err := staticFS.Open("version.json")
	if err != nil {
		l.Warning("Missing version identifier file in static resources, please delete \"statics\" folder and rebuild it.")
		return staticFS, nil
	}

	b, err := io.ReadAll(f)
	if err != nil {
		l.Warning("Failed to read version identifier file in static resources, please delete \"statics\" folder and rebuild it.")
		return staticFS, nil
	}

	var v version
	if err := json.Unmarshal(b, &v); err != nil {
		l.Warning("Failed to parse version identifier file in static resources: %s", err)
		return staticFS, nil
	}

	staticName := "cloudreve-frontend"
	if isPro {
		staticName += "-pro"
	}

	if v.Name != staticName {
		l.Error("Static resource version mismatch, please delete \"statics\" folder and rebuild it.")
	}

	if v.Version != constants.BackendVersion {
		l.Error("Static resource version mismatch [Current %s, Desired: %s]，please delete \"statics\" folder and rebuild it.", v.Version, constants.BackendVersion)
	}

	return staticFS, nil
}

func NewStaticFS(l logging.Logger) fs.FS {
	zipReader, err := zip.NewReader(strings.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		l.Panic("Static resource is not a valid zip file: %s", err)
	}

	var files []file
	modTime := getBuildTime()
	err = fs.WalkDir(zipReader, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("cannot walk into %q: %w", path, err)
		}

		if path == "." {
			return nil
		}

		f := file{modTime: modTime}
		if d.IsDir() {
			f.name = path + "/"
		} else {
			f.name = path

			rc, err := zipReader.Open(path)
			if err != nil {
				return fmt.Errorf("canot open %q: %w", path, err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return fmt.Errorf("cannot read %q: %w", path, err)
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
		l.Panic("Failed to initialize static resources: %s", err)
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

// Eject 抽离内置静态资源
func Eject(l logging.Logger, statics fs.FS) error {
	// 初始化静态资源
	embedFS, err := fs.Sub(statics, "assets/build")
	if err != nil {
		l.Panic("Failed to initialize static resources: %s", err)
	}

	var walk func(relPath string, d fs.DirEntry, err error) error
	walk = func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to read info of %q: %s, skipping...", relPath, err)
		}

		if !d.IsDir() {
			// 写入文件
			dst := util.DataPath(filepath.Join(StaticFolder, relPath))
			out, err := util.CreatNestedFile(dst)
			defer out.Close()

			if err != nil {
				return fmt.Errorf("failed to create file %q: %s, skipping...", dst, err)
			}

			l.Info("Ejecting %q...", dst)
			obj, _ := embedFS.Open(relPath)
			if _, err := io.Copy(out, bufio.NewReader(obj)); err != nil {
				return fmt.Errorf("cannot write file %q: %s, skipping...", relPath, err)
			}
		}
		return nil
	}

	// util.Log().Info("开始导出内置静态资源...")
	err = fs.WalkDir(embedFS, ".", walk)
	if err != nil {
		return fmt.Errorf("failed to eject static resources: %w", err)
	}

	l.Info("Finish ejecting static resources.")
	return nil
}

func getBuildTime() (buildTime time.Time) {
	buildTime = time.Now()
	exe, err := os.Executable()
	if err != nil {
		return
	}
	info, err := buildinfo.ReadFile(exe)
	if err != nil {
		return
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.time" && s.Value != "" {
			if t, err := time.Parse(time.RFC3339, s.Value); err == nil {
				buildTime = t
			}
			break
		}
	}
	return
}
