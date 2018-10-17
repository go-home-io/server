package settings

import (
	"bytes"
	"html/template"
	"io"
	"os"

	"github.com/go-home-io/server/plugins/common"
)

// ITemplateProvider defines template logic.
type ITemplateProvider interface {
	Process([]byte) []byte
}

// Template engine provider.
type provider struct {
	Logger    common.ILoggerProvider
	functions template.FuncMap
}

// Contains data required for a new template.
type constructTemplate struct {
	Secrets common.ISecretProvider
	Logger  common.ILoggerProvider
}

// Constructs a new template engine.
func newTemplateProvider(ctor *constructTemplate) *provider {
	provider := &provider{
		Logger: ctor.Logger,
	}

	provider.functions = template.FuncMap{
		"env": provider.getEnvVariable,
	}

	if ctor.Secrets != nil {
		provider.functions["sec"] = ctor.Secrets.Get
	}

	return provider
}

// Process validates config data and applying various functions to allow reading from
// environment variables, secrets store etc.
func (p *provider) Process(rawFile []byte) []byte {
	tpl, err := template.New("go-home").Funcs(p.functions).Parse(string(rawFile))
	if err != nil {
		p.Logger.Fatal("Failed to parse template", err, common.LogSystemToken, logSystem)
	}
	b := bytes.Buffer{}
	err = tpl.Execute(io.Writer(&b), nil)
	if err != nil {
		p.Logger.Fatal("Failed to execute template", err, common.LogSystemToken, logSystem)
	}

	return b.Bytes()
}

// Returns environment variable.
func (p *provider) getEnvVariable(name string) string {
	p.Logger.Debug("Template is requesting environment variable",
		common.LogNameToken, name, common.LogSystemToken, logSystem)
	return os.Getenv(name)
}
