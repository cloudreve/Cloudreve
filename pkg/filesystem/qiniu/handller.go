package qiniu

import (
	"context"
	"fmt"
	"github.com/HFO4/cloudreve/pkg/util"
	"io"
	"os"

	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/storage"
)

// Handler 本地策略适配器
type Handler struct{}

// Put 将文件流保存到指定目录
func (handler Handler) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	// 凭证生成
	putPolicy := storage.PutPolicy{
		Scope: "cloudrevetest",
	}
	mac := auth.New("YNzTBBpDUq4EEiFV0-vyJCZCJ0LvUEI0_WvxtEXE", "Clm9d9M2CH7pZ8vm049ZlGZStQxrRQVRTjU_T5_0")
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{
		Params: map[string]string{},
	}

	defer file.Close()

	err := formUploader.Put(ctx, &ret, upToken, dst, file, int64(size), &putExtra)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(ret.Key, ret.Hash)
	return nil
}

// Delete 删除一个或多个文件，
// 返回已删除的文件，及遇到的最后一个错误
func (handler Handler) Delete(ctx context.Context, files []string) ([]string, error) {
	deleted := make([]string, 0, len(files))
	var retErr error

	for _, value := range files {
		err := os.Remove(value)
		if err == nil {
			deleted = append(deleted, value)
		} else {
			util.Log().Warning("无法删除文件，%s", err)
			retErr = err
		}
	}

	return deleted, retErr
}
