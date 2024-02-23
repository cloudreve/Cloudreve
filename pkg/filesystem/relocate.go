package filesystem

import (
	"context"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

/* ================
	 存储策略迁移
   ================
*/

// Relocate 将目标文件转移到当前存储策略下
func (fs *FileSystem) Relocate(ctx context.Context, files []model.File, policy *model.Policy) error {
	// 重设存储策略为要转移的目的策略
	fs.Policy = policy
	if err := fs.DispatchHandler(); err != nil {
		return err
	}

	// 将目前文件根据存储策略分组
	fileGroup := fs.GroupFileByPolicy(ctx, files)

	// 按照存储策略分组处理每个文件
	for _, fileList := range fileGroup {
		// 如果存储策略一样，则跳过
		if fileList[0].GetPolicy().ID == fs.Policy.ID {
			util.Log().Debug("Skip relocating %d file(s), since they are already in desired policy.",
				len(fileList))
			continue
		}

		// 获取当前存储策略的处理器
		currentPolicy, _ := model.GetPolicyByID(fileList[0].PolicyID)
		currentHandler, err := getNewPolicyHandler(&currentPolicy)
		if err != nil {
			return err
		}

		// 记录转移完毕需要删除的文件
		toBeDeleted := make([]model.File, 0, len(fileList))

		// 循环处理每一个文件
		// for id, r := 0, len(fileList); id < r; id++ {
		for id, _ := range fileList {
			// 验证文件是否符合新存储策略的规定
			if err := HookValidateFile(ctx, fs, fileList[id]); err != nil {
				util.Log().Debug("File %q failed to pass validators in new policy %q, skipping.",
					fileList[id].Name, err)
				continue
			}

			// 为文件生成新存储策略下的物理路径
			savePath := fs.GenerateSavePath(ctx, fileList[id])

			// 获取原始文件
			src, err := currentHandler.Get(ctx, fileList[id].SourceName)
			if err != nil {
				util.Log().Debug("Failed to get file %q: %s, skipping.",
					fileList[id].Name, err)
				continue
			}

			// 转存到新的存储策略
			if err := fs.Handler.Put(ctx, &fsctx.FileStream{
				File:     src,
				SavePath: savePath,
				Size:     fileList[id].Size,
			}); err != nil {
				util.Log().Debug("Failed to migrate file %q: %s, skipping.",
					fileList[id].Name, err)
				continue
			}

			toBeDeleted = append(toBeDeleted, *fileList[id])

			// 更新文件信息
			fileList[id].Relocate(savePath, fs.Policy.ID)
		}

		// 排除带有软链接的文件
		toBeDeletedClean, err := model.RemoveFilesWithSoftLinks(toBeDeleted)
		if err != nil {
			util.Log().Warning("Failed to check soft links: %s", err)
		}

		deleteSourceNames := make([]string, 0, len(toBeDeleted))
		for i := 0; i < len(toBeDeletedClean); i++ {
			deleteSourceNames = append(deleteSourceNames, toBeDeletedClean[i].SourceName)
		}

		// 删除原始策略中的文件
		if _, err := currentHandler.Delete(ctx, deleteSourceNames); err != nil {
			util.Log().Warning("Cannot delete files in origin policy after relocating: %s", err)
		}
	}

	return nil
}
