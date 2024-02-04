// Copyright 2024 Maxim Tverdohleb <tverdohleb.maxim@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"allure-server/globals"
	"allure-server/token"
	"allure-server/utils"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type AppConfig struct {
	DevMode                  bool   //enables features for developing
	Host                     string //host of web app
	Port                     int    //port of web app
	Threads                  byte   //number of threads to work with
	UrlPrefix                string //adds prefix to app url: {UrlScheme}://{host}:{port}/{UrlPrefix}
	Origin                   string //defines origin of requests
	CheckResultsEverySeconds bool   // checks results every second to generate report
	KeepHistory              bool   //Allows to keep history of runs, otherwise new report will override current result
	KeepHistoryLatest        int8   // how many history id to keep. App will show only last {KeepHistoryLatest}
	OptimizeStorage          bool

	Languages                []string
	ApiResponseLessVerbose   bool
	GenerateReportProcess    string
	KeepHistoryProcess       string
	CleanHistoryProcess      string
	CleanResultsProcess      string
	RenderEmailReportProcess string
	AllureVersion            string
	SecuritySpecsPath        string
	StaticContent            string
	ProjectsDirectory        string

	EmailableReportTitle    string
	EmailableReportFileName string
	EmailableReportCss      string
	JWTConfig               token.JWTConfig
}

func DefaultConfig() AppConfig {
	return AppConfig{
		DevMode:                  false,
		Host:                     "0.0.0.0",
		Port:                     8080,
		Threads:                  7,
		UrlPrefix:                "",
		Origin:                   "api",
		CheckResultsEverySeconds: false,
		KeepHistory:              false,
		KeepHistoryLatest:        30,
		OptimizeStorage:          false,
		ApiResponseLessVerbose:   false,
		EmailableReportCss:       globals.GlobalCss,
		StaticContent:            "static",
		SecuritySpecsPath:        "swagger/security_specs",
		JWTConfig:                token.DefaultConfig(),
	}
}
func (appCfg *AppConfig) GetAddress() string {
	return fmt.Sprintf("%s://%s:%d%s", appCfg.JWTConfig.UrlScheme, appCfg.Host, appCfg.Port, appCfg.UrlPrefix)
}

func DefaultConfigEnv() (AppConfig, error) {
	root := os.Getenv(globals.Root)
	appConfig := DefaultConfig()
	appConfig.GenerateReportProcess = fmt.Sprintf("%s/generateAllureReport.sh", root)
	appConfig.KeepHistoryProcess = fmt.Sprintf("%s/keepAllureHistory.sh", root)
	appConfig.CleanHistoryProcess = fmt.Sprintf("%s/cleanAllureHistory.sh", root)
	appConfig.CleanResultsProcess = fmt.Sprintf("%s/cleanAllureResults.sh", root)
	appConfig.RenderEmailReportProcess = fmt.Sprintf("%s/renderEmailableReport.sh", root)
	//appConfig := AppConfig{
	//	DevMode:                   false,
	//	Host:                      "0.0.0.0",
	//	Threads:                   7,
	//	UrlScheme:                 "http",
	//	EnableSecurityLogin:       false,
	//	MakeViewerEndpointsPublic: false,
	//	Origin:                    "api",
	//	CheckResultsEverySeconds:  false,
	//	KeepHistory:               false,
	//	KeepHistoryLatest:         30,
	//	OptimizeStorage:           false,
	//	ApiResponseLessVerbose:    false,
	//
	//	GenerateReportProcess:    fmt.Sprintf("%s/generateAllureReport.sh", root),
	//	KeepHistoryProcess:       fmt.Sprintf("%s/keepAllureHistory.sh", root),
	//	CleanHistoryProcess:      fmt.Sprintf("%s/cleanAllureHistory.sh", root),
	//	CleanResultsProcess:      fmt.Sprintf("%s/cleanAllureResults.sh", root),
	//	RenderEmailReportProcess: fmt.Sprintf("%s/renderEmailableReport.sh", root),
	//
	//	AllureVersion:           os.Getenv("ALLURE_VERSION"),
	//	StaticContent:           os.Getenv("STATIC_CONTENT"),
	//	ProjectsDirectory:       os.Getenv("STATIC_CONTENT_PROJECTS"),
	//	EmailableReportFileName: os.Getenv("EMAILABLE_REPORT_FILE_NAME"),
	//	EmailableReportCss:      GlobalCss,
	//}

	port, err := strconv.Atoi(os.Getenv(globals.Port))
	if err != nil {
		return appConfig, err
	}
	appConfig.Port = port

	utils.UpdateKey(globals.AllureVersion, &appConfig.AllureVersion)
	utils.UpdateKey(globals.StaticContent, &appConfig.StaticContent)
	utils.UpdateKey(globals.StaticContentProjects, &appConfig.ProjectsDirectory)
	utils.UpdateKey(globals.EmailableReportFileName, &appConfig.EmailableReportFileName)

	utils.UpdateKey(globals.EmailableReportCssCdn, &appConfig.EmailableReportCss)
	utils.UpdateKey(globals.EmailableReportTitle, &appConfig.EmailableReportTitle)

	utils.UpdateOrDefaultFalse(globals.DevMode, &appConfig.DevMode)

	utils.UpdateOrDefaultFalse(globals.ApiResponseLessVerbose, &appConfig.ApiResponseLessVerbose)
	utils.UpdateOrDefaultFalse(globals.OptimizeStorage, &appConfig.OptimizeStorage)

	updateUrlPrefix(appConfig)
	appConfig.JWTConfig.UpdateFromEnv()
	return appConfig, nil
	//utils.UpdateKey("EMAILABLE_REPORT_CSS_CDN", &EmailableReportCss)
}
func updateUrlPrefix(config AppConfig) {
	if prefix, exists := os.LookupEnv(globals.UrlPrefix); exists {
		if config.DevMode {
			log.Printf("%s is not supported when %s is enabled", globals.UrlPrefix, globals.DevMode)
		} else {
			if prefix = strings.TrimSpace(prefix); prefix != "" {
				if !strings.HasPrefix(prefix, "/") {
					log.Printf("Adding slash at the beginning of %s", globals.UrlPrefix)
					config.UrlPrefix = "/" + prefix
				}
				log.Printf("Setting %s=%s", globals.UrlPrefix, config.UrlPrefix)
			} else {
				log.Printf("%s is empty. It won't be applied", globals.UrlPrefix)
			}
		}
	}

}
