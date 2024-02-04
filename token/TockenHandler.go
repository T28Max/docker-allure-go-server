package token

import (
	config "allure-server/globals"
	"allure-server/swagger"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
)

var blacklist = make(map[string]struct{})

func checkIfTokenInBlacklist(c *gin.Context, tokenString string) (bool, error) {
	// Check if JTI (JWT ID) is in the blacklist
	_, inBlacklist := blacklist[tokenString]
	return inBlacklist, nil
}
func invalidTokenLoader(msg string) (int, interface{}) {
	response := map[string]interface{}{
		"meta_data": map[string]interface{}{
			"message": fmt.Sprintf("Invalid Token - %s", msg),
		},
	}
	return http.StatusUnauthorized, response
}

func unauthorizedLoader(msg string) (int, interface{}) {
	response := map[string]interface{}{
		"meta_data": map[string]interface{}{
			"message": msg,
		},
	}
	return http.StatusUnauthorized, response
}

func myExpiredTokenCallback(expiredToken *jwt.Token) (int, interface{}) {
	tokenType := expiredToken.Claims.(jwt.MapClaims)["type"].(string)
	response := map[string]interface{}{
		"meta_data": map[string]interface{}{
			"message":    fmt.Sprintf("The %s token has expired", tokenType),
			"sub_status": 42,
		},
	}
	return http.StatusUnauthorized, response
}

func revokedTokenLoader() (int, interface{}) {
	response := map[string]interface{}{
		"meta_data": map[string]interface{}{
			"message": "Revoked Token",
		},
	}
	return http.StatusUnauthorized, response
}

func jwtRequired(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.ENABLE_SECURITY_LOGIN {
			if swagger.IsEndpointProtected(r.URL.Path) {
				verifyJWTInRequest(w, r)
			}
		}
		fn(w, r)
	}
}

func verifyJWTInRequest(responseWriter http.ResponseWriter, request *http.Request) {

}

func jwtRefreshTokenRequired(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.ENABLE_SECURITY_LOGIN {
			if swagger.IsEndpointProtected(r.URL.Path) {
				verifyJWTRefreshTokenInRequest(w, r)
			}
		}
		fn(w, r)
	}
}

func verifyJWTRefreshTokenInRequest(writer http.ResponseWriter, request *http.Request) {

}

func userLoaderCallback(identity string) interface{} {
	if _, ok := USERS_INFO[identity]; !ok {
		return nil
	}
	return UserAccess{
		UserName: identity,
		Roles:    USERS_INFO[identity].Roles,
	}
}
