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
