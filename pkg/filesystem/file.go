package filesystem

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/juju/ratelimit"
	"io"
	"strconv"
)

/* ============
	 文件相关
   ============
*/

// 限速后的ReaderSeeker
type lrs struct {
	response.RSCloser
	r io.Reader
}

func (r lrs) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

// withSpeedLimit 给原有的ReadSeeker加上限速
func (fs *FileSystem) withSpeedLimit(rs response.RSCloser) response.RSCloser {
	// 如果用户组有速度限制，就返回限制流速的ReaderSeeker
	if fs.User.Group.SpeedLimit != 0 {
		speed := fs.User.Group.SpeedLimit
		bucket := ratelimit.NewBucketWithRate(float64(speed), int64(speed))
		lrs := lrs{rs, ratelimit.Reader(rs, bucket)}
		return lrs
	}
	// 否则返回原始流
	return rs

}

// AddFile 新增文件记录
func (fs *FileSystem) AddFile(ctx context.Context, parent *model.Folder) (*model.File, error) {
	// 添加文件记录前的钩子
	err := fs.Trigger(ctx, "BeforeAddFile")
	if err != nil {
		return nil, err
	}

	file := ctx.Value(fsctx.FileHeaderCtx).(FileHeader)
	filePath := ctx.Value(fsctx.SavePathCtx).(string)

	newFile := model.File{
		Name:       file.GetFileName(),
		SourceName: filePath,
		UserID:     fs.User.ID,
		Size:       file.GetSize(),
		FolderID:   parent.ID,
		PolicyID:   fs.User.Policy.ID,
	}
	_, err = newFile.Create()

	if err != nil {
		if err := fs.Trigger(ctx, "AfterValidateFailed"); err != nil {
			util.Log().Debug("AfterValidateFailed 钩子执行失败，%s", err)
		}
		return nil, ErrFileExisted.WithError(err)
	}

	return &newFile, nil
}

// GetPhysicalFileContent 根据文件物理路径获取文件流
func (fs *FileSystem) GetPhysicalFileContent(ctx context.Context, path string) (response.RSCloser, error) {
	// 重设上传策略
	fs.Policy = &model.Policy{Type: "local"}
	_ = fs.dispatchHandler()

	// 获取文件流
	rs, err := fs.Handler.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return fs.withSpeedLimit(rs), nil
}

// GetDownloadContent 获取用于下载的文件流
func (fs *FileSystem) GetDownloadContent(ctx context.Context, path string) (response.RSCloser, error) {
	// 获取原始文件流
	rs, err := fs.GetContent(ctx, path)
	if err != nil {
		return nil, err
	}

	// 返回限速处理后的文件流
	return fs.withSpeedLimit(rs), nil

}

// GetContent 获取文件内容，path为虚拟路径
func (fs *FileSystem) GetContent(ctx context.Context, path string) (response.RSCloser, error) {
	// 触发`下载前`钩子
	err := fs.Trigger(ctx, "BeforeFileDownload")
	if err != nil {
		util.Log().Debug("BeforeFileDownload 钩子执行失败，%s", err)
		return nil, err
	}

	// 找到文件
	if len(fs.FileTarget) == 0 {
		exist, file := fs.IsFileExist(path)
		if !exist {
			return nil, ErrObjectNotExist
		}
		fs.FileTarget = []model.File{*file}
	}
	ctx = context.WithValue(ctx, fsctx.FileModelCtx, fs.FileTarget[0])

	// 将当前存储策略重设为文件使用的
	fs.Policy = fs.FileTarget[0].GetPolicy()
	err = fs.dispatchHandler()
	if err != nil {
		return nil, err
	}

	// 获取文件流
	rs, err := fs.Handler.Get(ctx, fs.FileTarget[0].SourceName)
	if err != nil {
		return nil, ErrIO.WithError(err)
	}

	return rs, nil
}

// deleteGroupedFile 对分组好的文件执行删除操作，
// 返回每个分组失败的文件列表
func (fs *FileSystem) deleteGroupedFile(ctx context.Context, files map[uint][]*model.File) map[uint][]string {
	// 失败的文件列表
	// TODO 并行删除
	failed := make(map[uint][]string, len(files))

	for policyID, toBeDeletedFiles := range files {
		// 列举出需要物理删除的文件的物理路径
		sourceNames := make([]string, 0, len(toBeDeletedFiles))
		for i := 0; i < len(toBeDeletedFiles); i++ {
			sourceNames = append(sourceNames, toBeDeletedFiles[i].SourceName)
		}

		// 切换上传策略
		fs.Policy = toBeDeletedFiles[0].GetPolicy()
		err := fs.dispatchHandler()
		if err != nil {
			failed[policyID] = sourceNames
			continue
		}

		// 执行删除
		failedFile, _ := fs.Handler.Delete(ctx, sourceNames)
		failed[policyID] = failedFile

	}

	return failed
}

// GroupFileByPolicy 将目标文件按照存储策略分组
func (fs *FileSystem) GroupFileByPolicy(ctx context.Context, files []model.File) map[uint][]*model.File {
	var policyGroup = make(map[uint][]*model.File)

	for key := range files {
		if file, ok := policyGroup[files[key].PolicyID]; ok {
			// 如果已存在分组，直接追加
			policyGroup[files[key].PolicyID] = append(file, &files[key])
		} else {
			// 分布不存在，创建
			policyGroup[files[key].PolicyID] = make([]*model.File, 0)
			policyGroup[files[key].PolicyID] = append(policyGroup[files[key].PolicyID], &files[key])
		}
	}

	return policyGroup
}

// GetDownloadURL 创建文件下载链接, timeout 为数据库中存储过期时间的字段
func (fs *FileSystem) GetDownloadURL(ctx context.Context, path string, timeout string) (string, error) {
	var fileTarget *model.File
	// 找到文件
	if len(fs.FileTarget) == 0 {
		exist, file := fs.IsFileExist(path)
		if !exist {
			return "", ErrObjectNotExist
		}
		fileTarget = file
	} else {
		fileTarget = &fs.FileTarget[0]
	}

	// 生成下載地址
	ttl, err := strconv.ParseInt(model.GetSettingByName(timeout), 10, 64)
	if err != nil {
		return "",
			serializer.NewError(
				serializer.CodeInternalSetting,
				"无法获取下载地址有效期",
				err,
			)
	}

	source, err := fs.signURL(
		ctx,
		fileTarget,
		ttl,
		true,
	)
	if err != nil {
		return "", err
	}

	return source, nil
}

// GetSource 获取可直接访问文件的外链地址
func (fs *FileSystem) GetSource(ctx context.Context, fileID uint) (string, error) {
	// 查找文件记录
	fileObject, err := model.GetFilesByIDs([]uint{fileID}, fs.User.ID)
	if err != nil || len(fileObject) == 0 {
		return "", ErrObjectNotExist.WithError(err)
	}

	// 检查存储策略是否可以获得外链
	if !fileObject[0].GetPolicy().IsOriginLinkEnable {
		return "", serializer.NewError(
			serializer.CodePolicyNotAllowed,
			"当前存储策略无法获得外链",
			nil,
		)
	}

	source, err := fs.signURL(ctx, &fileObject[0], 0, false)
	if err != nil {
		return "", serializer.NewError(serializer.CodeNotSet, "无法获取外链", err)
	}

	return source, nil
}

func (fs *FileSystem) signURL(ctx context.Context, file *model.File, ttl int64, isDownload bool) (string, error) {
	fs.FileTarget = []model.File{*file}
	ctx = context.WithValue(ctx, fsctx.FileModelCtx, *file)

	// 将当前存储策略重设为文件使用的
	fs.Policy = file.GetPolicy()
	err := fs.dispatchHandler()
	if err != nil {
		return "", err
	}

	// 签名最终URL
	// 生成外链地址
	siteURL := model.GetSiteURL()
	source, err := fs.Handler.Source(ctx, fs.FileTarget[0].SourceName, *siteURL, ttl, isDownload)
	if err != nil {
		return "", serializer.NewError(serializer.CodeNotSet, "无法获取外链", err)
	}

	return source, nil
}
