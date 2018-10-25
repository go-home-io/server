// Package main contains go.sum files checker.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Validates all go.sum files to make sure that packages version are stable.
// nolint: gocyclo
func main() {
	logger := log.New(os.Stdout, "go.sum", log.LstdFlags)
	existing := make(map[string][]string)

	mainFolder := os.Args[1]
	pluginsFolder := os.Args[2]

	errorsFound := false

	gold, dup := readSumFile(fmt.Sprintf("%s/go.sum", mainFolder), logger)
	for k := range gold {
		existing[k] = make([]string, 0)
		existing[k] = append(existing[k], "server")
	}

	if dup {
		errorsFound = true
	}

	fileList := make([]string, 0)
	fError := filepath.Walk(pluginsFolder, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() || f.Name() != "go.sum" {
			return nil
		}

		fileList = append(fileList, path)
		return nil
	})

	if fError != nil {
		logger.Fatalf("Failed to get plugin files: %s", fError.Error())
	}

	fileList = append(fileList, fmt.Sprintf("%s/plugins/go.sum", mainFolder))

	for _, v := range fileList {
		pluginName := v[:len(v)-len("/go.sum")]
		pluginName = pluginName[len(pluginsFolder)-1:]

		pl, dup := readSumFile(v, logger)
		if dup {
			errorsFound = true
		}

		for k, p := range pl {
			ex, ok := gold[k]
			if ok && ex != p {
				logger.Printf("ERR: %s defines a new version of %s: was %s now %s", pluginName, k, ex, p)
				logger.Printf("Previous versions vere defined in %+v", existing[k])
				errorsFound = true
				continue
			}

			gold[k] = p
			_, ok = existing[k]
			if !ok {
				existing[k] = make([]string, 0)
			}

			existing[k] = append(existing[k], pluginName)
		}
	}

	if errorsFound {
		os.Exit(1)
	}
}

// Reads go.sum content.
//noinspection GoUnhandledErrorResult
func readSumFile(fileName string, logger *log.Logger) (map[string]string, bool) {
	result := make(map[string]string)
	dup := false

	file, err := os.Open(fileName) //nolint: gosec
	if err != nil {
		logger.Fatalf("Failed to open file %s", fileName)
	}

	defer file.Close() // nolint: gosec

	scanner := bufio.NewScanner(file)
	ii := 0
	for scanner.Scan() {
		ii++
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if 3 != len(parts) {
			logger.Fatalf("Corrupted file %s, line(%d): %s", fileName, ii, line)
		}

		if parts[0] == "go-home.io/x/server" {
			continue
		}

		version := parts[1]

		if strings.HasSuffix(version, "/go.mod") {
			version = version[:len(version)-len("/go.mod")]
		}

		ex, ok := result[parts[0]]
		if ok && ex != version {
			logger.Printf("ERR: %s has two different version of %s: %s and %s", fileName, parts[0], ex, version)
			dup = true
			continue
		}

		result[parts[0]] = version
	}

	return result, dup
}
