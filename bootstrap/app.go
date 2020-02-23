package bootstrap

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/gob"
	"fmt"
	"github.com/HFO4/cloudreve/bootstrap/constant"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"strconv"
)

var matrix []byte
var APPID string

// InitApplication 初始化应用常量
func InitApplication() {
	data, err := ioutil.ReadFile(string([]byte{107, 101, 121, 46, 98, 105, 110}))
	if err != nil {
		util.Log().Panic("%s", err)
	}

	table := deSign(data)
	constant.HashIDTable = table["table"].([]int)
	APPID = table["id"].(string)
	matrix = table["pic"].([]byte)
}

// InitCustomRoute 初始化自定义路由
func InitCustomRoute(group *gin.RouterGroup) {
	group.GET(string([]byte{98, 103}), func(c *gin.Context) {
		c.Header("content-type", "image/png")
		c.Writer.Write(matrix)
	})
	group.GET("id", func(c *gin.Context) {
		c.String(200, APPID)
	})
}

func deSign(data []byte) map[string]interface{} {
	res := decode(data, seed())
	dec := gob.NewDecoder(bytes.NewReader(res))
	obj := map[string]interface{}{}
	err := dec.Decode(&obj)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	return obj
}

func seed() []byte {
	res := []int{10}
	s := "2020"
	m := 1 << 20
	a := 9
	b := 7
	for i := 1; i < 23; i++ {
		res = append(res, (a*res[i-1]+b)%m)
		s += strconv.Itoa(res[i])
	}
	return []byte(s)
}

func decode(cryted []byte, key []byte) []byte {
	block, _ := aes.NewCipher(key[:32])
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	orig := make([]byte, len(cryted))
	blockMode.CryptBlocks(orig, cryted)
	orig = pKCS7UnPadding(orig)
	return orig
}

func pKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
