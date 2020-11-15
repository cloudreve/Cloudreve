package model

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/duo-labs/webauthn/webauthn"
)

/*
	`webauthn.User` 接口的实现
*/

// WebAuthnID 返回用户ID
func (user User) WebAuthnID() []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(user.ID))
	return bs
}

// WebAuthnName 返回用户名
func (user User) WebAuthnName() string {
	return user.Email
}

// WebAuthnDisplayName 获得用于展示的用户名
func (user User) WebAuthnDisplayName() string {
	return user.Nick
}

// WebAuthnIcon 获得用户头像
func (user User) WebAuthnIcon() string {
	avatar, _ := url.Parse("/api/v3/user/avatar/" + hashid.HashID(user.ID, hashid.UserID) + "/l")
	base := GetSiteURL()
	base.Scheme = "https"
	return base.ResolveReference(avatar).String()
}

// WebAuthnCredentials 获得已注册的验证器凭证
func (user User) WebAuthnCredentials() []webauthn.Credential {
	var res []webauthn.Credential
	err := json.Unmarshal([]byte(user.Authn), &res)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

// RegisterAuthn 添加新的验证器
func (user *User) RegisterAuthn(credential *webauthn.Credential) error {
	exists := user.WebAuthnCredentials()
	exists = append(exists, *credential)
	res, err := json.Marshal(exists)
	if err != nil {
		return err
	}

	return DB.Model(user).Update("authn", string(res)).Error
}

// RemoveAuthn 删除验证器
func (user *User) RemoveAuthn(id string) {
	exists := user.WebAuthnCredentials()
	for i := 0; i < len(exists); i++ {
		idEncoded := base64.StdEncoding.EncodeToString(exists[i].ID)
		if idEncoded == id {
			exists[len(exists)-1], exists[i] = exists[i], exists[len(exists)-1]
			exists = exists[:len(exists)-1]
			break
		}
	}

	res, _ := json.Marshal(exists)
	DB.Model(user).Update("authn", string(res))
}
