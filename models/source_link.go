package model

import (
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/jinzhu/gorm"
	"net/url"
)

// SourceLink represent a shared file source link
type SourceLink struct {
	gorm.Model
	FileID    uint   // corresponding file ID
	Name      string // name of the file while creating the source link, for annotation
	Downloads int    // 下载数

	// 关联模型
	File File `gorm:"save_associations:false:false"`
}

// Link gets the URL of a SourceLink
func (s *SourceLink) Link() (string, error) {
	baseURL := GetSiteURL()
	linkPath, err := url.Parse(fmt.Sprintf("/f/%s/%s", hashid.HashID(s.ID, hashid.SourceLinkID), s.File.Name))
	if err != nil {
		return "", err
	}
	return baseURL.ResolveReference(linkPath).String(), nil
}

// GetTasksByID queries source link based on ID
func GetSourceLinkByID(id interface{}) (*SourceLink, error) {
	link := &SourceLink{}
	result := DB.Where("id = ?", id).First(link)
	files, _ := GetFilesByIDs([]uint{link.FileID}, 0)
	if len(files) > 0 {
		link.File = files[0]
	}

	return link, result.Error
}

// Viewed 增加访问次数
func (s *SourceLink) Downloaded() {
	s.Downloads++
	DB.Model(s).UpdateColumn("downloads", gorm.Expr("downloads + ?", 1))
}
