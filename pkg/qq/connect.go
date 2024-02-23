package qq

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gofrs/uuid"
	"net/url"
	"strings"
)

// LoginPage 登陆页面描述
type LoginPage struct {
	URL       string
	SecretKey string
}

// UserCredentials 登陆成功后的凭证
type UserCredentials struct {
	OpenID      string
	AccessToken string
}

// UserInfo 用户信息
type UserInfo struct {
	Nick   string
	Avatar string
}

var (
	// ErrNotEnabled 未开启登录功能
	ErrNotEnabled = serializer.NewError(serializer.CodeFeatureNotEnabled, "QQ Login is not enabled", nil)
	// ErrObtainAccessToken 无法获取AccessToken
	ErrObtainAccessToken = serializer.NewError(serializer.CodeNotSet, "Cannot obtain AccessToken", nil)
	// ErrObtainOpenID 无法获取OpenID
	ErrObtainOpenID = serializer.NewError(serializer.CodeNotSet, "Cannot obtain OpenID", nil)
	//ErrDecodeResponse 无法解析服务端响应
	ErrDecodeResponse = serializer.NewError(serializer.CodeNotSet, "Cannot parse serverside response", nil)
)

// NewLoginRequest 新建登录会话
func NewLoginRequest() (*LoginPage, error) {
	// 获取相关设定
	options := model.GetSettingByNames("qq_login", "qq_login_id")
	if options["qq_login"] == "0" {
		return nil, ErrNotEnabled
	}

	// 生成唯一ID
	u2, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	secret := fmt.Sprintf("%x", md5.Sum(u2.Bytes()))

	// 生成登录地址
	loginURL, _ := url.Parse("https://graph.qq.com/oauth2.0/authorize?response_type=code")
	queries := loginURL.Query()
	queries.Add("client_id", options["qq_login_id"])
	queries.Add("redirect_uri", getCallbackURL())
	queries.Add("state", secret)
	loginURL.RawQuery = queries.Encode()

	return &LoginPage{
		URL:       loginURL.String(),
		SecretKey: secret,
	}, nil
}

func getCallbackURL() string {
	//return "https://drive.aoaoao.me/Callback/QQ"
	// 生成回调地址
	gateway, _ := url.Parse("/login/qq")
	callback := model.GetSiteURL().ResolveReference(gateway).String()

	return callback
}

func getAccessTokenURL(code string) string {
	// 获取相关设定
	options := model.GetSettingByNames("qq_login_id", "qq_login_key")

	api, _ := url.Parse("https://graph.qq.com/oauth2.0/token?grant_type=authorization_code")
	queries := api.Query()
	queries.Add("client_id", options["qq_login_id"])
	queries.Add("redirect_uri", getCallbackURL())
	queries.Add("client_secret", options["qq_login_key"])
	queries.Add("code", code)
	api.RawQuery = queries.Encode()

	return api.String()
}

func getUserInfoURL(openid, ak string) string {
	// 获取相关设定
	options := model.GetSettingByNames("qq_login_id", "qq_login_key")

	api, _ := url.Parse("https://graph.qq.com/user/get_user_info")
	queries := api.Query()
	queries.Add("oauth_consumer_key", options["qq_login_id"])
	queries.Add("openid", openid)
	queries.Add("access_token", ak)
	api.RawQuery = queries.Encode()

	return api.String()
}

func getResponse(body string) (map[string]interface{}, error) {
	var res map[string]interface{}

	if !strings.Contains(body, "callback") {
		return res, nil
	}

	body = strings.TrimPrefix(body, "callback(")
	body = strings.TrimSuffix(body, ");\n")

	err := json.Unmarshal([]byte(body), &res)

	return res, err
}

// Callback 处理回调,返回openid和access key
func Callback(code string) (*UserCredentials, error) {
	// 获取相关设定
	options := model.GetSettingByNames("qq_login")
	if options["qq_login"] == "0" {
		return nil, ErrNotEnabled
	}

	api := getAccessTokenURL(code)

	// 获取AccessToken
	client := request.NewClient()
	res := client.Request("GET", api, nil)
	resp, err := res.GetResponse()
	if err != nil {
		return nil, ErrObtainAccessToken.WithError(err)
	}

	// 如果服务端返回错误
	errResp, err := getResponse(resp)
	if msg, ok := errResp["error_description"]; err == nil && ok {
		return nil, ErrObtainAccessToken.WithError(errors.New(msg.(string)))
	}

	// 获取AccessToken
	vals, err := url.ParseQuery(resp)
	if err != nil {
		return nil, ErrDecodeResponse.WithError(err)
	}
	accessToken := vals.Get("access_token")

	// 用 AccessToken 换取OpenID
	res = client.Request("GET", "https://graph.qq.com/oauth2.0/me?access_token="+accessToken, nil)
	resp, err = res.GetResponse()
	if err != nil {
		return nil, ErrObtainOpenID.WithError(err)
	}

	// 解析服务端响应
	errResp, err = getResponse(resp)
	if msg, ok := errResp["error_description"]; err == nil && ok {
		return nil, ErrObtainOpenID.WithError(errors.New(msg.(string)))
	}

	if openid, ok := errResp["openid"]; ok {
		return &UserCredentials{
			OpenID:      openid.(string),
			AccessToken: accessToken,
		}, nil
	}

	return nil, ErrDecodeResponse
}

// GetUserInfo 使用凭证获取用户信息
func GetUserInfo(credential *UserCredentials) (*UserInfo, error) {
	api := getUserInfoURL(credential.OpenID, credential.AccessToken)

	// 获取用户信息
	client := request.NewClient()
	res := client.Request("GET", api, nil)
	resp, err := res.GetResponse()
	if err != nil {
		return nil, ErrObtainAccessToken.WithError(err)
	}

	var resSerialized map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &resSerialized); err != nil {
		return nil, ErrDecodeResponse.WithError(err)
	}

	// 如果服务端返回错误
	if msg, ok := resSerialized["msg"]; ok && msg.(string) != "" {
		return nil, ErrObtainAccessToken.WithError(errors.New(msg.(string)))
	}

	if avatar, ok := resSerialized["figureurl_qq_2"]; ok {
		return &UserInfo{
			Nick:   resSerialized["nickname"].(string),
			Avatar: avatar.(string),
		}, nil
	}

	return nil, ErrDecodeResponse
}
