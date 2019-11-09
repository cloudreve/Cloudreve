package conf

import "io/ioutil"

// 当前后端版本号
const BackendVersion = string("3.0.0-b")

// WriteVersionLock 将当前版本信息写入 version.lock
func WriteVersionLock() error {
	err := ioutil.WriteFile("version.lock", []byte(BackendVersion), 0644)
	return err
}
