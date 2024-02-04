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

package app

import (
	config2 "allure-server/config"
	config "allure-server/globals"
	"allure-server/token"
	"allure-server/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	_ "log"
	"net/http"
	_ "os"
	"path/filepath"
	"strings"
	"time"
)

type App struct {
	Config config2.AppConfig
}

func NewApp(config config2.AppConfig) *App {
	return &App{Config: config}
}

//func resolveProject(projectID string) string {
//	// implementation not provided
//	return projectID
//}
//
//func isExistentProject(projectID string) bool {
//	// implementation not provided
//	return true
//}
//
//func swaggerJSONEndpoint(w http.ResponseWriter, r *http.Request) {
//	specificationFile := "swagger.json"
//	//if ENABLE_SECURITY_LOGIN {
//	//	specificationFile = "swagger_security.json"
//	//}
//	if URL_PREFIX != "" {
//		spec, err := getFileAsString(fmt.Sprintf("%s/swagger/%s", STATIC_CONTENT, specificationFile))
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusBadRequest)
//			return
//		}
//		var specJSON map[string]interface{}
//		err = json.Unmarshal([]byte(spec), &specJSON)
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusBadRequest)
//			return
//		}
//		serverURL := specJSON["servers"].([]interface{})[0].(map[string]interface{})["url"].(string)
//		specJSON["servers"].([]interface{})[0].(map[string]interface{})["url"] = fmt.Sprintf("%s%s", URL_PREFIX, serverURL)
//		json.NewEncoder(w).Encode(specJSON)
//	} else {
//		http.ServeFile(w, r, fmt.Sprintf("%s/swagger/%s", STATIC_CONTENT, specificationFile))
//	}
//}
//
//func versionEndpoint(w http.ResponseWriter, r *http.Request) {
//	version, err := getFileAsString(ALLURE_VERSION)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	json.NewEncoder(w).Encode(map[string]interface{}{
//		"data": map[string]interface{}{
//			"version": strings.TrimSpace(version),
//		},
//		"meta_data": map[string]interface{}{
//			"message": "Version successfully obtained",
//		},
//	})
//}
//
//func configEndpoint(w http.ResponseWriter, r *http.Request) {
//	version, err := getFileAsString(ALLURE_VERSION)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	json.NewEncoder(w).Encode(map[string]interface{}{
//		"data": map[string]interface{}{
//			"version":                      strings.TrimSpace(version),
//			"dev_mode":                     DEV_MODE,
//			"check_results_every_seconds":  CHECK_RESULTS_EVERY_SECONDS,
//			"keep_history":                 KEEP_HISTORY,
//			"keep_history_latest":          KEEP_HISTORY_LATEST,
//			"tls":                          JWT_COOKIE_SECURE,
//			"security_enabled":             ENABLE_SECURITY_LOGIN,
//			"url_prefix":                   URL_PREFIX,
//			"api_response_less_verbose":    API_RESPONSE_LESS_VERBOSE,
//			"optimize_storage":             OPTIMIZE_STORAGE,
//			"make_viewer_endpoints_public": MAKE_VIEWER_ENDPOINTS_PUBLIC,
//		},
//		"meta_data": map[string]interface{}{
//			"message": "Config successfully obtained",
//		},
//	})
//}
//
//func selectLanguageEndpoint(w http.ResponseWriter, r *http.Request) {
//	code := r.URL.Query().Get("code")
//	if code == "" {
//		http.Error(w, "'code' query parameter is required", http.StatusBadRequest)
//		return
//	}
//	code = strings.ToLower(code)
//	if !contains(LANGUAGES, code) {
//		http.Error(w, fmt.Sprintf("'code' not supported. Use values: %v", LANGUAGES), http.StatusBadRequest)
//		return
//	}
//	http.ServeFile(w, r, fmt.Sprintf("%s/%s", STATIC_CONTENT, LANGUAGE_TEMPLATE))
//}
//
//func latestReportEndpoint(w http.ResponseWriter, r *http.Request) {
//	projectID := resolveProject(r.URL.Query().Get("project_id"))
//	if !isExistentProject(projectID) {
//		http.Error(w, fmt.Sprintf("project_id '%s' not found", projectID), http.StatusNotFound)
//		return
//	}
//	projectReportLatestPath := fmt.Sprintf("/latest/%s", REPORT_INDEX_FILE)
//	url := fmt.Sprintf("%s/%s/%s", STATIC_CONTENT, projectID, projectReportLatestPath)
//	http.Redirect(w, r, url, http.StatusFound)
//}

func afterRequestFunc(response http.ResponseWriter, request *http.Request) {
	// CORS middleware configuration
	//origin := request.Header.Get("Origin")
	//config := cors.DefaultConfig()
	//config.AllowCredentials = true
	//config.AllowAllOrigins = true // Adjust this based on your security requirements
	//if request.Method == "OPTIONS" {
	//	config.AddAllowHeaders("Content-Type", "x-csrf-token")
	//	config.AddAllowMethods("GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE")
	//	//if origin != "" {
	//	//	config.AllowOriginFunc = func(origin string) bool {
	//	//		return true
	//	//	}
	//	//}
	//}
	//response.
	//	cors.New(config)

	origin := request.Header.Get("Origin")
	if request.Method == "OPTIONS" {
		//response = makeResponse()
		response.Header().Add("Access-Control-Allow-Credentials", "true")
		response.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		response.Header().Add("Access-Control-Allow-Headers", "x-csrf-token")
		response.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")
		if origin != "" {
			response.Header().Add("Access-Control-Allow-Origin", origin)
		}
	} else {
		response.Header().Add("Access-Control-Allow-Credentials", "true")
		if origin != "" {
			response.Header().Add("Access-Control-Allow-Origin", origin)
		}
	}
}

// Security Endpoints Section
func loginEndpoint(c *gin.Context, timeout time.Duration) {
	if config.ENABLE_SECURITY_LOGIN == false {
		body := gin.H{
			"meta_data": gin.H{
				"message": "SECURITY is not enabled",
			},
		}
		c.JSON(http.StatusNotFound, body)
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "application/json") {
		panic(errors.New("header 'Content-Type' must be 'application/json'"))
	}

	var jsonBody map[string]interface{}
	err := c.ShouldBindJSON(&jsonBody)
	if err != nil {
		panic(errors.New("missing JSON in body request"))
	}

	username, ok := jsonBody["username"].(string)
	if !ok || username == "" {
		panic(errors.New("missing 'username' attribute"))
	}
	username = strings.ToLower(username)

	if _, ok := token.USERS_INFO[username]; !ok {
		body := gin.H{
			"meta_data": gin.H{
				"message": "Invalid username/password",
			},
		}
		c.JSON(http.StatusUnauthorized, body)
		return
	}

	password, ok := jsonBody["password"].(string)
	if !ok || password == "" {
		raiseError(c, errors.New("missing 'password' attribute"))
		//return
		//panic(errors.New("missing 'password' attribute"))
	}

	if token.USERS_INFO[username].Pass != password {
		body := gin.H{
			"meta_data": gin.H{
				"message": "Invalid username/password",
			},
		}
		c.JSON(http.StatusUnauthorized, body)
		return
	}

	accessClaims := jwt.MapClaims{
		"identity": username,
		"exp":      time.Now().Add(timeout).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte("your_secret_key"))
	if err != nil {
		raiseError(c, err)
		return
	}

	refreshClaims := jwt.MapClaims{
		"identity": username,
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte("your_secret_key"))
	if err != nil {
		raiseError(c, err)
		return
		//panic(err)
	}

	expiresIn := timeout.Seconds()
	jsonBody = gin.H{
		"data": gin.H{
			"access_token":  accessTokenString,
			"refresh_token": refreshTokenString,
			"expires_in":    expiresIn,
			"roles":         token.USERS_INFO[username].Roles,
		},
		"meta_data": gin.H{
			"message": "Successfully logged",
		},
	}
	c.JSON(http.StatusOK, jsonBody)
	return
	//} catch ex {
	//	body := gin.H{
	//	"meta_data": gin.H{
	//	"message": ex.Error(),
	//},
	//}
	//	c.JSON(http.StatusBadRequest, body)
	//	return
	//}
}
func (appConfig *App) SwaggerJSONEndpoint(c *gin.Context) {
	specificationFile := "swagger.json"

	if appConfig.Config.EnableSecurityLogin {
		specificationFile = "swagger_security.json"
	}

	if appConfig.Config.UrlPrefix != "" {
		// Replace with the path to your static content directory
		staticContent := "/static"
		spec, err := utils.GetFileAsString(filepath.Join(staticContent, "swagger", specificationFile))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"meta_data": gin.H{"message": err.Error()}})
			return
		}

		var specJSON map[string]interface{}
		if err := json.Unmarshal([]byte(spec), &specJSON); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"meta_data": gin.H{"message": err.Error()}})
			return
		}

		serverURL, ok := specJSON["servers"].([]interface{})[0].(map[string]interface{})["url"].(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"meta_data": gin.H{"message": "Invalid server URL"}})
			return
		}

		specJSON["servers"].([]interface{})[0].(map[string]interface{})["url"] = fmt.Sprintf("%s%s", appConfig.Config.UrlPrefix, serverURL)

		c.JSON(http.StatusOK, specJSON)
		return
	}

	// Replace with the path to your static content directory
	filePath := filepath.Join(appConfig.Config.StaticContent, "swagger", specificationFile)
	c.File(filePath)
}
func raiseError(c *gin.Context, err error) {
	body := gin.H{
		"meta_data": gin.H{
			"message": err,
		}}
	c.JSON(http.StatusBadRequest, body)
}
