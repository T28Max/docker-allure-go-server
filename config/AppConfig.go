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
	"allure-server/utils"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	LanguageTemplate = "language.html"
	GlobalCss        = "https://stackpath.bootstrapcdn.com/bootswatch/4.3.1/cosmo/bootstrap.css"
	ReportIndexFile  = "index.html"
)

type AppConfig struct {
	DevMode                   bool   //enables features for developing
	Host                      string //host of web app
	Port                      int    //port of web app
	Threads                   byte   //number of threads to work with
	UrlScheme                 string // used schema, HTTP/HTTPS
	UrlPrefix                 string //adds prefix to app url: {UrlScheme}://{host}:{port}/{UrlPrefix}
	EnableSecurityLogin       bool   //enables login
	MakeViewerEndpointsPublic bool
	Origin                    string //defines origin of requests
	CheckResultsEverySeconds  bool   // checks results every second to generate report
	KeepHistory               bool   //Allows to keep history of runs, otherwise new report will override current result
	KeepHistoryLatest         int8   // how many history id to keep. App will show only last {KeepHistoryLatest}
	OptimizeStorage           bool

	Languages                []string
	ApiResponseLessVerbose   bool
	GenerateReportProcess    string
	KeepHistoryProcess       string
	CleanHistoryProcess      string
	CleanResultsProcess      string
	RenderEmailReportProcess string
	AllureVersion            string
	StaticContent            string
	ProjectsDirectory        string

	EmailableReportTitle    string
	EmailableReportFileName string
	EmailableReportCss      string
}

func (appCfg *AppConfig) GetAddress() string {
	return fmt.Sprintf("%s://%s:%d%s", appCfg.UrlScheme, appCfg.Host, appCfg.Port, appCfg.UrlPrefix)
}
func DefaultConfig() AppConfig {
	return AppConfig{
		DevMode:                   false,
		Host:                      "0.0.0.0",
		Port:                      8080,
		Threads:                   7,
		UrlScheme:                 "http",
		UrlPrefix:                 "",
		EnableSecurityLogin:       false,
		MakeViewerEndpointsPublic: false,
		Origin:                    "api",
		CheckResultsEverySeconds:  false,
		KeepHistory:               false,
		KeepHistoryLatest:         30,
		OptimizeStorage:           false,
		ApiResponseLessVerbose:    false,
		EmailableReportCss:        GlobalCss,
		StaticContent:             "static",
	}
}
func DefaultConfigEnv() (AppConfig, error) {
	root := os.Getenv("ROOT")
	appConfig := AppConfig{
		DevMode:                   false,
		Host:                      "0.0.0.0",
		Threads:                   7,
		UrlScheme:                 "http",
		EnableSecurityLogin:       false,
		MakeViewerEndpointsPublic: false,
		Origin:                    "api",
		CheckResultsEverySeconds:  false,
		KeepHistory:               false,
		KeepHistoryLatest:         30,
		OptimizeStorage:           false,
		ApiResponseLessVerbose:    false,

		GenerateReportProcess:    fmt.Sprintf("%s/generateAllureReport.sh", root),
		KeepHistoryProcess:       fmt.Sprintf("%s/keepAllureHistory.sh", root),
		CleanHistoryProcess:      fmt.Sprintf("%s/cleanAllureHistory.sh", root),
		CleanResultsProcess:      fmt.Sprintf("%s/cleanAllureResults.sh", root),
		RenderEmailReportProcess: fmt.Sprintf("%s/renderEmailableReport.sh", root),

		AllureVersion:           os.Getenv("ALLURE_VERSION"),
		StaticContent:           os.Getenv("STATIC_CONTENT"),
		ProjectsDirectory:       os.Getenv("STATIC_CONTENT_PROJECTS"),
		EmailableReportFileName: os.Getenv("EMAILABLE_REPORT_FILE_NAME"),
		EmailableReportCss:      GlobalCss,
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return appConfig, err
	}
	appConfig.Port = port

	utils.UpdateKey("EMAILABLE_REPORT_CSS_CDN", &appConfig.EmailableReportCss)
	utils.UpdateKey("EMAILABLE_REPORT_TITLE", &appConfig.EmailableReportTitle)

	utils.UpdateOrDefaultFalse("DEV_MODE", &appConfig.DevMode)
	utils.UpdateOrDefaultFalse("API_RESPONSE_LESS_VERBOSE", &appConfig.ApiResponseLessVerbose)
	utils.UpdateOrDefaultFalse("OPTIMIZE_STORAGE", &appConfig.OptimizeStorage)

	updateUrlPrefix(appConfig)
	return appConfig, nil
	//utils.UpdateKey("EMAILABLE_REPORT_CSS_CDN", &EmailableReportCss)
}
func updateUrlPrefix(config AppConfig) {
	if prefix, exists := os.LookupEnv("URL_PREFIX"); exists {
		if config.DevMode {
			log.Print("URL_PREFIX is not supported when DEV_MODE is enabled")
		} else {
			if prefix = strings.TrimSpace(prefix); prefix != "" {
				if !strings.HasPrefix(prefix, "/") {
					log.Print("Adding slash at the beginning of URL_PREFIX")
					config.UrlPrefix = "/" + prefix
				}
				log.Printf("Setting URL_PREFIX=%s", config.UrlPrefix)
			} else {
				log.Print("URL_PREFIX is empty. It won't be applied")
			}
		}
	}

}
