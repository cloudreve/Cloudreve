package thumb

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"testing"

	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/stretchr/testify/assert"
)

func CreateTestImage() *os.File {
	file, err := os.Create("TestNewThumbFromFile.jpeg")
	alpha := image.NewAlpha(image.Rect(0, 0, 500, 200))
	jpeg.Encode(file, alpha, nil)
	if err != nil {
		fmt.Println(err)
	}
	_, _ = file.Seek(0, 0)
	return file
}

func TestNewThumbFromFile(t *testing.T) {
	asserts := assert.New(t)
	file := CreateTestImage()
	defer file.Close()

	// 无扩展名时
	{
		thumb, err := NewThumbFromFile(file, "123")
		asserts.Error(err)
		asserts.Nil(thumb)
	}

	{
		thumb, err := NewThumbFromFile(file, "123.jpg")
		asserts.NoError(err)
		asserts.NotNil(thumb)
	}
	{
		thumb, err := NewThumbFromFile(file, "123.jpeg")
		asserts.Error(err)
		asserts.Nil(thumb)
	}
	{
		thumb, err := NewThumbFromFile(file, "123.png")
		asserts.Error(err)
		asserts.Nil(thumb)
	}
	{
		thumb, err := NewThumbFromFile(file, "123.gif")
		asserts.Error(err)
		asserts.Nil(thumb)
	}
	{
		thumb, err := NewThumbFromFile(file, "123.3211")
		asserts.Error(err)
		asserts.Nil(thumb)
	}
}

func TestThumb_GetSize(t *testing.T) {
	asserts := assert.New(t)
	file := CreateTestImage()
	defer file.Close()
	thumb, err := NewThumbFromFile(file, "123.jpg")
	asserts.NoError(err)

	w, h := thumb.GetSize()
	asserts.Equal(500, w)
	asserts.Equal(200, h)
}

func TestThumb_GetThumb(t *testing.T) {
	asserts := assert.New(t)
	file := CreateTestImage()
	defer file.Close()
	thumb, err := NewThumbFromFile(file, "123.jpg")
	asserts.NoError(err)

	asserts.NotPanics(func() {
		thumb.GetThumb(10, 10)
	})
}

func TestThumb_Thumbnail(t *testing.T) {
	asserts := assert.New(t)
	{
		img := image.NewRGBA(image.Rect(0, 0, 500, 200))
		thumb := Thumbnail(100, 100, img)
		asserts.Equal(thumb.Bounds(), image.Rect(0, 0, 100, 40))
	}
	{
		img := image.NewRGBA(image.Rect(0, 0, 200, 200))
		thumb := Thumbnail(100, 100, img)
		asserts.Equal(thumb.Bounds(), image.Rect(0, 0, 100, 100))
	}
	{
		img := image.NewRGBA(image.Rect(0, 0, 500, 500))
		thumb := Thumbnail(100, 100, img)
		asserts.Equal(thumb.Bounds(), image.Rect(0, 0, 100, 100))
	}
	{
		img := image.NewRGBA(image.Rect(0, 0, 200, 500))
		thumb := Thumbnail(100, 100, img)
		asserts.Equal(thumb.Bounds(), image.Rect(0, 0, 40, 100))
	}
}

func TestThumb_Save(t *testing.T) {
	asserts := assert.New(t)
	file := CreateTestImage()
	defer file.Close()
	thumb, err := NewThumbFromFile(file, "123.jpg")
	asserts.NoError(err)

	err = thumb.Save("/:noteexist/")
	asserts.Error(err)

	err = thumb.Save("TestThumb_Save.png")
	asserts.NoError(err)
	asserts.True(util.Exists("TestThumb_Save.png"))

}

func TestThumb_CreateAvatar(t *testing.T) {
	asserts := assert.New(t)
	file := CreateTestImage()
	defer file.Close()

	thumb, err := NewThumbFromFile(file, "123.jpg")
	asserts.NoError(err)

	cache.Set("setting_avatar_path", "tests", 0)
	cache.Set("setting_avatar_size_s", "50", 0)
	cache.Set("setting_avatar_size_m", "130", 0)
	cache.Set("setting_avatar_size_l", "200", 0)

	asserts.NoError(thumb.CreateAvatar(1))
	asserts.True(util.Exists(util.RelativePath("tests/avatar_1_1.png")))
	asserts.True(util.Exists(util.RelativePath("tests/avatar_1_2.png")))
	asserts.True(util.Exists(util.RelativePath("tests/avatar_1_0.png")))
}
