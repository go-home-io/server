package proxy

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
)

// Proxy Bintray downloading attempt.
func main() {
	logger := log.New(os.Stdout, "proxy", log.LstdFlags)
	pluginFolder := os.Args[1]
	port := 9090
	if len(os.Args) > 2 {
		port, _ = strconv.Atoi(os.Args[2])
	}

	p := &proxy{
		logger:       logger,
		pluginFolder: pluginFolder,
		in:           make(chan *downloadRequest, 50),
		done:         make(chan *downloadResponse, 20),
		stop:         make(chan bool, 1),
		downloading:  make(map[string][]chan bool, 50),
	}

	router := mux.NewRouter()
	router.HandleFunc(fmt.Sprintf("/{%s}/{%s}", archName, fileName), p.handle).Methods(http.MethodGet)

	go http.ListenAndServe(fmt.Sprintf(":%d", port), router)
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
func (p *proxy) download(file string, arch string, callback chan *downloadResponse) {
	archName := strings.Replace(file, "_", "/", -1)
	archName = fmt.Sprintf("%s/%s", p.pluginFolder, archName)
	pluginName := strings.Replace(archName, ".tar.gz", "", -1)
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

		defer out.Close()
		downloadURL := fmt.Sprintf("https://dl.bintray.com/go-home-io/%s/%s", arch, file)
		res, err := http.Get(downloadURL) // nolint: gosec
		if err != nil || res.StatusCode != http.StatusOK {
			p.logger.Println("Failed to get " + downloadURL)
			os.Remove(archName) // nolint: gosec
			callback <- dr
			return
		}

		defer res.Body.Close()
		_, err = io.Copy(out, res.Body)
		if err != nil {
			p.logger.Println("Failed to save " + file + ": " + err.Error())
			os.Remove(archName) // nolint: gosec
			callback <- dr
			return
		}
	}

	err := archiver.TarGz.Open(archName, filepath.Dir(pluginName))
	if err != nil {
		p.logger.Println("Failed to un-archive " + file + ": " + err.Error())
		os.Remove(archName) // nolint: gosec
		callback <- dr
		return
	}

	dr.success = true
	callback <- dr
}
