# 公布CloudrevePro去除授权检测方式

（昨天作者突然诈尸，论坛的帖子被隐藏了）

## 原因
+ 进来先喊一句：~~**哔~**~~
+ 还是因为Cloudreve只顾Pro版圈钱，不管issues提议，一堆Bug不修
+ 论坛提出来你当没看见，修好了发Github你举报，用户的心都被你们给伤透了
+ 所以正值v4即将发布之际，公布v3完整版"解锁"方式

## 注意
+ 本教程是给略懂编程的人看的，已经尽可能详细的讲解了操作步骤
+ 如果还是看不懂说明你不适合搞这个，请不要在评论区喷教程不好

## 后端
+ 众所周知，捐助版会检测授权文件 `key.bin`，没有它是连程序都打不开的
+ 那有人说了，在 `app.go` 的 `InitApplication` 函数里 删掉就可以了
+ 开发者能让你这么简单就破开吗，试过之后发现还是打不开程序
+ 他说的对，但不完全对，猫腻就藏在程序的依赖库里
+ 仔细看这个库 https://github.com/abslant/gzip/blob/v0.0.9/handler.go#L60
+ 看似只是一个fork版，但会在前端main.xxx.chunk.js中插入跳转官网403的代码
+ 作者的用户名为 `abslant`，乍一看不认识
+ 打开这个博客 https://hfo4.github.io/ ，注意头像下的联系邮箱，发现这就是开发者 `Aaron` 的小名
+ 这一切就说得通了，都是作者搞的鬼
+ ~~看过社区版源码的都知道，没看过的等你尝试用git对比整个仓库的时候就知道了~~
+ 首先将被加料的依赖项替换为原版
+ `github.com/abslant/mime => github.com/HFO4/aliyun-oss-go-sdk`
+ `github.com/abslant/gzip => github.com/gin-contrib/gzip`
+ VSC编辑器全局搜索，直接替换
+ `bootstrap/app.go` 不用多说，那个读取 `[]byte{107, 101, 121, 46, 98, 105, 110}` 的就是授权文件
+ `routers/router.go` 第128行 `r.Use(gzip.GzipHandler())` 改为 `r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/api/"})))`
+ 如果改完还是自动引入就把 `go.sum` 删了
+ 然后是一些小变动：
+ `pkg/hashid/hash.go` 最后一个函数 `constant.HashIDTable[t]` 改为 `t`
+ 基本上到这里就完成了
+ **注意前端打包时要保持目录结构 `assets.zip/assets/build/{前端文件}`**

## 前端
+ 忙活了半天，终于把程序跑起来了，打开页面一看，好家伙 **Backend not running**
+ 还是进不去，怎么想都进不去，因为前端还有一层验证
+ 但注意 **"任何前端加密和混淆都是纸老虎，自己玩玩无所谓，重要业务千万别乱来"**
+ 前端验证很好破解，还是先检查依赖项，打开 `package.json`
+ 头两行就是这个万恶的 `abslant`，删掉 `"@abslant/cd-image-loader"` 和 `"@abslant/cd-js-injector"`
+ 然后把引用它们的地方删掉就行...了 吗 ?
+ 位置在 `config/webpack.config.js:35_625` 和 `src/component/FileManager/FileManager.js:16_109`
+ 之后进是能进网盘了，但你想测试上传一个文件的时候就傻眼了，明明什么也没动，就是传不上去
+ 报错 `Cannot read properties of null (reading 'code')`，那是继3.5.3之后新增的一处验证
+ 将 `src/component/Uploader/core/utils/request.ts` 第12行整个 const 替换为以下内容即可解决
```js
const baseConfig = {
    transformResponse: [
        (response: any) => {
            try {
                return JSON.parse(response);
            } catch (e) {
                throw new TransformResponseError(response, e);
            }
        },
    ],
};
```
+ 最后就可以享受完整版带来的全新体验了 🎉

## 其它
+ 除了去除验证，Plus版本还增加了几处功能优化，修复遗留Bug，感兴趣的可以下载体验一下
+ 但因为是3.8.3泄露版和主线版拼凑而来的，存在不稳定因素，建议不要用于生产环境
+ 如果怕我在里面加料，可以自行检查源码，这程序十分的珍贵，尽快下载存档
+ 主地址 ↓
+ [cloudreveplus-windows-amd64v2.zip](https://github.com/cloudreve/Cloudreve/files/14327258/cloudreveplus-windows-amd64v2.zip)
+ [cloudreveplus-linux-amd64v2.zip](https://github.com/cloudreve/Cloudreve/files/14327249/cloudreveplus-linux-amd64v2.zip)
+ [cloudreveplus-linux-arm7.zip](https://github.com/cloudreve/Cloudreve/files/14327254/cloudreveplus-linux-arm7.zip)
+ [cloudreveplus-source-nogit.zip](https://github.com/cloudreve/Cloudreve/files/14327256/cloudreveplus-source-nogit.zip)
+ 备用地址 ↓ (以图片方式上传可以分到aws的地址，比githubusercontent快一些，但要分卷手动改名)
+ [cloudreveplus-source-nogit.zip](https://github.com/cloudreve/frontend/assets/100983035/4fe3ae36-275d-41e9-89fe-2a746f512bde)
+ [cloudreveplus-linux-amd64v2.001](https://github.com/cloudreve/frontend/assets/100983035/71dab1b8-8a01-4609-bf1d-ab8f6c5df57d)
+ [cloudreveplus-linux-amd64v2.002](https://github.com/cloudreve/frontend/assets/100983035/423cb9cb-9dae-47e9-baf3-43a48202fe06)
+ [cloudreveplus-linux-arm7.001](https://github.com/cloudreve/frontend/assets/100983035/a03f6c72-3ee8-44f4-96ed-ca385bc87c5c)
+ [cloudreveplus-linux-arm7.002](https://github.com/cloudreve/frontend/assets/100983035/e3f9a73d-9019-4c60-a41b-b53a9184aad9)
+ [cloudreveplus-windows-amd64v2.001](https://github.com/cloudreve/frontend/assets/100983035/a6d68487-3f40-4f6c-9cab-857d4128fb7d)
+ [cloudreveplus-windows-amd64v2.002](https://github.com/cloudreve/frontend/assets/100983035/c3620b29-8ced-4aa7-a02b-8d14c0bf4815)

......
等等，你是不是忘了什么？
📢 ~~**只可意会不可言传**~~

Cloudream云梦云盘
