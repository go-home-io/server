// Package main contains build utils.
package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/vkorn/go-bintray/bintray"
)

const (
	subject = "go-home-io"
	pkg     = "providers"
)

// Uploads plugins to bintray.
func main() {
	targetFolder := os.Args[1]
	version := os.Args[2]
	arch := os.Args[3]

	client := http.DefaultClient

	fileList := make([]string, 0)
	fError := filepath.Walk(targetFolder, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		if f.IsDir() {
			return nil
		}
		fileList = append(fileList, path)
		return nil
	})

	if fError != nil {
		println(fError.Error())
		os.Exit(1)
	}

	c := bintray.NewClient(client, subject, os.Getenv("BINTRAY_API_USER"), os.Getenv("BINTRAY_API_KEY"))
	exists, err := c.PackageExists("", arch, pkg)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if !exists {
		println(fmt.Sprintf("Package [%s] doesn't exist", pkg))
		os.Exit(1)
	}

	err = c.CreateVersion("", arch, pkg, version)
	if err != nil {
		println("Failed to create version")
		os.Exit(1)
	}

	for _, v := range fileList {
		ext := filepath.Ext(v)
		name := v[0 : len(v)-len(ext)]
		name = strings.Trim(name[len(targetFolder):], "/")
		name = strings.Replace(name, "/", "_", -1)
		name = fmt.Sprintf("%s-%s.so", name, version)
		println("Uploading " + name)

		newName := fmt.Sprintf("%s/%s", targetFolder, name)
		err = os.Rename(v, newName)
		if err != nil {
			println("Rename error: " + err.Error())
			os.Exit(1)
		}

		err = c.UploadFile("", arch, pkg, version,
			"", "", newName, "?override=1", false)

		if err != nil {
			println("Upload error: " + err.Error())
			os.Exit(1)
		}
	}

	err = c.Publish("", arch, pkg, version)
	if err != nil {
		println("Publish error: " + err.Error())
		os.Exit(1)
	}
}
