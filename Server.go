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
	"allure-server/globals"
	swaggerActions "allure-server/swagger"
	"allure-server/template"
	"allure-server/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	appConfig := config2.DefaultConfig()
	appConfig.JWTConfig.EnableSecurityLogin = true
	appConfig.AllureVersion = "2.5"
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
	//info
	router.GET(swaggerActions.NativePrefix+"/version", getVersion(appConfig))
	//action
	router.GET(swaggerActions.NativePrefix+"/clean-history", cleanHistory(appConfig))
	router.GET(swaggerActions.NativePrefix+"/clean-results", cleanResults(appConfig))
	router.GET(swaggerActions.NativePrefix+"/config", getConfig(appConfig))
	router.GET(swaggerActions.NativePrefix+"/emailable-report/export", exportEmail(appConfig))
	router.GET(swaggerActions.NativePrefix+"/emailable-report/render", renderEmail(appConfig))
	router.GET(swaggerActions.NativePrefix+"/generate-report", generateReport(appConfig))
	router.GET(swaggerActions.NativePrefix+"/latest-report", lastReport(appConfig))
	router.GET(swaggerActions.NativePrefix+"/report/export", reportExport(appConfig))
	router.POST(swaggerActions.NativePrefix+"/send-results", sendResults(appConfig))
	//project
	router.GET(swaggerActions.NativePrefix+"/projects/:project_id", getProject(appConfig))
	router.GET(swaggerActions.NativePrefix+"/projects/:project_id/reports/*path", getReports(appConfig))

	err := router.Run(fmt.Sprintf("%s:%d", appConfig.Host, appConfig.Port))
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func reportExport(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		projectID, found := utils.ExistsProjectID(context, appConfig.ProjectsDirectory, true)
		if !found {
			return
		}
		if !utils.CheckProcessStatus(context, appConfig.GenerateReportProcess, projectID) {
			return
		}
		projectPath := utils.GetProjectPath(projectID, appConfig.ProjectsDirectory)
		temp, err := os.MkdirTemp("", "report")
		if err != nil {
			utils.Error400(context, err)
			err := os.RemoveAll(temp)
			if err != nil {
				utils.Error400(context, err)
				return
			}
			return
		}
		tmpReport := filepath.Join(temp, "allure-report")
		err = utils.CopyDir(filepath.Join(projectPath, "reports", "latest"), tmpReport)
		if err != nil {
			utils.Error400(context, err)
			return
		}
		zipData, err := utils.ZipDirectory(tmpReport)
		if err != nil {
			utils.Error400(context, err)
			err := os.RemoveAll(temp)
			if err != nil {
				utils.Error400(context, err)
				return
			}
			return
		}
		// Set headers for the response
		context.Header("Content-Type", "application/zip")
		context.Header("Content-Disposition", "attachment; filename=allure-docker-service-report.zip")

		// Write the zip data to the response writer
		_, err = io.Copy(context.Writer, zipData)
		if err != nil {
			utils.Error400(context, err)
			//context.String(500, "Error writing zip data to response: %s", err)
			return
		}
		//context.Data()
		//context.DataFromReader(
		//	http.StatusOK,
		//	buffer.Len(),
		//	"application/zip",
		//	buffer,
		//	gin.H{
		//		"Content-Disposition": `attachment; filename*=UTF-8''`+url.QueryEscape("allure-docker-service-report.zip")
		//	}
		//	)
		//context.FileAttachment()

	}

}

func generateReport(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		projectID, found := utils.ExistsProjectID(context, appConfig.ProjectsDirectory, true)
		if !found {
			return
		}
		projectPath := utils.GetProjectPath(projectID, appConfig.ProjectsDirectory)
		resultsProject := filepath.Join(projectPath, "results")
		var files []os.DirEntry

		if !appConfig.ApiResponseLessVerbose {
			f, err := os.ReadDir(resultsProject)
			files = f
			if err != nil {
				utils.Error400(context, err)
				return
			}
		}
		executionName := context.Query("execution_name")
		if executionName == "" {
			executionName = "Execution On Demand"
		}
		executionFrom := context.Query("execution_from")
		executionType := context.Query("execution_type")

		if !utils.CheckProcessStatus(context, appConfig.KeepHistoryProcess, projectID) {
			return
		}
		if !utils.CheckProcessStatus(context, appConfig.GenerateReportProcess, projectID) {
			return
		}
		execStoreResultsProcess := "1"
		if utils.CallProcess(context, appConfig.KeepHistoryProcess, projectID, appConfig.Origin) {
			return
		}
		response, err := exec.Command(appConfig.GenerateReportProcess, execStoreResultsProcess,
			projectID, appConfig.Origin, executionName, executionFrom, executionType).Output()
		if err != nil {
			utils.Error400(context, err)
			return
		}

		if utils.CallProcess(context, appConfig.RenderEmailReportProcess, projectID, appConfig.Origin) {
			return
		}
		buildOrder := "latest"
		for _, line := range strings.Split(string(response), "\n") {
			if strings.HasPrefix(line, "BUILD_ORDER") {
				buildOrder = line[strings.Index(line, ":")+1:]
			}
		}
		reportUrl, err := url.JoinPath(swaggerActions.NativePrefix, projectID, buildOrder, globals.ReportIndexFile)
		if err != nil {
			utils.Error400(context, err)
			return
		}

		body := gin.H{
			"data": gin.H{
				"report_url": reportUrl,
			},
			"meta_data": gin.H{"message": fmt.Sprintf("Report successfully generated for project_id '%s'", projectID)},
		}
		var fName []string
		if files != nil {
			for _, f := range files {
				fName = append(fName, f.Name())
			}
			body["data"].(map[string]interface{})["allure_results_files"] = fName
		}
		context.JSON(http.StatusOK, body)
	}
}

func getConfig(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		version, err := utils.GetFileAsString(appConfig.AllureVersion)
		if err != nil {
			utils.Error400(context, err)
			return
		}
		checkResults := utils.GetEnvOrDefault(globals.CheckResultsEverySeconds, "1")
		keepHistory := utils.GetEnvOrDefault(globals.KeepHistory, "0")
		keepHistoryLatest := utils.GetEnvOrDefault(globals.KeepHistoryLatest, "20")
		tls, e := strconv.ParseBool(utils.GetEnvOrDefault(globals.Tls, "false"))
		if e != nil {
			tls = false
		}
		body := gin.H{"data": gin.H{
			"version":                      strings.TrimSpace(version),
			"dev_mode":                     appConfig.DevMode,
			"check_results_every_seconds":  checkResults,
			"keep_history":                 keepHistory,
			"keep_history_latest":          keepHistoryLatest,
			"tls":                          tls,
			"security_enabled":             appConfig.JWTConfig.EnableSecurityLogin,
			"url_prefix":                   appConfig.UrlPrefix,
			"api_response_less_verbose":    appConfig.ApiResponseLessVerbose,
			"optimize_storage":             appConfig.OptimizeStorage,
			"make_viewer_endpoints_public": appConfig.JWTConfig.MakeViewerEndpointsPublic,
		},
			"meta_data": gin.H{"message": "Config successfully obtained"},
		}

		context.JSON(http.StatusBadRequest, body)
	}
}

func getVersion(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		version, err := utils.GetFileAsString(appConfig.AllureVersion)
		if err != nil {
			utils.Error400(context, err)
			return
		}
		context.JSON(http.StatusOK, gin.H{"data": gin.H{"version": strings.TrimSpace(version)}, "meta_data": gin.H{"message": "Version successfully obtained"}})
	}
}
func cleanHistory(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		projectID, found := utils.ExistsProjectID(context, appConfig.ProjectsDirectory, true)
		if !found {
			return
		}
		if !utils.CheckProcessStatus(context, appConfig.CleanHistoryProcess, projectID) {
			return
		}
		if !utils.CallProcess(context, appConfig.CleanHistoryProcess, projectID, appConfig.Origin) {
			return
		}
		context.JSON(http.StatusOK, gin.H{"meta_data": gin.H{"message": fmt.Sprintf("History successfully cleaned for project_id '%s'", projectID)}})
	}
}
func cleanResults(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		projectID, found := utils.ExistsProjectID(context, appConfig.ProjectsDirectory, true)
		if !found {
			return
		}
		if !utils.CheckProcessStatus(context, appConfig.GenerateReportProcess, projectID) {
			return
		}
		if !utils.CheckProcessStatus(context, appConfig.CleanResultsProcess, projectID) {
			return
		}
		if !utils.CallProcess(context, appConfig.CleanResultsProcess, projectID, appConfig.Origin) {
			return
		}
		context.JSON(http.StatusOK, gin.H{"meta_data": gin.H{"message": fmt.Sprintf("Results successfully cleaned for project_id '%s'", projectID)}})
	}
}
func exportEmail(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		projectID, found := utils.ExistsProjectID(context, appConfig.ProjectsDirectory, true)
		if !found {
			return
		}
		if !utils.CheckProcessStatus(context, appConfig.GenerateReportProcess, projectID) {
			return
		}
		projectPath := utils.GetProjectPath(projectID, appConfig.ProjectsDirectory)
		reportPath := fmt.Sprintf("%s/reports/%s", projectPath, appConfig.EmailableReportFileName)
		context.FileAttachment(reportPath, appConfig.EmailableReportFileName)
	}
}
func getProject(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		projectID, found := utils.ExistsProjectID(context, appConfig.ProjectsDirectory, false)
		if !found {
			return
		}
		reportsPath := fmt.Sprintf("%s/repors", projectID)
		files, err := os.ReadDir(reportsPath)
		if err != nil {
			utils.Error400(context, err)
			return
		}

		var reportsEntity utils.ResultsEntity
		for _, file := range files {
			if !file.IsDir() {
				continue
			}
			//filepath := fmt.Sprintf("%s/%s/index.html", reportsPath, file.Name())
			c := context.Copy()
			entity := utils.Entity{
				Link: c.Request.URL.JoinPath(file.Name(), "index.html").String(),
				File: file,
			}
			reportsEntity = append(reportsEntity, entity)
			//// Make a GET request to the remote resource
			//response, err := http.Get(c.Request.URL.String())
			//if err != nil {
			//	utils.Error400(context, err)
			//	return
			//}
			//defer func(Body io.ReadCloser) {
			//	err := Body.Close()
			//	if err != nil {
			//		utils.Error400(context, err)
			//		return
			//	}
			//}(response.Body)
			//var content []byte
			//_, err = response.Body.Read(content)
			//if err != nil {
			//	utils.Error400(context, err)
			//	return
			//}
			//re := &reportsEntity{
			//	response: response,
			//	file:     file,
			//}
			//reportsEntity = append(reportsEntity, reportsEntity{
			//	response : response,
			//	file : file,
			//})
		}
		sort.Sort(sort.Reverse(sort.Interface(reportsEntity)))
		var reports []string
		var reportsID []string
		lastReport := ""
		for _, entity := range reportsEntity {
			if strings.ToLower(entity.File.Name()) != "latest" {
				reports = append(reports, entity.Link)
				reportsID = append(reportsID, entity.File.Name())
			} else {
				lastReport = entity.Link
			}
		}
		//sort.Sort(sort.Reverse(sort.StringSlice(reportsID)))
		//sort.Sort(sort.Reverse(sort.StringSlice(reports)))
		if lastReport != "" {
			reports = append([]string{lastReport}, reports...)
			reportsID = append([]string{"latest"}, reportsID...)
		}
		body := gin.H{
			"data": gin.H{
				"project": gin.H{
					"id":         projectID,
					"reports":    reports,
					"reports_id": reportsID,
				},
			},
			"meta_data": gin.H{
				"message": "Project successfully obtained",
			},
		}
		context.JSON(http.StatusOK, body)
	}
}
func getReports(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		projectID := context.Param("project_id")
		if context.Query("redirect") == "true" {
			context.Redirect(http.StatusFound, fmt.Sprintf("%s/projects/%s", swaggerActions.NativePrefix, projectID))
			return
		}
		path := context.Param("path")
		// Join the paths
		projectPath, err := url.JoinPath(projectID, "reports", path)
		if err != nil {
			utils.Error400(context, err)
			return
		}
		joinedPath := filepath.Join(appConfig.ProjectsDirectory, projectPath)
		_, err = os.Stat(joinedPath)
		if err != nil {
			context.Redirect(http.StatusFound, fmt.Sprintf("%s/projects/%s", swaggerActions.NativePrefix, projectID))
			return
		}
		context.File(joinedPath)
		//defer context.Redirect(http.StatusFound, fmt.Sprintf("%s/projects/%s", swaggerActions.NativePrefix, projectID))

	}
}
func lastReport(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		projectID, found := utils.ExistsProjectID(context, appConfig.ProjectsDirectory, true)
		if !found {
			return
		}
		if !utils.CheckProcessStatus(context, appConfig.GenerateReportProcess, projectID) {
			return
		}
		context.Redirect(http.StatusFound, utils.GetLatestURL(swaggerActions.NativePrefix, projectID))
	}
}
func renderEmail(appConfig config2.AppConfig) func(context *gin.Context) {
	return func(context *gin.Context) {
		projectID, found := utils.ExistsProjectID(context, appConfig.ProjectsDirectory, true)
		if !found {
			return
		}
		if !utils.CheckProcessStatus(context, appConfig.GenerateReportProcess, projectID) {
			return
		}
		projectPath := utils.GetProjectPath(projectID, appConfig.ProjectsDirectory)
		testcaseLatestReport := fmt.Sprintf("%s/reports/latest/data/test-cases/*.json", projectPath)
		matchingFiles, err := filepath.Glob(testcaseLatestReport)
		if err != nil {
			utils.Error400(context, err)
			return
		}
		// Get file information for each matching file
		var fileInfos []os.FileInfo
		for _, file := range matchingFiles {
			info, err := os.Stat(file)
			if err != nil {
				fmt.Println("Error getting file info:", err)
				continue
			}
			fileInfos = append(fileInfos, info)
		}
		// Sort the list of files by modification time in descending order
		sort.Sort(utils.FileInfoByModTimeDesc(fileInfos))
		var testCases []template.TestCase

		for _, fileName := range fileInfos {
			readFile, err := os.ReadFile(fileName.Name())
			if err != nil {
				utils.Error400(context, err)
				return
			}
			var testCase template.TestCase
			err = json.Unmarshal(readFile, &testCase)
			if err != nil {
				utils.Error400(context, err)
				return
			}
			if testCase.Hidden == "" {
				testCases = append(testCases, testCase)
			}
			//server_url = url_for('latest_report_endpoint', project_id=project_id, _external=True)

		}
		serverUrl := utils.GetLatestURL(swaggerActions.NativePrefix, projectID)
		env, ok := os.LookupEnv(globals.ServerUrl)
		if ok {
			serverUrl = env
		}

		emailableReportPath := fmt.Sprintf("%s/reports/%s", projectPath, appConfig.EmailableReportFileName)
		data := template.EmailTemplateData{
			Title:     appConfig.EmailableReportTitle,
			CSS:       globals.GlobalCss,
			ProjectID: projectID,
			ServerURL: serverUrl,
			TestCases: testCases,
		}
		// Generate email content from the template
		emailContent, err := template.GenerateEmail(data, appConfig.EmailableReportTitle)
		if err != nil {
			utils.Error400(context, err)
			return
		}
		err = os.WriteFile(emailableReportPath, []byte(emailContent), 0644)
		if err != nil {
			utils.Error400(context, err)
			return
		}
		context.File(emailableReportPath)
	}

}
