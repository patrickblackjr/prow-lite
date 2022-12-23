package test_data

import "github.com/patrickblackjr/prow-lite/cmd/app/config"

func init() {
	err := config.LoadConfig("/config")
	if err != nil {
		panic(err)
	}
}

func GetTestCaseFolder() string {
	return "/test_data/test_case_data"
}
