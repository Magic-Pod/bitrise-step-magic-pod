package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/magic-pod/magic-pod-api-client/common"
	"github.com/mholt/archiver"
	"gopkg.in/resty.v1"
)

// Config : Configuration for this step
type Config struct {
	BaseURL             string          `env:"base_url,required"`
	APIToken            stepconf.Secret `env:"magic_pod_api_token,required"`
	OrganizationName    string          `env:"organization_name,required"`
	ProjectName         string          `env:"project_name,required"`
	AppPath             string          `env:"app_path"`
	TestConditionNumber int             `env:"test_condition_number`
	WaitForResult       bool            `env:"wait_for_result"`
	DeleteAppAfterTest  string          `env:"delete_app_after_test"`
}

// UploadFile : Response from upload-file API
type UploadFile struct {
	FileName string `json:"file_name"`
	FileNo   int    `json:"file_no"`
}

// TestCases : Part of response from batch-run API. It stands for number of test cases
type TestCases struct {
	Succeeded  int `json:"succeeded"`
	Failed     int `json:"failed"`
	Unresolved int `json:"unresolved"`
	Total      int `json:"total"`
}

// BatchRun : Response from batch-run API
type BatchRun struct {
	Organizationname string    `json:"organization_name"`
	ProjectName      string    `json:"project_name"`
	BatchRunNumber   int       `json:"batch_run_number"`
	Status           string    `json:"status"`
	TestCases        TestCases `json:"test_cases"`
	URL              string    `json:"url"`
}

// ErrorResponse : Response from APIs when they are not finished with status 200
type ErrorResponse struct {
	Detail   string                 `json:"detail"`
	ErrorMap map[string]interface{} `json:"."`
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func handleError(resp *resty.Response, err error) {
	if err != nil {
		failf(resp.Status())
	}
	if resp.StatusCode() != 200 {
		errorResp := resp.Error().(*ErrorResponse)
		if errorResp.Detail != "" {
			failf("%s: %s", resp.Status(), errorResp.Detail)
		} else {
			var result map[string][]string
			log.Errorf("%s:", resp.Status())
			if err := json.Unmarshal([]byte(resp.String()), &result); err != nil {
				// Unexpectedly returned HTML
				os.Exit(1)
			}
			for key, value := range result {
				log.Errorf("\t%s: %s", key, strings.Join(value, ","))
			}
			os.Exit(1)
		}
	}
}

func createBaseRequest(cfg Config) *resty.Request {
	return resty.
		SetHostURL(cfg.BaseURL).R().
		SetHeader("Authorization", "Token "+string(cfg.APIToken)).
		SetPathParams(map[string]string{
			"organization_name": cfg.OrganizationName,
			"project_name":      cfg.ProjectName,
		}).
		SetError(ErrorResponse{})
}

func zipAppDir(dirPath string) string {
	log.Infof("Zip app directory %s", dirPath)
	zipPath := dirPath + ".zip"
	if err := os.RemoveAll(zipPath); err != nil {
		failf(err.Error())
	}
	if err := archiver.Archive([]string{dirPath}, zipPath); err != nil {
		failf(err.Error())
	}
	fmt.Println()
	return zipPath
}

func uploadAppFile(cfg Config) int {
	appPath := cfg.AppPath
	log.Infof("Upload app file %s to Magic Pod cloud", appPath)
	fileNoBytes, err := exec.Command(
		"./magic-pod-api-client", "--url-base", cfg.BaseURL, "upload-app", "-a", cfg.AppPath,
	).Output()
	if err != nil {
		panic(err)
	}
	fileNoStr := string(fileNoBytes)
	fileNo, err := strconv.Atoi(strings.TrimRight(fileNoStr, "\n"))
	if err != nil {
		panic(err)
	}
	log.Donef("Done. File number = %d\n", fileNo)
	return fileNo
}

func deleteAppFile(cfg Config, appFileNumber int) {
	log.Infof("Delete app file on the server")
	err := exec.Command(
		"./magic-pod-api-client", "--url-base", cfg.BaseURL, "delete-app", "-a", strconv.Itoa(appFileNumber),
	).Run()
	if err != nil {
		panic(err)
	}
}

func getBatchRun(cfg Config, batchRunNumber int) *BatchRun {
	resp, err := createBaseRequest(cfg).
		SetPathParams(map[string]string{
			"batch_run_number": strconv.Itoa(batchRunNumber),
		}).
		SetResult(BatchRun{}).
		Get("/{organization_name}/{project_name}/batch-run/{batch_run_number}/")
	handleError(resp, err)
	return resp.Result().(*BatchRun)
}


func main() {

	// Parse configuration
	var cfg Config
	if err := stepconf.Parse(&cfg); err != nil {
		failf(err.Error())
	}
	os.Setenv("MAGIC_POD_API_TOKEN", string(cfg.APIToken))
	os.Setenv("MAGIC_POD_ORGANIZATION", cfg.OrganizationName)
	os.Setenv("MAGIC_POD_PROJECT", cfg.ProjectName)

	stepconf.Print(cfg)
	fmt.Println()

	if err := os.Unsetenv("magic_pod_api_token"); err != nil {
		failf("Failed to remove API key data from envs, error: %s", err)
	}

	common.DeleteApp(cfg.BaseURL, string(cfg.APIToken), cfg.OrganizationName, cfg.ProjectName, {}, 12)

	// Upload app file if necessary
	appFileNumber := -1
	if cfg.AppPath != "" {
		appFileNumber = uploadAppFile(cfg)
	}
	fmt.Printf("appFileNumber = %d\n", appFileNumber)

	// // Post request to start batch run
	// batchRun := startBatchRun(cfg, appFileNumber)
	// tools.ExportEnvironmentWithEnvman("MAGIC_POD_TEST_URL", batchRun.URL)

	if !cfg.WaitForResult {
		log.Successf("Exit this step because 'Wait for result' is set to false")
		os.Exit(0)
	}
	// TODO other option
	if appFileNumber != -1 {
		switch cfg.DeleteAppAfterTest {
		case "Always delete":
			deleteAppFile(cfg, appFileNumber)
			break
		}
	}

	// // Show result
	// testCases := batchRun.TestCases
	// message := fmt.Sprintf("\nMagic Pod test %s: \n"+
	// 	"\tSucceeded : %d\n"+
	// 	"\tFailed : %d\n"+
	// 	"\tUnresolved : %d\n"+
	// 	"\tTotal : %d\n"+
	// 	"Please see %s for detail",
	// 	batchRun.Status, testCases.Succeeded, testCases.Failed, testCases.Unresolved, testCases.Total, batchRun.URL)
	// tools.ExportEnvironmentWithEnvman("MAGIC_POD_TEST_STATUS", batchRun.Status)
	// tools.ExportEnvironmentWithEnvman("MAGIC_POD_TEST_SUCCEEDED_COUNT", strconv.Itoa(testCases.Succeeded))
	// tools.ExportEnvironmentWithEnvman("MAGIC_POD_TEST_FAILED_COUNT", strconv.Itoa(testCases.Failed))
	// tools.ExportEnvironmentWithEnvman("MAGIC_POD_TEST_UNRESOLVED_COUNT", strconv.Itoa(testCases.Unresolved))
	// tools.ExportEnvironmentWithEnvman("MAGIC_POD_TEST_TOTAL_COUNT", strconv.Itoa(testCases.Total))
	// switch batchRun.Status {
	// case "succeeded":
	// 	log.Successf(message)
	// default:
	// 	failf(message)
	// }
	failf("not yet completed")

	os.Exit(0)
}
