package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type FrontendRouteInfo struct {
	// EmbedTitle is used to generate the title information.
	EmbedTitle func(*gin.Context) string
}

var routes = map[string]FrontendRouteInfo{
	"/": {
		EmbedTitle: func(c *gin.Context) string { return "Home" },
	},
	"/ping": {
		EmbedTitle: func(c *gin.Context) string {
			host := c.Query("host")
			if host != "" {
				return "Ping results for " + host
			}
			return "Ping"
		},
	},
	"/traceroute": {
		EmbedTitle: func(c *gin.Context) string {
			host := c.Query("host")
			if host != "" {
				return "Traceroute results for " + host
			}
			return "Traceroute"
		},
	},
	"/whois": {
		EmbedTitle: func(c *gin.Context) string {
			host := c.Query("host")
			if host != "" {
				return "WHOIS results for " + host
			}
			return "WHOIS"
		},
	},
	"/dns": {
		EmbedTitle: func(c *gin.Context) string {
			host := c.Query("host")
			if host != "" {
				return "DNS results for " + host
			}
			return "DNS"
		},
	},
	"/reverse-dns": {
		EmbedTitle: func(c *gin.Context) string {
			ip := c.Query("ip")
			if ip != "" {
				return "Reverse DNS results for " + ip
			}
			return "Reverse DNS"
		},
	},
	"/bgp-route": {
		EmbedTitle: func(c *gin.Context) string {
			ip := c.Query("ip")
			if ip != "" {
				return "BGP route for " + ip
			}
			return "BGP"
		},
	},
}

type NonBlankString string

func (e *NonBlankString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if s == "" {
		return errors.New("non-blank string expected for yaml field")
	}
	*e = NonBlankString(s)
	return nil
}

type region struct {
	// ID is used to define the ID of the region.
	ID NonBlankString `json:"id" yaml:"id"`

	// Name is used to define the name of the region.
	Name NonBlankString `json:"name" yaml:"name"`

	// URL is used to define the region URL.
	URL NonBlankString `json:"url" yaml:"url"`
}

//go:embed template.html
var templateHTMLString string

type assetManifestPartial struct {
	Files map[string]string `json:"files"`
}

func initFrontend(r *gin.Engine, f fs.FS, logger *zap.Logger) {
	// Initialize the template.
	tpl := template.Must(template.New("template").Parse(templateHTMLString))

	// Look for asset-manifest.json and load it.
	file, err := f.Open("asset-manifest.json")
	if err != nil {
		logger.Error("no asset-manifest.json found - frontend will not be rendered", zap.Error(err))
		return
	}
	b, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	var assetsPartial assetManifestPartial
	if err := json.Unmarshal(b, &assetsPartial); err != nil {
		logger.Error("error parsing asset-manifest.json", zap.Error(err))
		return
	}

	// Find the JS/CSS entrypoints.
	var jsEntrypoint, cssEntrypoint string
	for k, v := range assetsPartial.Files {
		if strings.HasSuffix(k, ".js") {
			if jsEntrypoint == "" {
				jsEntrypoint = v
			} else {
				logger.Warn("multiple JS entrypoints found", zap.String("entrypoint", v))
			}
		} else if strings.HasSuffix(k, ".css") {
			if cssEntrypoint == "" {
				cssEntrypoint = v
			} else {
				logger.Warn("multiple CSS entrypoints found", zap.String("entrypoint", v))
			}
		}
	}

	// Return with an error if the JS entrypoint is not found.
	if jsEntrypoint == "" {
		logger.Error("no JS entrypoint found - frontend will not be rendered", zap.String("entrypoint", jsEntrypoint))
		return
	}

	// Return with an error if the CSS entrypoint is not found.
	if cssEntrypoint == "" {
		logger.Error("no CSS entrypoint found - frontend will not be rendered", zap.String("entrypoint", cssEntrypoint))
		return
	}

	// Find the regions blob on the filesystem.
	b, err = os.ReadFile("regions.yml")
	if err != nil {
		logger.Error("error reading regions.yml", zap.Error(err))
		return
	}

	// Attempt to unmarshal the YAML.
	var regions []region
	if err := yaml.Unmarshal(b, &regions); err != nil {
		logger.Error("error parsing regions.yml", zap.Error(err))
		return
	}
	jBlob, err := json.Marshal(regions)
	if err != nil {
		logger.Error("error marshaling regions.yml", zap.Error(err))
		return
	}
	jBlobStr := string(jBlob)

	// Handle each route.
	for k, v := range routes {
		routeInfoCpy := v
		r.GET(k, func(c *gin.Context) {
			title := routeInfoCpy.EmbedTitle(c)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Status(http.StatusOK)
			err := tpl.Execute(c.Writer, map[string]string{
				"JavaScriptPath": jsEntrypoint,
				"CSSPath":        cssEntrypoint,
				"EmbedTitle":     title,
				"Regions":        jBlobStr,
			})
			if err != nil {
				c.Error(err)
			}
		})
	}
}
