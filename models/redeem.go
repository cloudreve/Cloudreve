package model

import "github.com/jinzhu/gorm"

// Redeem 兑换码
type Redeem struct {
	gorm.Model
	Type      int    // 订单类型
	ProductID int64  // 商品ID
	Num       int    // 商品数量
	Code      string `gorm:"size:64,index:redeem_code"` // 兑换码
	Used      bool   // 是否已被使用
}

// GetAvailableRedeem 根据code查找可用兑换码
func GetAvailableRedeem(code string) (*Redeem, error) {
	redeem := &Redeem{}
	result := DB.Where("code = ? and used = ?", code, false).First(redeem)
	return redeem, result.Error
}

// Use 设定为已使用状态
func (redeem *Redeem) Use() {
	DB.Model(redeem).Updates(map[string]interface{}{
		"used": true,
	})
}
