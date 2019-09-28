// Package main contains proxy utility.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/mholt/archiver"
)

const (
	// File name API key.
	fileName = "fileName"
	// Arch name API key.
	archName = "arch"
	// Temp folder name
	tempDir = "/tmp/plugins"
)

// Proxy Bintray downloading attempt.
//noinspection GoUnhandledErrorResult
func main() {
	logger := log.New(os.Stdout, "proxy", log.LstdFlags)
	pluginFolder := os.Args[1]
	port := 9090
	if len(os.Args) > 2 {
		port, _ = strconv.Atoi(os.Args[2]) // nolint: gosec
	}

	p := &proxy{
		logger:       logger,
		pluginFolder: pluginFolder,
		in:           make(chan *downloadRequest, 50),
		done:         make(chan *downloadResponse, 20),
		stop:         make(chan bool, 1),
		downloading:  make(map[string][]chan bool, 50),
	}

	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		logger.Fatalf("Failed to create temp folder %s", tempDir)
	}

	router := mux.NewRouter()
	router.HandleFunc(fmt.Sprintf("/{%s}/{%s}", archName, fileName), p.handle).Methods(http.MethodGet)

	go http.ListenAndServe(fmt.Sprintf(":%d", port), router) // nolint: errcheck
	logger.Printf("Started proxy on port %d. Plugins folder is %s", port, pluginFolder)
	go p.wait()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for range c {
		p.stop <- true
		os.Exit(0)
	}
}

// File download request.
type downloadRequest struct {
	file     string
	arch     string
	callback chan bool
}

// File download response.
type downloadResponse struct {
	archNameKey string
	success     bool
}

// Bintray proxy.
type proxy struct {
	sync.Mutex

	pluginFolder string
	in           chan *downloadRequest
	done         chan *downloadResponse
	stop         chan bool
	downloading  map[string][]chan bool
	logger       *log.Logger
}

// Processes proxy logic.
func (p *proxy) wait() {
	for {
		select {
		case <-p.stop:
			return
		case in := <-p.in:
			p.Lock()
			key := fmt.Sprintf("%s_%s", in.arch, in.file)
			_, ok := p.downloading[key]
			if !ok {
				p.logger.Println("New download request for " + in.file)
				p.downloading[key] = make([]chan bool, 0)
				go p.download(in.file, in.arch, p.done)
			} else {
				p.logger.Println(in.file + " is downloading already")
			}
			p.downloading[key] = append(p.downloading[key], in.callback)
			p.Unlock()
		case completed := <-p.done:
			p.Lock()
			p.logger.Printf("Finished with %s. Success: %v", completed.archNameKey, completed.success)
			waiting := p.downloading[completed.archNameKey]
			for _, v := range waiting {
				v <- completed.success
			}
			delete(p.downloading, completed.archNameKey)
			p.Unlock()
		}
	}
}

// Handles REST API request.
func (p *proxy) handle(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	plugin := vars[fileName]

	p.logger.Println("Download request for " + plugin)

	callback := make(chan bool, 1)
	dr := &downloadRequest{
		file:     plugin,
		callback: callback,
		arch:     vars[archName],
	}

	p.in <- dr
	result := <-callback

	if result {
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

// Downloads the file.
//noinspection GoUnhandledErrorResult
func (p *proxy) download(file string, arch string, callback chan *downloadResponse) {
	archName := strings.Replace(file, "_", "/", -1)
	pluginName := strings.Replace(archName, ".tar.gz", "", -1)
	archName = fmt.Sprintf("%s/%s", tempDir, archName)

	dr := &downloadResponse{
		archNameKey: fmt.Sprintf("%s_%s", arch, file),
		success:     false,
	}

	if _, err := os.Stat(archName); err != nil {
		p.logger.Println("Downloading " + file)
		err = os.MkdirAll(filepath.Dir(archName), os.ModePerm)
		if err != nil {
			p.logger.Println("Failed to load " + file + ": " + err.Error())
			callback <- dr
			return
		}
		out, err := os.Create(archName)
		if err != nil {
			p.logger.Println("Failed to load " + file + ": " + err.Error())
			callback <- dr
			return
		}

		defer out.Close() // nolint: errcheck
		downloadURL := fmt.Sprintf("https://dl.bintray.com/go-home-io/%s/%s", arch, file)
		res, err := http.Get(downloadURL) // nolint: gosec
		if err != nil || res.StatusCode != http.StatusOK {
			p.logger.Println("Failed to get " + downloadURL)
			os.Remove(archName) // nolint: gosec, errcheck
			callback <- dr
			return
		}

		defer res.Body.Close() // nolint: errcheck
		_, err = io.Copy(out, res.Body)
		if err != nil {
			p.logger.Println("Failed to save " + file + ": " + err.Error())
			os.Remove(archName) // nolint: gosec, errcheck
			callback <- dr
			return
		}
	}

	tmpPlugin := fmt.Sprintf("%s/%s", tempDir, pluginName)
	err := archiver.TarGz.Open(archName, filepath.Dir(archName))
	if err != nil {
		p.logger.Println("Failed to un-archive " + file + ": " + err.Error())
		os.Remove(tmpPlugin) // nolint: gosec, errcheck
		os.Remove(archName)  // nolint: gosec, errcheck
		callback <- dr
		return
	}

	pluginName = fmt.Sprintf("%s/%s", p.pluginFolder, pluginName)
	err = os.MkdirAll(filepath.Dir(pluginName), os.ModePerm)
	if err != nil {
		p.logger.Println("Failed to create target folder for " + file + ": " + err.Error())
		callback <- dr
		return
	}

	// To avoid load tries from go-home servers
	tmpPluginName := pluginName + "_tmp"
	err = copyFile(tmpPlugin, tmpPluginName)
	if err != nil {
		p.logger.Println("Failed to move " + file + ": " + err.Error())
		os.Remove(tmpPluginName) // nolint: gosec, errcheck
		callback <- dr
		return
	}

	// Much faster than copy on a shared PV.
	err = os.Rename(tmpPluginName, pluginName)
	if err != nil {
		p.logger.Println("Failed to move from temp to final " + file + ": " + err.Error())
		os.Remove(tmpPluginName) // nolint: gosec, errcheck
		os.Remove(pluginName)    // nolint: gosec, errcheck
		callback <- dr
		return
	}

	dr.success = true
	callback <- dr
}

// Copying files for cross-device.
//noinspection GoUnhandledErrorResult
func copyFile(src string, dst string) (err error) {
	source, err := os.Open(src) // nolint: gosec
	if err != nil {
		return err
	}

	defer source.Close() // nolint: errcheck

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(destination, source)
	if closeErr := destination.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}

	sInfo, err := os.Stat(src)
	if err == nil {
		err = os.Chmod(dst, sInfo.Mode())
	}
	return err
}
