package crontab

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

func garbageCollect() {
	// 清理打包下载产生的临时文件
	collectArchiveFile()

	// 清理过期的内置内存缓存
	if store, ok := cache.Store.(*cache.MemoStore); ok {
		collectCache(store)
	}

	util.Log().Info("Crontab job \"cron_garbage_collect\" complete.")
}

func collectArchiveFile() {
	// 读取有效期、目录设置
	tempPath := util.RelativePath(model.GetSettingByName("temp_path"))
	expires := model.GetIntSetting("download_timeout", 30)

	// 列出文件
	root := filepath.Join(tempPath, "archive")
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() &&
			strings.HasPrefix(filepath.Base(path), "archive_") &&
			time.Now().Sub(info.ModTime()).Seconds() > float64(expires) {
			util.Log().Debug("Delete expired batch download temp file %q.", path)
			// 删除符合条件的文件
			if err := os.Remove(path); err != nil {
				util.Log().Debug("Failed to delete temp file %q: %s", path, err)
			}
		}
		return nil
	})

	if err != nil {
		util.Log().Debug("Crontab job cannot list temp batch download folder: %s", err)
	}

}

func collectCache(store *cache.MemoStore) {
	util.Log().Debug("Cleanup memory cache.")
	store.GarbageCollect()
}

func uploadSessionCollect() {
	placeholders := model.GetUploadPlaceholderFiles(0)

	// 将过期的上传会话按照用户分组
	userToFiles := make(map[uint][]uint)
	for _, file := range placeholders {
		_, sessionExist := cache.Get(filesystem.UploadSessionCachePrefix + *file.UploadSessionID)
		if sessionExist {
			continue
		}

		if _, ok := userToFiles[file.UserID]; !ok {
			userToFiles[file.UserID] = make([]uint, 0)
		}

		userToFiles[file.UserID] = append(userToFiles[file.UserID], file.ID)
	}

	// 删除过期的会话
	for uid, filesIDs := range userToFiles {
		user, err := model.GetUserByID(uid)
		if err != nil {
			util.Log().Warning("Owner of the upload session cannot be found: %s", err)
			continue
		}

		fs, err := filesystem.NewFileSystem(&user)
		if err != nil {
			util.Log().Warning("Failed to initialize filesystem: %s", err)
			continue
		}

		if err = fs.Delete(context.Background(), []uint{}, filesIDs, false, false); err != nil {
			util.Log().Warning("Failed to delete upload session: %s", err)
		}

		fs.Recycle()
	}

	util.Log().Info("Crontab job \"cron_recycle_upload_session\" complete.")
}
