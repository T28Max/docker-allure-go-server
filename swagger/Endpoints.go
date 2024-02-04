package swagger

import (
	config "allure-server/globals"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
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

func IsEndpointProtected(endpoint string) bool {
	if config.MAKE_VIEWER_ENDPOINTS_PUBLIC == false {
		return true
	}
	for _, info := range protectedEndpoints {
		if endpoint == info.Endpoint {
			return true
		}
	}
	return false
}
func isEndpointSwaggerProtected(method, path string) bool {
	if config.MAKE_VIEWER_ENDPOINTS_PUBLIC == false {
		return true
	}
	for _, info := range protectedEndpoints {
		if info.Method == method && info.Path == path {
			return true
		}
	}
	return false
}

func getReportsEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["project_id"]
	path := vars["path"]
	redirect := vars["redirect"]
	if redirect == "false" {
		http.ServeFile(w, r, fmt.Sprintf("%s/%s/%s", config.STATIC_CONTENT, projectID, path))
	} else {
		http.Redirect(w, r, fmt.Sprintf("%s/%s/%s", config.STATIC_CONTENT, projectID, path), http.StatusFound)
	}
}
