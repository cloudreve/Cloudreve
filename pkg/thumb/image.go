package thumb

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

// Thumb 缩略图
type Thumb struct {
	src image.Image
	ext string
}

// NewThumbFromFile 从文件数据获取新的Thumb对象，
// 尝试通过文件名name解码图像
func NewThumbFromFile(file io.ReadSeeker, name string) (*Thumb, error) {
	ext := strings.ToLower(filepath.Ext(name))
	// 无扩展名时
	if len(ext) == 0 {
		return nil, errors.New("未知的图像类型")
	}

	var err error
	var img image.Image
	switch ext[1:] {
	case "jpg":
		img, err = jpeg.Decode(file)
	case "jpeg":
		img, err = jpeg.Decode(file)
	case "gif":
		img, err = gif.Decode(file)
	case "png":
		img, err = png.Decode(file)
	default:
		return nil, errors.New("未知的图像类型")
	}
	if err != nil {
		return nil, err
	}

	return &Thumb{
		src: img,
		ext: ext[1:],
	}, nil
}

// GetThumb 生成给定最大尺寸的缩略图
func (image *Thumb) GetThumb(width, height uint) {
	image.src = resize.Thumbnail(width, height, image.src, resize.Lanczos3)
}

// GetSize 获取图像尺寸
func (image *Thumb) GetSize() (int, int) {
	b := image.src.Bounds()
	return b.Max.X, b.Max.Y
}

// Save 保存图像到给定路径
func (image *Thumb) Save(path string) (err error) {
	out, err := os.Create(path)
	defer out.Close()

	if err != nil {
		return err
	}

	err = jpeg.Encode(out, image.src, nil)
	return err

}
