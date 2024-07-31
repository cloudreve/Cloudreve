package thumb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gofrs/uuid"
)

func init() {
	RegisterGenerator(&LibRawGenerator{})
}

type LibRawGenerator struct {
	exts        []string
	lastRawExts string
}

func (f *LibRawGenerator) Generate(ctx context.Context, file io.Reader, _ string, name string, options map[string]string) (*Result, error) {
	const (
		thumbLibRawPath = "thumb_libraw_path"
		thumbLibRawExt  = "thumb_libraw_exts"
		thumbTempPath   = "temp_path"
	)

	opts := model.GetSettingByNames(thumbLibRawPath, thumbLibRawExt, thumbTempPath)

	if f.lastRawExts != opts[thumbLibRawExt] {
		f.exts = strings.Split(opts[thumbLibRawExt], ",")
		f.lastRawExts = opts[thumbLibRawExt]
	}

	if !util.IsInExtensionList(f.exts, name) {
		return nil, fmt.Errorf("unsupported image format: %w", ErrPassThrough)
	}

	inputFilePath := filepath.Join(
		util.RelativePath(opts[thumbTempPath]),
		"thumb",
		fmt.Sprintf("thumb_%s", uuid.Must(uuid.NewV4()).String()),
	)
	defer func() { _ = os.Remove(inputFilePath) }()

	inputFile, err := util.CreatNestedFile(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err = io.Copy(inputFile, file); err != nil {
		_ = inputFile.Close()
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}
	_ = inputFile.Close()

	cmd := exec.CommandContext(ctx, opts[thumbLibRawPath], "-e", inputFilePath)

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	if err = cmd.Run(); err != nil {
		util.Log().Warning("Failed to invoke LibRaw: %s", stdErr.String())
		return nil, fmt.Errorf("failed to invoke LibRaw: %w", err)
	}

	outputFilePath := inputFilePath + ".thumb.jpg"
	defer func() { _ = os.Remove(outputFilePath) }()

	ff, err := os.Open(outputFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer func() { _ = ff.Close() }()

	// use builtin generator
	result, err := new(Builtin).Generate(ctx, ff, outputFilePath, filepath.Base(outputFilePath), options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	orientation, err := getJpegOrientation(outputFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get jpeg orientation: %w", err)
	}
	if orientation == 1 {
		return result, nil
	}

	if err = rotateImg(result.Path, orientation); err != nil {
		return nil, fmt.Errorf("failed to rotate image: %w", err)
	}
	return result, nil
}

func rotateImg(filePath string, orientation int) error {
	resultImg, err := os.OpenFile(filePath, os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	defer func() { _ = resultImg.Close() }()

	imgFlag := make([]byte, 3)
	if _, err = io.ReadFull(resultImg, imgFlag); err != nil {
		return err
	}
	if _, err = resultImg.Seek(0, 0); err != nil {
		return err
	}

	var img image.Image
	if bytes.Equal(imgFlag, []byte{0xFF, 0xD8, 0xFF}) {
		img, err = jpeg.Decode(resultImg)
	} else {
		img, err = png.Decode(resultImg)
	}
	if err != nil {
		return err
	}

	switch orientation {
	case 8:
		img = rotate90(img)
	case 3:
		img = rotate90(rotate90(img))
	case 6:
		img = rotate90(rotate90(rotate90(img)))
	case 2:
		img = mirrorImg(img)
	case 7:
		img = rotate90(mirrorImg(img))
	case 4:
		img = rotate90(rotate90(mirrorImg(img)))
	case 5:
		img = rotate90(rotate90(rotate90(mirrorImg(img))))
	}

	if err = resultImg.Truncate(0); err != nil {
		return err
	}
	if _, err = resultImg.Seek(0, 0); err != nil {
		return err
	}

	if bytes.Equal(imgFlag, []byte{0xFF, 0xD8, 0xFF}) {
		return jpeg.Encode(resultImg, img, nil)
	}
	return png.Encode(resultImg, img)
}

func getJpegOrientation(fileName string) (int, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()

	header := make([]byte, 6)
	defer func() { header = nil }()
	if _, err = io.ReadFull(f, header); err != nil {
		return 0, err
	}

	// jpeg format header
	if !bytes.Equal(header[:3], []byte{0xFF, 0xD8, 0xFF}) {
		return 0, errors.New("not a jpeg")
	}

	// not a APP1 marker
	if header[3] != 0xE1 {
		return 1, nil
	}

	// exif data total length
	totalLen := int(header[4])<<8 + int(header[5]) - 2
	buf := make([]byte, totalLen)
	defer func() { buf = nil }()
	if _, err = io.ReadFull(f, buf); err != nil {
		return 0, err
	}

	// remove Exif identifier code
	buf = buf[6:]

	// byte order
	parse16, parse32, err := initParseMethod(buf[:2])
	if err != nil {
		return 0, err
	}

	// version
	_ = buf[2:4]

	// first IFD offset
	offset := parse32(buf[4:8])

	// first DE offset
	offset += 2
	buf = buf[offset:]

	const (
		orientationTag = 0x112
		deEntryLength  = 12
	)
	for len(buf) > deEntryLength {
		tag := parse16(buf[:2])
		if tag == orientationTag {
			return int(parse32(buf[8:12])), nil
		}
		buf = buf[deEntryLength:]
	}

	return 0, errors.New("orientation not found")
}

func initParseMethod(buf []byte) (func([]byte) int16, func([]byte) int32, error) {
	if bytes.Equal(buf, []byte{0x49, 0x49}) {
		return littleEndian16, littleEndian32, nil
	}
	if bytes.Equal(buf, []byte{0x4D, 0x4D}) {
		return bigEndian16, bigEndian32, nil
	}
	return nil, nil, errors.New("invalid byte order")
}

func littleEndian16(buf []byte) int16 {
	return int16(buf[0]) | int16(buf[1])<<8
}

func bigEndian16(buf []byte) int16 {
	return int16(buf[1]) | int16(buf[0])<<8
}

func littleEndian32(buf []byte) int32 {
	return int32(buf[0]) | int32(buf[1])<<8 | int32(buf[2])<<16 | int32(buf[3])<<24
}

func bigEndian32(buf []byte) int32 {
	return int32(buf[3]) | int32(buf[2])<<8 | int32(buf[1])<<16 | int32(buf[0])<<24
}

func rotate90(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newImg := image.NewRGBA(image.Rect(0, 0, height, width))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			newImg.Set(y, width-x-1, img.At(x, y))
		}
	}
	return newImg
}

func mirrorImg(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			newImg.Set(width-x-1, y, img.At(x, y))
		}
	}
	return newImg
}

func (f *LibRawGenerator) Priority() int {
	return 250
}

func (f *LibRawGenerator) EnableFlag() string {
	return "thumb_libraw_enabled"
}

var _ Generator = (*LibRawGenerator)(nil)
