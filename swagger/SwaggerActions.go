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
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
	"os"
	"path/filepath"
)

type appSwagger struct {
	Config config.AppConfig
}

func (s *appSwagger) ReadDoc() string {
	specificationFile := "swagger.json"
	if s.Config.JWTConfig.EnableSecurityLogin {
		specificationFile = "swagger_security.json"
	}
	// Replace with the path to your static content directory
	staticContent := "static"
	spec, err := utils.GetFileAsString(filepath.Join(staticContent, "swagger", specificationFile))
	if err != nil {
		panic(err)
	}
	if s.Config.UrlPrefix != "" {
		var specJSON map[string]interface{}
		if err := json.Unmarshal([]byte(spec), &specJSON); err != nil {
			panic(err)
		}

		serverURL, ok := specJSON["servers"].([]interface{})[0].(map[string]interface{})["url"].(string)
		if !ok {
			panic("Invalid server URL")
		}

		specJSON["servers"].([]interface{})[0].(map[string]interface{})["url"] = fmt.Sprintf("%s%s", s.Config.UrlPrefix, serverURL)

		b, err := json.Marshal(specJSON)
		if err != nil {
			panic(err)
		}
		return string(b)
	}
	return spec

}

//func init() {
//	doc := &appSwagger{}
//	swag.Register("swagger", doc)
//}

type Config struct {
	NativePrefix        string
	SwaggerEndpoint     string
	SwaggerSpecFile     string
	SwaggerEndpointPath string
	SwaggerSpec         string
}

const (
	NativePrefix = "/allure-docker-service"
	Endpoint     = "/swagger"
	SpecFile     = "/swagger.json"
	//SWAGGER_ENDPOINT_PATH = fmt.Sprintf("%s%s", NATIVE_PREFIX, SWAGGER_ENDPOINT)
	//SWAGGER_SPEC          = fmt.Sprintf("%s%s", NATIVE_PREFIX, SWAGGER_SPEC_FILE)
)

func DefaultConfig() Config {
	return Config{
		NativePrefix:        NativePrefix,
		SwaggerEndpoint:     Endpoint,
		SwaggerSpecFile:     SpecFile,
		SwaggerEndpointPath: fmt.Sprintf("%s%s", NativePrefix, Endpoint),
		SwaggerSpec:         fmt.Sprintf("%s%s", NativePrefix, SpecFile),
	}

}
func (swagger Config) Update(appConfig config.AppConfig) (*gin.Engine, error) {
	err := error(nil)
	if appConfig.UrlPrefix != "" {
		swagger.SwaggerEndpointPath = fmt.Sprintf("%s%s%s", appConfig.UrlPrefix, swagger.NativePrefix, swagger.SwaggerEndpoint)
		swagger.SwaggerSpec = fmt.Sprintf("%s%s%s", appConfig.UrlPrefix, swagger.NativePrefix, swagger.SwaggerSpecFile)
	}
	router := gin.Default()
	// Serve Swagger UI
	router.GET(swagger.SwaggerEndpointPath+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// Security Section
	if appConfig.JWTConfig.EnableSecurityLogin {
		err = swagger.GenerateSecuritySwaggerSpec(appConfig)
	}
	doc := &appSwagger{Config: appConfig}
	swag.Register("swagger", doc)
	return router, err
}

func (swagger Config) getSecuritySpecs(appConfig config.AppConfig) (map[string]interface{}, error) {
	securitySpecs := make(map[string]interface{})
	files, err := os.ReadDir(fmt.Sprintf("%s/%s/", appConfig.StaticContent, appConfig.SecuritySpecsPath))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		filePath := fmt.Sprintf("%s/%s/%s", appConfig.StaticContent, appConfig.SecuritySpecsPath, file.Name())
		content, err := utils.GetFileAsString(filePath)
		if err != nil {
			return nil, err
		}
		var spec interface{}
		err = json.Unmarshal([]byte(content), &spec)
		if err != nil {
			return nil, err
		}
		securitySpecs[file.Name()] = spec
	}
	return securitySpecs, nil
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func (swagger Config) GenerateSecuritySwaggerSpec(appConfig config.AppConfig) error {
	securitySpecs, err := swagger.getSecuritySpecs(appConfig)
	if err != nil {
		return err
	}
	swaggerFilePath := fmt.Sprintf("%s/swagger/swagger.json", appConfig.StaticContent)
	data, err := os.ReadFile(swaggerFilePath)
	if err != nil {
		return err
	}
	var swaggerData map[string]interface{}
	err = json.Unmarshal(data, &swaggerData)
	if err != nil {
		return err
	}

	securityTags := securitySpecs["security_tags.json"].(map[string]interface{})
	swaggerData["tags"] = append(swaggerData["tags"].([]interface{}), securityTags)
	swaggerData["paths"].(map[string]interface{})["/login"] = securitySpecs["login_spec.json"]
	swaggerData["paths"].(map[string]interface{})["/refresh"] = securitySpecs["refresh_spec.json"]
	swaggerData["paths"].(map[string]interface{})["/logout"] = securitySpecs["logout_spec.json"]
	swaggerData["paths"].(map[string]interface{})["/logout-refresh-token"] = securitySpecs["logout_refresh_spec.json"]
	swaggerData["components"].(map[string]interface{})["schemas"].(map[string]interface{})["login"] = securitySpecs["login_scheme.json"]
	ensureTags := []string{"Action", "Project"}
	securityType := securitySpecs["security_type.json"]
	security401Response := securitySpecs["security_unauthorized_response.json"]
	security403Response := securitySpecs["security_forbidden_response.json"]
	securityCrsf := securitySpecs["security_csrf.json"]
	paths := swaggerData["paths"].(map[string]interface{})
	for path := range paths {
		for method := range paths[path].(map[string]interface{}) {
			if isEndpointSwaggerProtected(method, path, appConfig) {
				tags := paths[path].(map[string]interface{})[method].(map[string]interface{})["tags"].([]interface{})
				for _, tag := range tags {
					if contains(ensureTags, tag.(string)) {
						paths[path].(map[string]interface{})[method].(map[string]interface{})["security"] = securityType
						paths[path].(map[string]interface{})[method].(map[string]interface{})["responses"].(map[string]interface{})["401"] = security401Response
						paths[path].(map[string]interface{})[method].(map[string]interface{})["responses"].(map[string]interface{})["403"] = security403Response
						if method == "post" || method == "put" || method == "patch" || method == "delete" {
							parameters, ok := paths[path].(map[string]interface{})[method].(map[string]interface{})["parameters"].([]interface{})
							if ok {
								parameters = append(parameters, securityCrsf)
								paths[path].(map[string]interface{})[method].(map[string]interface{})["parameters"] = parameters
							} else {
								paths[path].(map[string]interface{})[method].(map[string]interface{})["parameters"] = []interface{}{securityCrsf}
							}
						}
					}
				}
			}
		}
	}
	swaggerSecurityFilePath := fmt.Sprintf("%s/swagger/swagger_security.json", appConfig.StaticContent)
	marshal, err := json.Marshal(swaggerData)
	if err != nil {
		return err
	}
	err = os.WriteFile(swaggerSecurityFilePath, marshal, 0644)
	if err != nil {
		return err
	}
	return nil
}
