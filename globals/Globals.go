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

const (
	SecurityEnabled    = "SECURITY_ENABLED"
	SecurityUser       = "SECURITY_USER"
	SecurityPass       = "SECURITY_PASS"
	SecurityViewerUser = "SECURITY_VIEWER_USER"
	SecurityViewerPass = "SECURITY_VIEWER_PASS"

	Root                   = "ROOT"
	Port                   = "PORT"
	ApiResponseLessVerbose = "API_RESPONSE_LESS_VERBOSE"
	DevMode                = "DEV_MODE"
	OptimizeStorage        = "OPTIMIZE_STORAGE"
	UrlPrefix              = "URL_PREFIX"
	AllureVersion          = "ALLURE_VERSION"

	StaticContent         = "STATIC_CONTENT"
	StaticContentProjects = "STATIC_CONTENT_PROJECTS"

	EmailableReportFileName = "EMAILABLE_REPORT_FILE_NAME"
	EmailableReportCssCdn   = "EMAILABLE_REPORT_CSS_CDN"
	EmailableReportTitle    = "EMAILABLE_REPORT_TITLE"

	LanguageTemplate = "language.html"
	GlobalCss        = "https://stackpath.bootstrapcdn.com/bootswatch/4.3.1/cosmo/bootstrap.css"
	ReportIndexFile  = "index.html"

	/*JWT constants*/

	AccessTokenExpiresInSeconds  = "ACCESS_TOKEN_EXPIRES_IN_SECONDS"
	AccessTokenExpiresInMinutes  = "ACCESS_TOKEN_EXPIRES_IN_MINUTES"
	RefreshTokenExpiresInSeconds = "REFRESH_TOKEN_EXPIRES_IN_SECONDS"
	RefreshTokenExpiresInDays    = "REFRESH_TOKEN_EXPIRES_IN_DAYS"
	JwtSecretKey                 = "JWT_SECRET_KEY"
	Tls                          = "TLS"
)
