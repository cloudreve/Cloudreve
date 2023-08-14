package thumb

import (
	"context"
	"fmt"
		"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gofrs/uuid"
	//"github.com/nfnt/resize"
	"golang.org/x/image/draw"
)

func init() {
	RegisterGenerator(&Builtin{})
}

// Thumb 缩略图
type Thumb struct {
	src image.Image
	ext string
}

// NewThumbFromFile 从文件数据获取新的Thumb对象，
// 尝试通过文件名name解码图像
func NewThumbFromFile(file io.Reader, name string) (*Thumb, error) {
	ext := strings.ToLower(filepath.Ext(name))
	// 无扩展名时
	if len(ext) == 0 {
		return nil, fmt.Errorf("unknown image format: %w", ErrPassThrough)
	}

	var err error
	var img image.Image
	switch ext[1:] {
	case "jpg", "jpeg":
		img, err = jpeg.Decode(file)
	case "gif":
		img, err = gif.Decode(file)
	case "png":
		img, err = png.Decode(file)
	default:
		return nil, fmt.Errorf("unknown image format: %w", ErrPassThrough)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse image: %w (%w)", err, ErrPassThrough)
	}

	return &Thumb{
		src: img,
		ext: ext[1:],
	}, nil
}

// GetThumb 生成给定最大尺寸的缩略图
func (image *Thumb) GetThumb(width, height uint) {
	//image.src = resize.Thumbnail(width, height, image.src, resize.Lanczos3)
	image.src = Thumbnail(width, height, image.src)
}

// GetSize 获取图像尺寸
func (image *Thumb) GetSize() (int, int) {
	b := image.src.Bounds()
	return b.Max.X, b.Max.Y
}

// Save 保存图像到给定路径
func (image *Thumb) Save(w io.Writer) (err error) {
	switch model.GetSettingByNameWithDefault("thumb_encode_method", "jpg") {
	case "png":
		err = png.Encode(w, image.src)
	default:
		err = jpeg.Encode(w, image.src, &jpeg.Options{Quality: model.GetIntSetting("thumb_encode_quality", 85)})
	}

	return err

}

// Thumbnail will downscale provided image to max width and height preserving
// original aspect ratio and using the interpolation function interp.
// It will return original image, without processing it, if original sizes
// are already smaller than provided constraints.
func Thumbnail(maxWidth, maxHeight uint, img image.Image) image.Image {
	origBounds := img.Bounds()
	origWidth := uint(origBounds.Dx())
	origHeight := uint(origBounds.Dy())
	newWidth, newHeight := origWidth, origHeight

	// Return original image if it have same or smaller size as constraints
	if maxWidth >= origWidth && maxHeight >= origHeight {
		return img
	}

	// Preserve aspect ratio
	if origWidth > maxWidth {
		newHeight = uint(origHeight * maxWidth / origWidth)
		if newHeight < 1 {
			newHeight = 1
		}
		newWidth = maxWidth
	}

	if newHeight > maxHeight {
		newWidth = uint(newWidth * maxHeight / newHeight)
		if newWidth < 1 {
			newWidth = 1
		}
		newHeight = maxHeight
	}
	return Resize(newWidth, newHeight, img)
}

func Resize(newWidth, newHeight uint, img image.Image) image.Image {
	// Set the expected size that you want:
	dst := image.NewRGBA(image.Rect(0, 0, int(newWidth), int(newHeight)))
	// Resize:
	draw.BiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Src, nil)
	return dst
}

// CreateAvatar 创建头像
func (image *Thumb) CreateAvatar(uid uint) error {
	// 读取头像相关设定
	savePath := util.RelativePath(model.GetSettingByName("avatar_path"))
	s := model.GetIntSetting("avatar_size_s", 50)
	m := model.GetIntSetting("avatar_size_m", 130)
	l := model.GetIntSetting("avatar_size_l", 200)

	// 生成头像缩略图
	src := image.src
	for k, size := range []int{s, m, l} {
		out, err := util.CreatNestedFile(filepath.Join(savePath, fmt.Sprintf("avatar_%d_%d.png", uid, k)))

		if err != nil {
			return err
		}
		defer out.Close()

		image.src = Resize(uint(size), uint(size), src)
		err = image.Save(out)
		if err != nil {
			return err
		}
	}

	return nil

}

type Builtin struct{}

func (b Builtin) Generate(ctx context.Context, file io.Reader, src, url, name string, options map[string]string) (*Result, error) {
	img, err := NewThumbFromFile(file, name)
	if err != nil {
		return nil, err
	}

	img.GetThumb(thumbSize(options))
	tempPath := filepath.Join(
		util.RelativePath(model.GetSettingByName("temp_path")),
		"thumb",
		fmt.Sprintf("thumb_%s", uuid.Must(uuid.NewV4()).String()),
	)

	thumbFile, err := util.CreatNestedFile(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	defer thumbFile.Close()
	if err := img.Save(thumbFile); err != nil {
		return nil, err
	}

	return &Result{Path: tempPath}, nil
}

func (b Builtin) Priority() int {
	return 300
}

func (b Builtin) EnableFlag() string {
	return "thumb_builtin_enabled"
}
