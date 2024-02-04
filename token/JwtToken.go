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
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"time"
)

func InitJWTMiddleware(tls bool) (*jwt.GinJWTMiddleware, error) {
	// Initialize JWT middleware settings
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:           "test zone",
		Key:             []byte("secret key"),
		Timeout:         time.Hour,
		MaxRefresh:      time.Hour * 24,
		IdentityKey:     "id",
		PayloadFunc:     payloadFunc,
		IdentityHandler: identityHandler,
		Authenticator:   authenticator,
		Authorizator:    authorizator,
		Unauthorized:    unauthorized,
		TokenLookup:     "cookie:jwt_token",
		CookieName:      "jwt_token",
		CookieMaxAge:    time.Duration(int(time.Hour.Seconds())),
		CookieHTTPOnly:  true,
		SecureCookie:    tls, // Set to true for HTTPS
		SendCookie:      true,
	})
	if err != nil {
		return nil, err
	}

	return authMiddleware, nil
}

func payloadFunc(data interface{}) jwt.MapClaims {
	// Customize JWT payload
	if v, ok := data.(UserAccess); ok {
		return jwt.MapClaims{
			"name":  v.UserName,
			"roles": v.Roles,
		}
	}
	return jwt.MapClaims{}
}

func identityHandler(c *gin.Context) interface{} {
	// Extract user information from JWT claims
	claims := jwt.ExtractClaims(c)
	return &UserAccess{
		UserName: claims["name"].(string),
		Roles:    claims["roles"].([]string),
	}
}

func authenticator(c *gin.Context) (interface{}, error) {
	// Authenticate user (e.g., check credentials)
	var user UserAccess
	if err := c.ShouldBindWith(&user, binding.Form); err != nil {
		return nil, jwt.ErrMissingLoginValues
	}

	// Add your authentication logic here
	// ...

	// Return authenticated user
	return &user, nil
}

func authorizator(data interface{}, c *gin.Context) bool {
	// Add your authorization logic here
	// ...

	// Return true if authorized, false otherwise
	return true
}
func unauthorized(c *gin.Context, code int, message string) {
	// Handle unauthorized requests
	c.JSON(code, gin.H{"error": message})
}

func protectedHandler(c *gin.Context) {
	// Handler for the protected endpoint
	user := c.MustGet("id").(*UserAccess)
	c.JSON(http.StatusOK, gin.H{"user_id": user.GetUsername(), "message": "This is a protected endpoint"})
}
