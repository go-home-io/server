package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/config"
	"github.com/go-home-io/server/utils"
)

// Default file system config loader.
type fsConfig struct {
	location string
	logger   common.ILoggerProvider
}

// Init is a main entry-point for plugin.
// In this case we're just providing default implementation.
func (c *fsConfig) Init(data *config.InitDataConfig) error {
	c.logger = data.Logger
	loc, ok := data.Options["location"]
	if !ok {
		loc = fmt.Sprintf("%s/configs", utils.GetCurrentWorkingDir())
		c.logger.Info("Using default location", "location", loc)
	}

	c.location = loc
	return nil
}

// Load files from local file system.
func (c *fsConfig) Load() chan []byte {
	fileList := make([]string, 0)
	fError := filepath.Walk(c.location, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			c.logger.Warn("Failed get folder files", common.LogFileToken, path)
			return err
		}
		if f.IsDir() {
			return nil
		}
		fileList = append(fileList, path)
		return nil
	})

	if fError != nil {
		c.logger.Error("Failed to walk through files", fError)
		return nil
	}

	filesChan := make(chan []byte)

	go func() {
		for _, v := range fileList {
			if !config.IsValidConfigFileName(v) {
				continue
			}

			reader, err := os.Open(v)
			if err != nil {
				c.logger.Error("Failed to load config file", err, common.LogFileToken, v)
				c.closeReader(reader, v)
				continue
			}

			fi, err := reader.Stat()
			if err != nil {
				c.logger.Error("Failed to stat config file", err, common.LogFileToken, v)
				c.closeReader(reader, v)
				continue
			}

			fileData := make([]byte, fi.Size())
			_, err = reader.Read(fileData)
			if err != nil {
				c.logger.Error("Failed to read config file", err, common.LogFileToken, v)
				c.closeReader(reader, v)
				continue
			}

			c.logger.Info("Processing config file", common.LogFileToken, v)
			c.closeReader(reader, v)
			filesChan <- fileData
		}

		close(filesChan)
	}()

	return filesChan
}

// Makes an attempt to close a file descriptor.
func (c *fsConfig) closeReader(reader *os.File, name string) {
	if err := reader.Close(); err != nil {
		c.logger.Error("Failed to close file descriptor", err, common.LogFileToken, name)
	}
}
