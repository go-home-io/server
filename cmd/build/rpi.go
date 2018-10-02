package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
)

const (
	// Build tag.
	buildTag = "buildTag"
	// Secret key tag.
	keyTag = "keyTag"
)

// Serves simple REST proxy for travis.
func main() {
	logger := log.New(os.Stdout, "rpi", log.LstdFlags)

	key := os.Args[1]
	port := 9080
	bintrayUser := os.Args[2]
	bintrayKey := os.Args[3]

	if len(os.Args) > 4 {
		port, _ = strconv.Atoi(os.Args[4])
	}

	cd, _ := os.Getwd()

	h := &handler{
		key:         key,
		bintrayKey:  bintrayKey,
		bintrayUser: bintrayUser,
		shell:       fmt.Sprintf("%s/build_rpi.sh", cd),
		logger:      logger,
	}
	router := mux.NewRouter()
	router.HandleFunc(fmt.Sprintf("/{%s}/{%s}", buildTag, keyTag), h.handle).Methods(http.MethodGet)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	logger.Printf("Started build agent on port %d", port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for range c {
		os.Exit(0)
	}
}

// Requests handler.
type handler struct {
	isBuilding  bool
	key         string
	bintrayKey  string
	bintrayUser string
	shell       string
	logger      *log.Logger
}

// Handles the request.
func (h *handler) handle(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	version := vars[buildTag]
	h.logger.Printf("Build request for TAG %s", version)

	if vars[keyTag] != h.key {
		h.logger.Printf("WARNING: wrong key %s from %s", vars[keyTag], request.RemoteAddr)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if h.isBuilding {
		h.logger.Println("Declining since build is already running")
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	h.logger.Println("Start building")
	h.isBuilding = true
	defer func() {
		h.isBuilding = false
	}()

	cmd := exec.Command("/bin/bash", "-c",
		fmt.Sprintf("%s %s %s %s", h.shell, version, h.bintrayUser, h.bintrayKey))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("BINTRAY_API_USER=%s", h.bintrayUser))
	cmd.Env = append(cmd.Env, fmt.Sprintf("BINTRAY_API_KEY=%s", h.bintrayKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TRAVIS_TAG=%s", version))

	cmdOut, err := cmd.CombinedOutput()

	if 0 != len(cmdOut) {
		ioutil.WriteFile("./"+version+".log", cmdOut, 0644)
	}

	if err != nil {
		h.logger.Printf("Error during the buld: %v", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.logger.Println("Build succeeded")
	writer.WriteHeader(http.StatusOK)
}
