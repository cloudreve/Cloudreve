package onedrive

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"testing"
)

func TestDriver_replaceSourceHost(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		cdn     string
		want    string
		wantErr bool
	}{
		{"TestNoReplace", "http://1dr.ms/download.aspx?123456", "", "http://1dr.ms/download.aspx?123456", false},
		{"TestReplaceCorrect", "http://1dr.ms/download.aspx?123456", "https://test.com:8080", "https://test.com:8080/download.aspx?123456", false},
		{"TestCdnFormatError", "http://1dr.ms/download.aspx?123456", string([]byte{0x7f}), "", true},
		{"TestSrcFormatError", string([]byte{0x7f}), "https://test.com:8080", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &model.Policy{}
			policy.OptionsSerialized.OdProxy = tt.cdn
			handler := Driver{
				Policy: policy,
			}
			got, err := handler.replaceSourceHost(tt.origin)
			if (err != nil) != tt.wantErr {
				t.Errorf("replaceSourceHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("replaceSourceHost() got = %v, want %v", got, tt.want)
			}
		})
	}
}
