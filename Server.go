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

package main

import (
	config2 "allure-server/config"
	swaggerActions "allure-server/swagger"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func main() {
	appConfig := config2.DefaultConfig()
	appConfig.JWTConfig.EnableSecurityLogin = true
	swaggerCfg := swaggerActions.DefaultConfig()
	router, e := swaggerCfg.Update(appConfig)
	if e != nil {
		panic(e)
	}
	//a := app.NewApp(appConfig)
	////config.ReadConfig()
	//router := gin.Default()
	//
	////endpoint := []gin.HandlerFunc{a.SwaggerJSONEndpoint}
	//router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//router.GET("/allure-docker-service/*any", a.SwaggerJSONEndpoint)
	//
	//if appConfig.DevMode {
	//	log.Println("Starting in DEV_MODE")
	//
	//}
	err := router.Run(fmt.Sprintf("%s:%d", appConfig.Host, appConfig.Port))
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
