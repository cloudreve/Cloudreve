package model

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
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
	return "https://cdn4.buysellads.net/uu/1/46074/1559075156-slack-carbon-red_2x.png"
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
func (user *User) RegisterAuthn(credential *webauthn.Credential) {
	res, err := json.Marshal([]webauthn.Credential{*credential})
	if err != nil {
		fmt.Println(err)
	}
	DB.Model(user).Update("authn", string(res))
}
