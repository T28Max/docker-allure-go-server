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
	"crypto/rand"
	"encoding/hex"
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const infinite = time.Hour * 24 * 360 * 10 //10 years
// User struct represents the user information
//
//	type User struct {
//		ID    string `form:"id" json:"id"`
//		Email string `form:"email" json:"email"`
//	}
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

var USERS_INFO map[string]UserInfo

func updateTLS() (*jwt.GinJWTMiddleware, error) {
	if val, ok := os.LookupEnv("TLS"); ok {
		// string to int
		if i, err := strconv.Atoi(val); err != nil && i == 1 {
			globals.URL_SCHEME = "https"
			//app.config['JWT_COOKIE_SECURE'] = True
			log.Printf("Enabling TLS=%d", i)
			return InitJWTMiddleware(true)
		}
	}
	log.Printf("Wrong env var value. Setting TLS=0 by default\n")
	return InitJWTMiddleware(false)
}

func getKey() string {
	// Set JWT secret key
	jwtSecretKey, exists := os.LookupEnv("JWT_SECRET_KEY")
	if !exists {
		randomBytes := make([]byte, 16)
		_, err := rand.Read(randomBytes)
		if err != nil {
			log.Fatalf("Failed to generate random bytes: %v", err)
		}
		jwtSecretKey = hex.EncodeToString(randomBytes)
	}
	return jwtSecretKey
}
func makeSecure() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := updateTLS()
	// Set JWT secret key
	authMiddleware.Key = []byte(getKey())
	getSecurityEnabled()

	updateAccessToken(authMiddleware, true)
	updateAccessToken(authMiddleware, false)
	updateRefreshToken(authMiddleware, true)
	updateRefreshToken(authMiddleware, false)
	return authMiddleware, err
}
func updateAccessToken(app *jwt.GinJWTMiddleware, isSeconds bool) {
	amount := time.Second
	key := "ACCESS_TOKEN_EXPIRES_IN_"
	if isSeconds {
		key += "SECONDS"
	} else {
		key += "MINUTES"
		amount = time.Minute
	}
	// For ACCESS_TOKEN_EXPIRES_IN_SECONDS
	if accessTokenExpiresInSecondsStr, exists := os.LookupEnv(key); exists {
		if accessTokenExpiresInSeconds, err := strconv.Atoi(accessTokenExpiresInSecondsStr); err == nil {
			if accessTokenExpiresInSeconds > 0 {
				seconds := time.Duration(accessTokenExpiresInSeconds) * amount
				app.Timeout = seconds
				log.Printf("Setting %s=%d\n", key, accessTokenExpiresInSeconds)
			} else {
				app.Timeout = infinite
				log.Println("Disabling ACCESS_TOKEN expiration")
			}
		} else {
			log.Printf("Wrong env var value. Setting %s by default to 15 mins\n", key)
		}
	}
}

func updateRefreshToken(app *jwt.GinJWTMiddleware, isSeconds bool) {
	amount := time.Second
	key := "REFRESH_TOKEN_EXPIRES_IN_"
	if isSeconds {
		key += "SECONDS"
	} else {
		key += "DAYS"
		amount = time.Hour * 24
	}
	// For REFRESH_TOKEN_EXPIRES_IN_SECONDS
	if refreshTokenExpiresInSecondsStr, exists := os.LookupEnv(key); exists {
		if refreshTokenExpiresInSeconds, err := strconv.Atoi(refreshTokenExpiresInSecondsStr); err == nil {
			if refreshTokenExpiresInSeconds > 0 {
				seconds := time.Duration(refreshTokenExpiresInSeconds) * amount
				app.MaxRefresh = seconds
				log.Printf("Setting %s=%d\n", key, refreshTokenExpiresInSeconds)
			} else {
				app.MaxRefresh = infinite
				log.Println("Disabling REFRESH_TOKEN expiration")
			}
		} else {
			log.Printf("Wrong env var value. Setting %s keeps disabled\n", key)
		}
	}
}
func getSecurityEnabled() {
	// Check if "SECURITY_ENABLED" environment variable exists
	if securityEnabledTmp, exists := os.LookupEnv("SECURITY_ENABLED"); exists {
		// Try to convert the value to an integer
		if enableSecurityLoginTmp, err := strconv.Atoi(securityEnabledTmp); err == nil {
			// Additional checks for SECURITY_USER and SECURITY_PASS
			if globals.SECURITY_USER != "" && globals.SECURITY_PASS != "" {
				// Ensure SECURITY_USER and SECURITY_VIEWER_USER are different
				if globals.SECURITY_USER != globals.SECURITY_VIEWER_USER {
					// Check the value of ENABLE_SECURITY_LOGIN_TMP
					if enableSecurityLoginTmp == 1 {
						globals.ENABLE_SECURITY_LOGIN = true
						log.Printf("Enabling Security Login. SECURITY_ENABLED=%d\n", enableSecurityLoginTmp)

						// Populate USERS_INFO with ADMIN user information
						USERS_INFO[globals.SECURITY_USER] = UserInfo{
							Pass:  globals.SECURITY_PASS,
							Roles: []string{globals.ADMIN_ROLE_NAME},
						}

						// If SECURITY_VIEWER_USER and SECURITY_VIEWER_PASS are provided, populate USERS_INFO with VIEWER user information
						if globals.SECURITY_VIEWER_USER != "" && globals.SECURITY_VIEWER_PASS != "" {
							USERS_INFO[globals.SECURITY_VIEWER_USER] = UserInfo{
								Pass:  globals.SECURITY_VIEWER_PASS,
								Roles: []string{globals.VIEWER_ROLE_NAME},
							}
						}
					} else {
						log.Println("Setting SECURITY_ENABLED=0 by default")
					}
				} else {
					log.Println("SECURITY_USER and SECURITY_VIEWER_USER should be different")
					log.Println("Setting SECURITY_ENABLED=0 by default")
				}
			} else {
				log.Println("To enable security you need 'SECURITY_USER' & 'SECURITY_PASS' env vars")
				log.Println("Setting SECURITY_ENABLED=0 by default")
			}
		} else {
			log.Printf("Wrong env var value. Setting SECURITY_ENABLED=0 by default. Error: %v\n", err)
		}
	} else {
		log.Println("Setting SECURITY_ENABLED=0 by default")
	}
}
