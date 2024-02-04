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
	"allure-server/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "log"
	"net/http"
	_ "os"
	"path/filepath"
)

type App struct {
	Config config2.AppConfig
}

func NewApp(config config2.AppConfig) *App {
	return &App{Config: config}
}

func afterRequestFunc(response http.ResponseWriter, request *http.Request) {
	// CORS middleware configuration

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
//
//	func loginEndpoint(c *gin.Context, timeout time.Duration) {
//		if config.ENABLE_SECURITY_LOGIN == false {
//			body := gin.H{
//				"meta_data": gin.H{
//					"message": "SECURITY is not enabled",
//				},
//			}
//			c.JSON(http.StatusNotFound, body)
//			return
//		}
//
//		contentType := c.GetHeader("Content-Type")
//		if contentType == "" || !strings.HasPrefix(contentType, "application/json") {
//			panic(errors.New("header 'Content-Type' must be 'application/json'"))
//		}
//
//		var jsonBody map[string]interface{}
//		err := c.ShouldBindJSON(&jsonBody)
//		if err != nil {
//			panic(errors.New("missing JSON in body request"))
//		}
//
//		username, ok := jsonBody["username"].(string)
//		if !ok || username == "" {
//			panic(errors.New("missing 'username' attribute"))
//		}
//		username = strings.ToLower(username)
//
//		if _, ok := token.USERS_INFO[username]; !ok {
//			body := gin.H{
//				"meta_data": gin.H{
//					"message": "Invalid username/password",
//				},
//			}
//			c.JSON(http.StatusUnauthorized, body)
//			return
//		}
//
//		password, ok := jsonBody["password"].(string)
//		if !ok || password == "" {
//			raiseError(c, errors.New("missing 'password' attribute"))
//			//return
//			//panic(errors.New("missing 'password' attribute"))
//		}
//
//		if token.USERS_INFO[username].Pass != password {
//			body := gin.H{
//				"meta_data": gin.H{
//					"message": "Invalid username/password",
//				},
//			}
//			c.JSON(http.StatusUnauthorized, body)
//			return
//		}
//
//		accessClaims := jwt.MapClaims{
//			"identity": username,
//			"exp":      time.Now().Add(timeout).Unix(),
//		}
//		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
//		accessTokenString, err := accessToken.SignedString([]byte("your_secret_key"))
//		if err != nil {
//			raiseError(c, err)
//			return
//		}
//
//		refreshClaims := jwt.MapClaims{
//			"identity": username,
//		}
//		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
//		refreshTokenString, err := refreshToken.SignedString([]byte("your_secret_key"))
//		if err != nil {
//			raiseError(c, err)
//			return
//			//panic(err)
//		}
//
//		expiresIn := timeout.Seconds()
//		jsonBody = gin.H{
//			"data": gin.H{
//				"access_token":  accessTokenString,
//				"refresh_token": refreshTokenString,
//				"expires_in":    expiresIn,
//				"roles":         token.USERS_INFO[username].Roles,
//			},
//			"meta_data": gin.H{
//				"message": "Successfully logged",
//			},
//		}
//		c.JSON(http.StatusOK, jsonBody)
//		return
//		//} catch ex {
//		//	body := gin.H{
//		//	"meta_data": gin.H{
//		//	"message": ex.Error(),
//		//},
//		//}
//		//	c.JSON(http.StatusBadRequest, body)
//		//	return
//		//}
//	}
func (appConfig *App) SwaggerJSONEndpoint(c *gin.Context) {
	specificationFile := "swagger.json"

	if appConfig.Config.JWTConfig.EnableSecurityLogin {
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
