package filesystem

import (
	"context"
	"io"
)

/* ==============
     压缩/解压缩
   ==============
*/

// Compress 创建给定目录和文件的压缩文件
func (fs *FileSystem) Compress(ctx context.Context, dirs, files []uint) (io.ReadSeeker, error) {
	return nil, nil
}
