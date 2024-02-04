package main

import (
	"allure-server/app"
	config2 "allure-server/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	//r := mux.NewRouter()
	//r.HandleFunc("/refresh", refreshEndpoint).Methods("POST")
	//r.HandleFunc("/logout", logoutEndpoint).Methods("DELETE")
	//r.HandleFunc("/login", loginEndpoint).Methods("POST")
	//r.HandleFunc("/logout-refresh-token", logoutRefreshTokenEndpoint).Methods("DELETE")
	//r.HandleFunc("/send-results", sendResultsEndpoint).Methods("POST")
	//r.HandleFunc("/generate-report", generateReportEndpoint).Methods("GET")
	//r.HandleFunc("/clean-results", cleanResultsEndpoint).Methods("GET")
	//r.HandleFunc("/clean-history", cleanHistoryEndpoint).Methods("GET")
	//r.HandleFunc("/projects", createProjectEndpoint).Methods("POST")
	//r.HandleFunc("/projects/{project_id}", deleteProjectEndpoint).Methods("DELETE")
	//
	//r.HandleFunc("/allure-docker-service/{project_id}/{path:.*}", getReportsEndpoint).Methods("GET")
	//r.HandleFunc("/swagger.json", swaggerJSONEndpoint).Methods("GET")
	//r.HandleFunc("/version", versionEndpoint).Methods("GET")
	//r.HandleFunc("/config", configEndpoint).Methods("GET")
	//r.HandleFunc("/select-language", selectLanguageEndpoint).Methods("GET")
	//r.HandleFunc("/latest-report", latestReportEndpoint).Methods("GET")
	//r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	//r.HandleFunc("/swagger.json", swaggerJSONEndpoint)
	//r.HandleFunc("/version", versionEndpoint)
	//r.HandleFunc("/config", configEndpoint).Methods("GET")
	//r.HandleFunc("/select-language", selectLanguageEndpoint).Methods("GET")
	//r.HandleFunc("/latest-report", latestReportEndpoint).Methods("GET")
	//r.HandleFunc("/emailable-report/render", emailableReportRenderEndpoint).Methods("GET")
	//r.HandleFunc("/emailable-report/export", emailableReportExportEndpoint).Methods("GET")
	//r.HandleFunc("/report/export", reportExportEndpoint).Methods("GET")
	//r.HandleFunc("/projects/{project_id}", getProjectEndpoint).Methods("GET")
	//r.HandleFunc("/projects", getProjectsEndpoint).Methods("GET")
	//r.HandleFunc("/projects/search", getProjectsSearchEndpoint).Methods("GET")
	//r.HandleFunc("/projects/{project_id}/reports/{path}", getReportsEndpoint).Methods("GET")
	//handler := Default().Handler(r)
	//http.ListenAndServe(":8080", handler)
	//err := os.Setenv("STATIC_CONTENT", "static")
	//if err != nil {
	//	panic(err)
	//}
	appConfig := config2.DefaultConfig()

	a := app.NewApp(appConfig)
	//config.ReadConfig()
	router := gin.Default()

	// Your other routes go here

	endpoint := []gin.HandlerFunc{a.SwaggerJSONEndpoint}
	router.GET("/swagger.json", endpoint...)
	router.GET("/allure-docker-service/swagger.json", endpoint...)

	if appConfig.DevMode {
		log.Println("Starting in DEV_MODE")

	}
	//else {
	//	middleware := alice.New()
	//	handler := http.NewServeMux()
	//
	//	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//		w.WriteHeader(http.StatusOK)
	//		fmt.Fprintf(w, "Hello, World!")
	//	})
	//
	//	waitressServe(handler, middleware)
	//}

	//err := router.Run(appConfig.GetAddress())
	//if err != nil {
	//	panic(err)
	//}
	err := router.Run(fmt.Sprintf("%s:%d", appConfig.Host, appConfig.Port))
	if err != nil {
		return
	}
}

//func waitressServe(handler http.Handler, middleware alice.Chain) {
//	http.Handle("/", middleware.Then(handler))
//	server := &http.Server{
//		Addr:    fmt.Sprintf("%s:%d", HOST, PORT),
//		Handler: http.DefaultServeMux,
//	}
//
//	err := server.ListenAndServe()
//	if err != nil {
//		log.Fatal(err)
//	}
//}
