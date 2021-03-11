package crontab

import (
	"net/http"
	"encoding/json"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
)

type AddressJson struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Result    struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
		Address            string `json:"address"`
		FormattedAddresses struct {
			Recommend string `json:"recommend"`
			Rough     string `json:"rough"`
		} `json:"formatted_addresses"`
		AddressComponent struct {
			Nation       string `json:"nation"`
			Province     string `json:"province"`
			City         string `json:"city"`
			District     string `json:"district"`
			Street       string `json:"street"`
			StreetNumber string `json:"street_number"`
		} `json:"address_component"`
	} `json:"result"`
}

type AddressExif struct{
	Address      string `json:"address"`
	Nation       string `json:"nation"`
	Province     string `json:"province"`
	City         string `json:"city"`
	District     string `json:"district"`
	Street       string `json:"street"`
	StreetNumber string `json:"street_number"`
}

func syncPhotoAddress() {
	// 同步照片的经纬度为文本地址
	syncPhotoLatLongToAddress()

	util.Log().Info("定时任务 [cron_sync_photo_lat_long_to_address] 执行完毕")
}

func syncPhotoLatLongToAddress() {
	page := 1
	pageSize := 10
	for true{
		files , _ := model.GetEmptyLocationFilesByPage(uint(page), uint(pageSize))
		if len(files) <= 0 {
			break
		}
		for i := 0; i < len(files); i++{
			file := files[i]
			util.Log().Debug("file name: %s",file.Name)
			util.Log().Debug("file ExifLatLong: %s",file.ExifLatLong)
			util.Log().Debug("file ExifAddress: %s",file.ExifAddress)

			// 获取文件数据流
			url := "https://apis.map.qq.com/ws/geocoder/v1/?location="+ file.ExifLatLong +"&get_poi=1&key=OB4BZ-D4W3U-B7VVO-4PJWW-6TKDJ-WPB77"
			client := request.HTTPClient{}
			resp := client.Request(
				"GET",
				url,
				nil,
				request.WithHeader(
					http.Header{"Referer": {"https://lbs.qq.com/"}},
				),
			)

			respString, err := resp.GetResponse()
			if err != nil{
				util.Log().Warning("response error: %s",err)
			}

			var addressJson AddressJson
			err = json.Unmarshal([]byte(respString), &addressJson)
			if err != nil {
				util.Log().Warning("解析经纬度结果错误原始文本: %s",respString)
				util.Log().Warning("解析经纬度结果错误: %s",err)
				continue
			}
			var addressInfo AddressExif
			addressInfo.Address = addressJson.Result.Address
			addressInfo.StreetNumber = addressJson.Result.AddressComponent.StreetNumber
			addressInfo.District = addressJson.Result.AddressComponent.District
			addressInfo.City = addressJson.Result.AddressComponent.City
			addressInfo.Province = addressJson.Result.AddressComponent.Province
			addressInfo.Nation = addressJson.Result.AddressComponent.Nation

			if addressInfoStr, err := json.Marshal(addressInfo); err == nil {
				file.UpdatePicExifAddress(string(addressInfoStr))
				util.Log().Debug("解析经纬度结果重组: %s",string(addressInfoStr))
			}
		}

		page++
	}
}
