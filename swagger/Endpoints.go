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

package swagger

import (
	"allure-server/config"
	"allure-server/utils"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// EndpointInfo represents information about a protected endpoint
type EndpointInfo struct {
	Method   string
	Path     string
	Endpoint string
}

var protectedEndpoints = []EndpointInfo{
	{"post", "/refresh", "refresh_endpoint"},
	{"delete", "/logout", "logout_endpoint"},
	{"delete", "/logout-refresh-token", "logout_refresh_token_endpoint"},
	{"post", "/send-results", "send_results_endpoint"},
	{"get", "/generate-report", "generate_report_endpoint"},
	{"get", "/clean-results", "clean_results_endpoint"},
	{"get", "/clean-history", "clean_history_endpoint"},
	{"post", "/projects", "create_project_endpoint"},
	{"delete", "/projects/{id}", "delete_project_endpoint"},
}

func IsEndpointProtected(endpoint string, appConfig config.AppConfig) bool {
	if appConfig.MakeViewerEndpointsPublic == false {
		return true
	}
	for _, info := range protectedEndpoints {
		if endpoint == info.Endpoint {
			return true
		}
	}
	return false
}
func isEndpointSwaggerProtected(method, path string, appConfig config.AppConfig) bool {
	if appConfig.MakeViewerEndpointsPublic == false {
		return true
	}
	for _, info := range protectedEndpoints {
		if info.Method == method && info.Path == path {
			return true
		}
	}
	return false
}

func getReportsEndpoint(w http.ResponseWriter, r *http.Request, appConfig config.AppConfig) {
	vars := mux.Vars(r)
	projectID := vars["project_id"]
	path := vars["path"]
	redirect := vars["redirect"]
	if redirect == "false" {
		http.ServeFile(w, r, fmt.Sprintf("%s/%s/%s", appConfig.StaticContent, projectID, path))
	} else {
		http.Redirect(w, r, fmt.Sprintf("%s/%s/%s", appConfig.StaticContent, projectID, path), http.StatusFound)
	}
}
func createProject(jsonBody map[string]interface{}, appConfig config.AppConfig) (string, error) {
	if _, ok := jsonBody["id"]; !ok {
		return "", errors.New("'id' is required in the body")
	}

	id, ok := jsonBody["id"].(string)
	if !ok {
		return "", errors.New("'id' should be a string")
	}

	if strings.TrimSpace(id) == "" {
		return "", errors.New("'id' should not be empty")
	}

	if len(id) > 100 {
		return "", errors.New("'id' should not contain more than 100 characters")
	}

	projectIDPattern := regexp.MustCompile("^[a-z\\d]([a-z\\d -]*[a-z\\d])?$")
	if !projectIDPattern.MatchString(id) {
		return "", errors.New("'id' should contain alphanumeric lowercase characters or hyphens. For example: 'my-project-id'")
	}

	if utils.IsExistentProject(id, appConfig.ProjectsDirectory) {
		return "", errors.New("project_id '" + id + "' is existent")
	}

	if id == "default" {
		return "", errors.New("the id 'default' is not allowed. Try with another project_id")
	}

	projectPath := utils.GetProjectPath(id, appConfig.ProjectsDirectory)
	latestReportProject := projectPath + "/reports/latest"
	resultsProject := projectPath + "/results"

	if _, err := os.Stat(latestReportProject); os.IsNotExist(err) {
		err := os.MkdirAll(latestReportProject, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(resultsProject); os.IsNotExist(err) {
		err := os.MkdirAll(resultsProject, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}
