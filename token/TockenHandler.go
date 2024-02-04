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

package token

import (
	"allure-server/globals"
	"allure-server/utils"
	jwt "github.com/appleboy/gin-jwt/v2"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type JWTConfig struct {
	EnableSecurityLogin       bool //enables login
	MakeViewerEndpointsPublic bool
	SecurityUser              string
	SecurityPass              string
	SecurityViewerUser        string
	SecurityViewerPass        string
	AdminRoleName             string
	ViewerRoleName            string
	JwtCookieSecure           bool
	UrlScheme                 string // used schema, HTTP/HTTPS
	UsersInfo                 map[string]UserInfo
}

func New(makeViewerEndpointsPublic bool) *JWTConfig {
	return &JWTConfig{
		MakeViewerEndpointsPublic: makeViewerEndpointsPublic,
		EnableSecurityLogin:       false,
		AdminRoleName:             "admin",
		ViewerRoleName:            "viewer",
		JwtCookieSecure:           true,
		UrlScheme:                 "http",
		SecurityUser:              "",
		SecurityPass:              "",
		SecurityViewerUser:        "",
		SecurityViewerPass:        "",
	}

}

func DefaultConfig() JWTConfig {
	return JWTConfig{
		MakeViewerEndpointsPublic: false,
		EnableSecurityLogin:       false,
		AdminRoleName:             "admin",
		ViewerRoleName:            "viewer",
		JwtCookieSecure:           true,
		UrlScheme:                 "http",

		SecurityUser:       "",
		SecurityPass:       "",
		SecurityViewerUser: "",
		SecurityViewerPass: "",
	}
}
func (jwtLocal JWTConfig) UpdateFromEnv() *jwt.GinJWTMiddleware {

	jwtLocal.updateSecurityUser()
	jwtLocal.updateSecurityPass()
	jwtLocal.updateSecurityViewer()
	jwtLocal.updateSecurityEnabled()
	tls, err := jwtLocal.updateTLS()
	if err != nil {
		tls.Key = []byte(utils.GetKey())
		jwtLocal.updateAccessToken(tls)
		jwtLocal.updateRefreshToken(tls)
		return tls
	}
	return nil
}
func (jwtLocal JWTConfig) updateSecurityEnabled() {
	// Check if "SECURITY_ENABLED" environment variable exists
	securityEnabledTmp, exists := os.LookupEnv(globals.SecurityEnabled)
	if exists {
		if enableSecurityLoginTmp, err := strconv.ParseBool(securityEnabledTmp); err != nil {
			log.Printf("Wrong env var value. Setting %s=0 by default. Error: %v\n", globals.SecurityEnabled, err)
		} else if jwtLocal.SecurityUser == "" || jwtLocal.SecurityPass == "" {
			log.Printf("To enable security you need '%s' & '%s' env vars\n", globals.SecurityUser, globals.SecurityPass)
		} else if jwtLocal.SecurityUser == jwtLocal.SecurityViewerUser {
			log.Printf("%s and %s should be different\n", globals.SecurityUser, globals.SecurityViewerUser)
		} else if enableSecurityLoginTmp {
			jwtLocal.EnableSecurityLogin = true
			log.Printf("Enabling Security Login. %s=%t\n", globals.SecurityEnabled, enableSecurityLoginTmp)

			// Populate USERS_INFO with ADMIN user information
			jwtLocal.UsersInfo[jwtLocal.SecurityUser] = UserInfo{
				Pass:  jwtLocal.SecurityPass,
				Roles: []string{jwtLocal.AdminRoleName},
			}

			// If SECURITY_VIEWER_USER and SECURITY_VIEWER_PASS are provided, populate USERS_INFO with VIEWER user information
			if jwtLocal.SecurityViewerUser != "" && jwtLocal.SecurityViewerPass != "" {
				jwtLocal.UsersInfo[jwtLocal.SecurityViewerUser] = UserInfo{
					Pass:  jwtLocal.SecurityViewerPass,
					Roles: []string{jwtLocal.ViewerRoleName},
				}
			}
			return
		}
	}
	log.Printf("Setting %s=0 by default\n", globals.SecurityEnabled)
}

func (jwtLocal JWTConfig) updateSecurityUser() {
	if securityUserTmp, exists := os.LookupEnv(globals.SecurityUser); exists {
		if trimmedUser := strings.TrimSpace(securityUserTmp); trimmedUser != "" {
			jwtLocal.SecurityUser = strings.ToLower(trimmedUser)
			log.Printf("Setting %s\n", globals.SecurityUser)
		}
	}
}
func (jwtLocal JWTConfig) updateSecurityPass() {
	if securityPassTmp, exists := os.LookupEnv(globals.SecurityPass); exists {
		if trimmedPass := strings.TrimSpace(securityPassTmp); trimmedPass != "" {
			jwtLocal.SecurityPass = trimmedPass
			log.Printf("Setting %s\n", globals.SecurityPass)
		}
	}
}
func (jwtLocal JWTConfig) updateSecurityViewer() {
	if !jwtLocal.MakeViewerEndpointsPublic {
		if securityViewerUserTmp, exists := os.LookupEnv(globals.SecurityViewerUser); exists {
			if trimmedViewerUser := strings.TrimSpace(securityViewerUserTmp); trimmedViewerUser != "" {
				jwtLocal.SecurityViewerUser = strings.ToLower(trimmedViewerUser)
				log.Printf("Setting %s\n", globals.SecurityViewerUser)
			}
		}

		if securityViewerPassTmp, exists := os.LookupEnv(globals.SecurityViewerPass); exists {
			if trimmedViewerPass := strings.TrimSpace(securityViewerPassTmp); trimmedViewerPass != "" {
				jwtLocal.SecurityViewerPass = trimmedViewerPass
				log.Printf("Setting %s\n", globals.SecurityViewerPass)

			}
		}
	}
}
func (jwtLocal JWTConfig) CheckAdminAccess(access UserAccess) bool {
	if !jwtLocal.EnableSecurityLogin {
		return true
	}
	return access.CheckAccess(jwtLocal.AdminRoleName)
}

var blacklist = make(map[string]struct{})

func (jwtLocal JWTConfig) updateTLS() (*jwt.GinJWTMiddleware, error) {
	if val, ok := os.LookupEnv(globals.Tls); ok {
		// string to int
		if i, err := strconv.ParseBool(val); err != nil && i {
			jwtLocal.UrlScheme = "https"
			//app.config['JWT_COOKIE_SECURE'] = True
			log.Printf("Enabling %s=%t", globals.Tls, i)
			return InitJWTMiddleware(true)
		}
	}
	log.Printf("Wrong env var value. Setting TLS=0 by default\n")
	return InitJWTMiddleware(false)
}
func (jwtLocal JWTConfig) updateAccessToken(app *jwt.GinJWTMiddleware) {
	accessSec, existsS := os.LookupEnv(globals.AccessTokenExpiresInSeconds)
	accessMin, existsM := os.LookupEnv(globals.AccessTokenExpiresInMinutes)
	if !jwtLocal.EnableSecurityLogin {
		return //skip token expiration configuration
	}
	if existsS == existsM {
		app.Timeout = time.Minute * 15
		log.Printf("One of %s or %s should be pressent. Setting %s by default to 15 mins\n", globals.AccessTokenExpiresInSeconds, globals.AccessTokenExpiresInMinutes, globals.AccessTokenExpiresInMinutes)
		return
	}
	if existsS {
		if accessTokenExpiresInSeconds, err := strconv.Atoi(accessSec); err == nil && accessTokenExpiresInSeconds > 0 {
			seconds := time.Duration(accessTokenExpiresInSeconds) * time.Second
			app.Timeout = seconds
			log.Printf("Setting %s=%d\n", globals.AccessTokenExpiresInSeconds, accessTokenExpiresInSeconds)

		}

	}
	if existsM {
		if accessTokenExpiresInMinutes, err := strconv.Atoi(accessMin); err == nil && accessTokenExpiresInMinutes > 0 {
			minutes := time.Duration(accessTokenExpiresInMinutes) * time.Minute
			app.Timeout = minutes
			log.Printf("Setting %s=%d\n", globals.AccessTokenExpiresInMinutes, accessTokenExpiresInMinutes)

		}

	}
}
func (jwtLocal JWTConfig) updateRefreshToken(app *jwt.GinJWTMiddleware) {
	refreshSec, existsS := os.LookupEnv(globals.RefreshTokenExpiresInSeconds)
	refreshDay, existsD := os.LookupEnv(globals.RefreshTokenExpiresInDays)
	if !jwtLocal.EnableSecurityLogin {
		return //skip token refresh configuration
	}
	if existsS == existsD {
		app.MaxRefresh = time.Minute * 15
		log.Printf("One of %s or %s should be pressent. Setting %s by default to 15 mins\n", globals.RefreshTokenExpiresInSeconds, globals.RefreshTokenExpiresInDays, globals.RefreshTokenExpiresInSeconds)
		return
	}
	if existsS {
		if refreshTokenExpiresInSeconds, err := strconv.Atoi(refreshSec); err == nil && refreshTokenExpiresInSeconds > 0 {
			seconds := time.Duration(refreshTokenExpiresInSeconds) * time.Second
			app.MaxRefresh = seconds
			log.Printf("Setting %s=%d\n", globals.RefreshTokenExpiresInSeconds, refreshTokenExpiresInSeconds)

		}

	}
	if existsD {
		if refreshTokenExpiresInDays, err := strconv.Atoi(refreshDay); err == nil && refreshTokenExpiresInDays > 0 {
			days := time.Duration(refreshTokenExpiresInDays) * time.Hour * 24
			app.MaxRefresh = days
			log.Printf("Setting %s=%d\n", globals.RefreshTokenExpiresInDays, refreshTokenExpiresInDays)

		}

	}
}

//func checkIfTokenInBlacklist(c *gin.Context, tokenString string) (bool, error) {
//	// Check if JTI (JWT ID) is in the blacklist
//	_, inBlacklist := blacklist[tokenString]
//	return inBlacklist, nil
//}
//func invalidTokenLoader(msg string) (int, interface{}) {
//	response := map[string]interface{}{
//		"meta_data": map[string]interface{}{
//			"message": fmt.Sprintf("Invalid Token - %s", msg),
//		},
//	}
//	return http.StatusUnauthorized, response
//}
//
//func unauthorizedLoader(msg string) (int, interface{}) {
//	response := map[string]interface{}{
//		"meta_data": map[string]interface{}{
//			"message": msg,
//		},
//	}
//	return http.StatusUnauthorized, response
//}
//
//func myExpiredTokenCallback(expiredToken *jwt.Token) (int, interface{}) {
//	tokenType := expiredToken.Claims.(jwt.MapClaims)["type"].(string)
//	response := map[string]interface{}{
//		"meta_data": map[string]interface{}{
//			"message":    fmt.Sprintf("The %s token has expired", tokenType),
//			"sub_status": 42,
//		},
//	}
//	return http.StatusUnauthorized, response
//}
//
//func revokedTokenLoader() (int, interface{}) {
//	response := map[string]interface{}{
//		"meta_data": map[string]interface{}{
//			"message": "Revoked Token",
//		},
//	}
//	return http.StatusUnauthorized, response
//}
//
//func jwtRequired(fn http.HandlerFunc) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		if config.ENABLE_SECURITY_LOGIN {
//			if swagger.IsEndpointProtected(r.URL.Path) {
//				verifyJWTInRequest(w, r)
//			}
//		}
//		fn(w, r)
//	}
//}
//
//func verifyJWTInRequest(responseWriter http.ResponseWriter, request *http.Request) {
//
//}
//
//func jwtRefreshTokenRequired(fn http.HandlerFunc) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		if config.ENABLE_SECURITY_LOGIN {
//			if swagger.IsEndpointProtected(r.URL.Path) {
//				verifyJWTRefreshTokenInRequest(w, r)
//			}
//		}
//		fn(w, r)
//	}
//}
//
//func verifyJWTRefreshTokenInRequest(writer http.ResponseWriter, request *http.Request) {
//
//}
//
//func userLoaderCallback(identity string) interface{} {
//	if _, ok := USERS_INFO[identity]; !ok {
//		return nil
//	}
//	return UserAccess{
//		UserName: identity,
//		Roles:    USERS_INFO[identity].Roles,
//	}
//}
