package scripts

import "github.com/cloudreve/Cloudreve/v3/models/scripts/invoker"

func Init() {
	invoker.Register("ResetAdminPassword", ResetAdminPassword(0))
	invoker.Register("CalibrateUserStorage", UserStorageCalibration(0))
	invoker.Register("UpgradeTo3.4.0", UpgradeTo340(0))
}
