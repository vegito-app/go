package firebase

import (
	"os"

	"github.com/spf13/viper"
)

var config = viper.New()

const (
	adminSDKserviceAccountIDConfig = "adminsdk_serviceaccount_id"
	projectIDconfig                = "project_id"
	pubSubEmulatorHostConfig       = "pubsub_emulator_host"
)

func init() {
	config.SetEnvPrefix("firebase")
	config.AutomaticEnv()
	if pubSubEmulatorHost := config.GetString(pubSubEmulatorHostConfig); pubSubEmulatorHost != "" {
		os.Setenv("PUBSUB_EMULATOR_HOST", pubSubEmulatorHost)
	}
}
