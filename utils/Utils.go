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
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
func GetEnvOrDefault(key, defaultVal string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return defaultVal
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
	return filepath.Join(projectDir, projectID)
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

func CheckProcess(processFile, projectID string) error {
	cmd := fmt.Sprintf("ps -Af | grep -w %s", processFile)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("failed to execute command: %s", cmd)
	}
	// Check if the output contains the projectID
	isRunning := strings.Contains(string(out), projectID)
	if isRunning {
		return fmt.Errorf("processing files for project_id '%s'. Try later", projectID)
	}
	return nil
}

func CopyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the destination path by joining the destination directory with the relative path
		destinationPath := filepath.Join(dest, path[len(src):])

		if info.IsDir() {
			// Create the directory in the destination
			err := os.MkdirAll(destinationPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			// Copy the file to the destination
			err := CopyFile(path, destinationPath)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func CopyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func ZipDirectory(srcDir string) (*bytes.Buffer, error) {
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	// Closure to add files to the zip archive
	addFile := func(filePath string, file os.FileInfo) error {
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Create a file header
		fileHeader, err := zip.FileInfoHeader(file)
		if err != nil {
			return err
		}

		fileHeader.Name = filepath.ToSlash(filepath.Join(srcDir, filePath))

		// Create a new zip file entry
		writer, err := zipWriter.CreateHeader(fileHeader)
		if err != nil {
			return err
		}

		// Write file data to the zip entry
		_, err = writer.Write(fileData)
		if err != nil {
			return err
		}

		return nil
	}

	err := filepath.Walk(srcDir, func(filePath string, file os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !file.IsDir() {
			// Add the file to the zip archive
			err := addFile(filePath, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Close the zip writer to finish writing the zip archive
	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return &zipBuffer, nil
}

type FileInfoByModTimeDesc []os.FileInfo

func (f FileInfoByModTimeDesc) Len() int           { return len(f) }
func (f FileInfoByModTimeDesc) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f FileInfoByModTimeDesc) Less(i, j int) bool { return f[i].ModTime().After(f[j].ModTime()) }

type Entity struct {
	Link string
	File os.DirEntry
}
type ResultsEntity []Entity

func (f ResultsEntity) Len() int      { return len(f) }
func (f ResultsEntity) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f ResultsEntity) Less(i, j int) bool {
	infoI, _ := f[i].File.Info()
	infoJ, _ := f[j].File.Info()

	return infoI.ModTime().After(infoJ.ModTime())
}
