package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Magic-Pod/magic-pod-api-client/common"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/bitrise-tools/go-steputils/tools"
	"github.com/urfave/cli"
)

// Config : Configuration for this step
type Config struct {
	BaseURL            string          `env:"base_url,required"`
	APIToken           stepconf.Secret `env:"magic_pod_api_token,required"`
	OrganizationName   string          `env:"organization_name,required"`
	ProjectName        string          `env:"project_name,required"`
	AppPath            string          `env:"app_path"`
	TestSettingsNumber int             `env:"test_settings_number"`
	TestSettings	   string	       `env:"test_settings"`
	WaitForResult      bool            `env:"wait_for_result"`
	DeleteAppAfterTest string          `env:"delete_app_after_test"`
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func uploadAppFile(cfg Config) int {
	appPath := cfg.AppPath
	log.Infof("Upload app file %s to Magic Pod cloud", appPath)
	fileNo, err := common.UploadApp(cfg.BaseURL, string(cfg.APIToken), cfg.OrganizationName, cfg.ProjectName, make(map[string]string), appPath)
	if err != nil {
		failf(err.Error())
	}
	log.Infof("Done. File number = %d\n", fileNo)
	return fileNo
}

func deleteAppFile(cfg Config, appFileNumber int) {
	log.Infof("Delete app file on the server")
	err := common.DeleteApp(cfg.BaseURL, string(cfg.APIToken), cfg.OrganizationName, cfg.ProjectName, make(map[string]string), appFileNumber)
	if err != nil {
		// don't exit here
		log.Errorf(err.Error())
	}
}

func startBatchRun(cfg Config, appFileNumber int) ([]common.BatchRun, *cli.ExitError) {
	log.Infof("test settings number = %d", cfg.TestSettingsNumber)
	return common.StartBatchRun(cfg.BaseURL, string(cfg.APIToken), cfg.OrganizationName, cfg.ProjectName, make(map[string]string), cfg.TestSettingsNumber, "")
}

func getBatchRun(cfg Config, batchRunNumber int) *common.BatchRun {
	batchRun, err := common.GetBatchRun(cfg.BaseURL, string(cfg.APIToken), cfg.OrganizationName, cfg.ProjectName, make(map[string]string), batchRunNumber)
	if err != nil {
		failf(err.Error())
	}
	return batchRun
}

func main() {

	// Parse configuration
	var cfg Config
	if err := stepconf.Parse(&cfg); err != nil {
		failf(err.Error())
	}
	var testSettings map[string]interface{}
	if cfg.TestSettings == "" {
		if cfg.TestSettingsNumber == 0 {
			failf("You have to specify either of Test settings number or Test settings")
		}
	} else {
		if errJSON := json.Unmarshal([]byte(cfg.TestSettings), &testSettings); errJSON != nil {
			failf(errJSON.Error())
		}
	}

	os.Setenv("MAGIC_POD_API_TOKEN", string(cfg.APIToken))
	os.Setenv("MAGIC_POD_ORGANIZATION", cfg.OrganizationName)
	os.Setenv("MAGIC_POD_PROJECT", cfg.ProjectName)

	stepconf.Print(cfg)
	fmt.Println()

	if err := os.Unsetenv("magic_pod_api_token"); err != nil {
		failf("Failed to remove API key data from envs, error: %s", err)
	}

	settingsStr := cfg.TestSettings
	// Upload app file if necessary and merge it into settings
	appFileNumber := -1
	if cfg.AppPath != "" {
		appFileNumber = uploadAppFile(cfg)
		if settingsStr == "" {
			settingsStr = fmt.Sprintf("{\"app_file_number\": %d}", appFileNumber)
		} else {
			// cfg.TestSettings is already parsed into testSettings
			individualSettings, exists := testSettings["test_settings"]
			if exists {
				settingArray, ok := individualSettings.([]interface{})
				if !ok {
					failf("Test settings are invalid. Value of \"test_settings\" should be an array")
				}
				for _, setting := range settingArray {
					settingMap := setting.(map[string]interface{})
					settingMap["app_file_number"] = appFileNumber
				}
			} else {
				testSettings["app_file_number"] = appFileNumber
			}
			settingsBytes, _ := json.Marshal(testSettings)
			settingsStr = string(settingsBytes)
		}
	}

	log.Infof(settingsStr)

	batchRuns, existsErr, existsUnresolved, cliErr := common.ExecuteBatchRun(cfg.BaseURL, string(cfg.APIToken),
		cfg.OrganizationName, cfg.ProjectName, make(map[string]string), cfg.TestSettingsNumber,
		settingsStr, cfg.WaitForResult, 0, true)
	if cliErr != nil {
		failf(cliErr.Error())
	}
	succeeded := !existsErr && !existsUnresolved

	if !cfg.WaitForResult {
		log.Infof("Exit this step because 'Wait for result' is set to false")
		os.Exit(0)
	}
	if appFileNumber != -1 {
		switch cfg.DeleteAppAfterTest {
		case "Always delete":
			deleteAppFile(cfg, appFileNumber)
			break
		case "Delete only when tests succeeded":
			if succeeded {
				deleteAppFile(cfg, appFileNumber)
			}
			break
		}
	}

	resultBytes, err := json.Marshal(batchRuns)
	result := string(resultBytes)
	if err != nil {
		failf(err.Error())
	}
	tools.ExportEnvironmentWithEnvman("MAGIC_POD_TEST_SUCCEEDED", strconv.FormatBool(succeeded))
	tools.ExportEnvironmentWithEnvman("MAGIC_POD_TEST_RESULT", result)
	if succeeded {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
