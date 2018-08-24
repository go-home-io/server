// Package main contains build utils.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
	"github.com/vkorn/go-bintray/bintray"
)

const (
	subject = "go-home-io"
	pkg     = "providers"
)

// Uploads plugins to bintray.
func main() {
	logger := log.New(os.Stdout, "build", log.LstdFlags)
	targetFolder := os.Args[1]
	version := os.Args[2]
	arch := os.Args[3]

	client := http.DefaultClient

	fileList := make([]string, 0)
	fError := filepath.Walk(targetFolder, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			logger.Fatal(err.Error())
		}
		if f.IsDir() {
			return nil
		}
		fileList = append(fileList, path)
		return nil
	})

	if fError != nil {
		logger.Fatal(fError.Error())
	}

	c := bintray.NewClient(client, subject, os.Getenv("BINTRAY_API_USER"), os.Getenv("BINTRAY_API_KEY"))
	exists, err := c.PackageExists("", arch, pkg)
	if err != nil {
		logger.Fatal(err.Error())
	}

	if !exists {
		logger.Fatalf("Package [%s] doesn't exist", pkg)
	}

	err = c.CreateVersion("", arch, pkg, version)
	if err != nil {
		logger.Fatal("Failed to create version")
	}

	for _, v := range fileList {
		ext := filepath.Ext(v)
		name := v[0 : len(v)-len(ext)]
		name = strings.Trim(name[len(targetFolder):], "/")
		newName := fmt.Sprintf("%s/%s-%s.so", targetFolder, name, version)

		name = strings.Replace(name, "/", "_", -1)
		name = fmt.Sprintf("%s-%s.so", name, version)

		err = os.Rename(v, newName)
		if err != nil {
			logger.Fatal("Rename error: " + err.Error())
		}

		archName := fmt.Sprintf("%s/%s.tar.gz", targetFolder, name)

		err = archiver.TarGz.Make(archName, []string{newName})
		if err != nil {
			logger.Fatal("Failed to archive the file: " + err.Error())
		}

		logger.Println("Uploading " + archName)

		err = c.UploadFile("", arch, pkg, version,
			"", "", archName, "?override=1", false)

		if err != nil {
			logger.Fatal("Upload error: " + err.Error())
		}
	}

	err = c.Publish("", arch, pkg, version)
	if err != nil {
		logger.Fatal("Publish error: " + err.Error())
	}
}
