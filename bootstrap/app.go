package bootstrap

import (
	// "bytes"
	// "crypto/aes"
	// "crypto/cipher"
	// "encoding/gob"
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	// "os"
	// "strconv"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/vol"

	// "github.com/cloudreve/Cloudreve/v3/pkg/request"
	// "github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
)

var matrix []byte
var APPID string

// InitApplication 初始化应用常量
func InitApplication() {
	fmt.Print(`
   ___ _                 _                    
  / __\ | ___  _   _  __| |_ __ _____   _____ 
 / /  | |/ _ \| | | |/ _  | '__/ _ \ \ / / _ \	
/ /___| | (_) | |_| | (_| | | |  __/\ V /  __/
\____/|_|\___/ \__,_|\__,_|_|  \___| \_/ \___|

   V` + conf.BackendVersion + `  Commit #` + conf.LastCommit + `  Plus=` + conf.IsPlus + `
================================================

`)
	// data, err := ioutil.ReadFile(util.RelativePath(string([]byte{107, 101, 121, 46, 98, 105, 110})))
	// if err != nil {
	// 	util.Log().Panic("%s", err)
	// }

	//table := deSign(data)
	//constant.HashIDTable = table["table"].([]int)
	//APPID = table["id"].(string)
	//matrix = table["pic"].([]byte)
	APPID = `1145141919810`
	matrix = []byte{1, 1, 4, 5, 1, 4, 1, 9, 1, 9, 8, 1, 0}
	vol.ClientSecret = APPID
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

// func deSign(data []byte) map[string]interface{} {
// 	res := decode(data, seed())
// 	dec := gob.NewDecoder(bytes.NewReader(res))
// 	obj := map[string]interface{}{}
// 	err := dec.Decode(&obj)
// 	if err != nil {
// 		util.Log().Panic("您仍在使用旧版的授权文件，请前往 https://pro.cloudreve.org/ 登录下载最新的授权文件")
// 		os.Exit(-1)
// 	}
// 	return checkKeyUpdate(obj)
// }

// func checkKeyUpdate(table map[string]interface{}) map[string]interface{} {
// 	if table["version"].(string) != conf.KeyVersion {
// 		util.Log().Info("正在自动更新授权文件...")
// 		reqBody := map[string]string{
// 			"secret": table["secret"].(string),
// 			"id":     table["id"].(string),
// 		}
// 		reqBodyString, _ := json.Marshal(reqBody)
// 		client := request.NewClient()
// 		resp := client.Request("POST", "https://pro.cloudreve.org/Api/UpdateKey",
// 			bytes.NewReader(reqBodyString)).CheckHTTPResponse(200)
// 		if resp.Err != nil {
// 			util.Log().Panic("授权文件更新失败, %s", resp.Err)
// 		}
// 		keyContent, _ := ioutil.ReadAll(resp.Response.Body)
// 		ioutil.WriteFile(util.RelativePath(string([]byte{107, 101, 121, 46, 98, 105, 110})), keyContent, os.ModePerm)

// 		return deSign(keyContent)
// 	}

// 	return table
// }

// func seed() []byte {
// 	res := []int{8}
// 	s := "20210323"
// 	m := 1 << 20
// 	a := 9
// 	b := 7
// 	for i := 1; i < 23; i++ {
// 		res = append(res, (a*res[i-1]+b)%m)
// 		s += strconv.Itoa(res[i])
// 	}
// 	return []byte(s)
// }

// func decode(cryted []byte, key []byte) []byte {
// 	block, _ := aes.NewCipher(key[:32])
// 	blockSize := block.BlockSize()
// 	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
// 	orig := make([]byte, len(cryted))
// 	blockMode.CryptBlocks(orig, cryted)
// 	orig = pKCS7UnPadding(orig)
// 	return orig
// }

// func pKCS7UnPadding(origData []byte) []byte {
// 	length := len(origData)
// 	unpadding := int(origData[length-1])
// 	return origData[:(length - unpadding)]
// }
