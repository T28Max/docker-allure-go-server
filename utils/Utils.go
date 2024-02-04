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

package utils

import (
	"allure-server/globals"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

func GetFileAsString(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
func IntInRange(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func StringInArray(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func UpdateKey(key string, currentVar *string) {
	val, ok := os.LookupEnv(key)
	if ok {
		*currentVar = val
		log.Printf("Overriding: %s=%s\n", key, val)
	}
}
func UpdateOrDefault0(key string, currentVar *int, list []int) {
	UpdateOrDefault(key, currentVar, list, 0)
}
func UpdateOrDefault(key string, currentVar *int, list []int, defaultVal int) {
	val, ok := os.LookupEnv(key)
	if ok {
		// string to int
		i, err := strconv.Atoi(val)
		if err != nil && IntInRange(i, list) {
			*currentVar = i
			log.Printf("Overriding:  %s=%s\n", key, val)
			return
		}
	}
	*currentVar = defaultVal
	log.Printf("Wrong env var value. Setting %s=%d by default\n", key, defaultVal)
}
func UpdateOrDefaultBool(key string, currentVar *bool, defaultVal bool) {
	val, ok := os.LookupEnv(key)
	if ok {
		// string to int
		i, err := strconv.ParseBool(val)
		if err != nil {
			*currentVar = i
			log.Printf("Overriding:  %s=%s\n", key, val)
			return
		}
	}
	*currentVar = defaultVal
	log.Printf("Wrong env var value. Setting %s=%t by default\n", key, defaultVal)
}
func UpdateOrDefaultFalse(key string, currentVar *bool) {
	UpdateOrDefaultBool(key, currentVar, false)

}

func IsExistentProject(projectID, projectDir string) bool {
	if projectID == "" {
		return false
	}
	projectPath := GetProjectPath(projectID, projectDir)
	_, err := os.Stat(projectPath)
	return err == nil
}

func GetProjectPath(projectID, projectDir string) string {
	if projectDir == "" {
		return projectID
	}
	// Implementation of getting project path
	return fmt.Sprintf("%s/%s", projectDir, projectID)
}
func ResolveProject(projectIdParam string) string {
	projectId := "default"
	if projectIdParam != "" {
		projectId = projectIdParam
	}
	return projectId

}
func GetKey() string {
	// Set JWT secret key
	jwtSecretKey, exists := os.LookupEnv(globals.JwtSecretKey)
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
