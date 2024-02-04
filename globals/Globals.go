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

package globals

import (
	"allure-server/utils"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	DEV_MODE                     = 0
	HOST                         = "0.0.0.0"
	PORT                         = os.Getenv("PORT")
	THREADS                      = 7
	URL_SCHEME                   = "http"
	ENABLE_SECURITY_LOGIN        = false
	MAKE_VIEWER_ENDPOINTS_PUBLIC = false
	SECURITY_USER                = ""
	SECURITY_PASS                = ""
	SECURITY_VIEWER_USER         = ""
	SECURITY_VIEWER_PASS         = ""

	ADMIN_ROLE_NAME             = "admin"
	VIEWER_ROLE_NAME            = "viewer"
	SECURITY_SPECS_PATH         = "swagger/security_specs"
	ORIGIN                      = "api"
	URL_PREFIX                  = ""
	CHECK_RESULTS_EVERY_SECONDS = "1"
	KEEP_HISTORY                = "0"
	KEEP_HISTORY_LATEST         = "20"
	JWT_COOKIE_SECURE           = "true"
	OPTIMIZE_STORAGE            = 0
	LANGUAGES                   = []string{"en", "ru", "zh", "de", "nl", "he", "br", "pl", "ja", "es", "kr", "fr"}
	LANGUAGE_TEMPLATE           = "language.html"
	GLOBAL_CSS                  = "https://stackpath.bootstrapcdn.com/bootswatch/4.3.1/cosmo/bootstrap.css"
	REPORT_INDEX_FILE           = "index.html"
	EMAILABLE_REPORT_TITLE      = "Emailable Report"
	API_RESPONSE_LESS_VERBOSE   = 0
	ROOT                        = os.Getenv("ROOT")
	GENERATE_REPORT_PROCESS     = fmt.Sprintf("%s/generateAllureReport.sh", ROOT)
	KEEP_HISTORY_PROCESS        = fmt.Sprintf("%s/keepAllureHistory.sh", ROOT)
	CLEAN_HISTORY_PROCESS       = fmt.Sprintf("%s/cleanAllureHistory.sh", ROOT)
	CLEAN_RESULTS_PROCESS       = fmt.Sprintf("%s/cleanAllureResults.sh", ROOT)
	RENDER_EMAIL_REPORT_PROCESS = fmt.Sprintf("%s/renderEmailableReport.sh", ROOT)
	ALLURE_VERSION              = os.Getenv("ALLURE_VERSION")
	STATIC_CONTENT              = os.Getenv("STATIC_CONTENT")
	PROJECTS_DIRECTORY          = os.Getenv("STATIC_CONTENT_PROJECTS")
	EMAILABLE_REPORT_FILE_NAME  = os.Getenv("EMAILABLE_REPORT_FILE_NAME")
	EMAILABLE_REPORT_CSS        = GLOBAL_CSS
)

func ReadConfig() {

	// Configure JWT settings
	// Initialize JWT middleware
	utils.UpdateKey("EMAILABLE_REPORT_CSS_CDN", &EMAILABLE_REPORT_CSS)
	utils.UpdateKey("EMAILABLE_REPORT_TITLE", &EMAILABLE_REPORT_TITLE)
	utils.UpdateOrDefault0("API_RESPONSE_LESS_VERBOSE", &API_RESPONSE_LESS_VERBOSE, []int{0, 1})
	utils.UpdateOrDefault0("DEV_MODE", &DEV_MODE, []int{0, 1})

	updateUrlPrefix()
	utils.UpdateOrDefault0("OPTIMIZE_STORAGE", &OPTIMIZE_STORAGE, []int{0, 1})
	updateMakeViewerEndpointsPublic()

	getSecurityUser()

	getSecurityPass()

	getSecurityViewer()

}
func updateUrlPrefix() {
	if prefix, exists := os.LookupEnv("URL_PREFIX"); exists {
		if DEV_MODE == 1 {
			log.Print("URL_PREFIX is not supported when DEV_MODE is enabled")
		} else {
			if prefix = strings.TrimSpace(prefix); prefix != "" {
				if !strings.HasPrefix(prefix, "/") {
					log.Print("Adding slash at the beginning of URL_PREFIX")
					URL_PREFIX = "/" + prefix
				}
				log.Printf("Setting URL_PREFIX=%s", URL_PREFIX)
			} else {
				log.Print("URL_PREFIX is empty. It won't be applied")
			}
		}
	}

}
func updateMakeViewerEndpointsPublic() {
	if makeViewerEndpointsPublic, exists := os.LookupEnv("MAKE_VIEWER_ENDPOINTS_PUBLIC"); exists {
		if viewerEndpointsPublicTmp, err := strconv.Atoi(makeViewerEndpointsPublic); err == nil {
			if viewerEndpointsPublicTmp == 1 {
				MAKE_VIEWER_ENDPOINTS_PUBLIC = true
				log.Printf("Overriding MAKE_VIEWER_ENDPOINTS_PUBLIC=%d\n", viewerEndpointsPublicTmp)
			}
		} else {
			log.Printf("Wrong env var value. Setting VIEWER_ENDPOINTS_PUBLIC=0 by default. Error: %v\n", err)
		}
	}
}
func getSecurityViewer() {
	if !MAKE_VIEWER_ENDPOINTS_PUBLIC {
		if securityViewerUserTmp, exists := os.LookupEnv("SECURITY_VIEWER_USER"); exists {
			if trimmedViewerUser := strings.TrimSpace(securityViewerUserTmp); trimmedViewerUser != "" {
				SECURITY_VIEWER_USER = strings.ToLower(trimmedViewerUser)
				log.Println("Setting SECURITY_VIEWER_USER")
			}
		}

		if securityViewerPassTmp, exists := os.LookupEnv("SECURITY_VIEWER_PASS"); exists {
			if trimmedViewerPass := strings.TrimSpace(securityViewerPassTmp); trimmedViewerPass != "" {
				SECURITY_VIEWER_PASS = trimmedViewerPass
				log.Println("Setting SECURITY_VIEWER_PASS")
			}
		}
	}
}
func getSecurityPass() {
	if securityPassTmp, exists := os.LookupEnv("SECURITY_PASS"); exists {
		if trimmedPass := strings.TrimSpace(securityPassTmp); trimmedPass != "" {
			SECURITY_PASS = trimmedPass
			log.Println("Setting SECURITY_PASS")
		}
	}
}
func getSecurityUser() {
	if securityUserTmp, exists := os.LookupEnv("SECURITY_USER"); exists {
		if trimmedUser := strings.TrimSpace(securityUserTmp); trimmedUser != "" {
			SECURITY_USER = strings.ToLower(trimmedUser)
			log.Println("Setting SECURITY_USER")
		}
	}
}
