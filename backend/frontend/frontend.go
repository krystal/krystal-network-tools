package frontend

import (
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type RouteInfo struct {
	// EmbedTitle is used to generate the title information.
	EmbedTitle func(*gin.Context) string
}

var routes = map[string]RouteInfo{
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

type assetManifestPartial struct {
	Files map[string]string `json:"files"`
}

func errorFrontend(r gin.IRouter, err error, message string) {
	for k := range routes {
		r.GET(k, func(c *gin.Context) {
			slog.With("service", "frontend", "message", message).Error("failed to load frontend", "err", err.Error())
			c.String(http.StatusInternalServerError,
				"failed to load frontend - please check console for details")
		})
	}
}

var (
	//go:embed frontend_blobs/*
	blobs embed.FS
	//go:embed template.html
	templateHTMLString string
)

func InitFrontend(r gin.IRouter) {
	// Parse the template HTML.
	tpl := template.Must(template.New("template").Parse(templateHTMLString))
	f, err := fs.Sub(blobs, "frontend_blobs")
	if err != nil {
		errorFrontend(r, err, "failed to open frontend blobs")
		return
	}

	// Find the asset-manifest.json file on the filesystem.
	file, err := f.Open("asset-manifest.json")
	if err != nil {
		errorFrontend(r, err, "failed to open asset-manifest.json")
		return
	}

	// Read the asset-manifest.json file.
	b, err := io.ReadAll(file)
	if err != nil {
		errorFrontend(r, err, "error reading asset-manifest.json")
		return
	}

	// Attempt to unmarshal the JSON manifest file.
	var assetsPartial assetManifestPartial
	if err = json.Unmarshal(b, &assetsPartial); err != nil {
		errorFrontend(r, err, "error parsing asset-manifest.json")
		return
	}

	// Find the JS/CSS entrypoints.
	var jsEntrypoint, cssEntrypoint string
	for k, v := range assetsPartial.Files {
		if strings.HasSuffix(k, ".js") {
			if jsEntrypoint == "" {
				jsEntrypoint = v
			} else {
				slog.Warn("multiple JS entrypoints found", "entrypoint", v)
			}
		} else if strings.HasSuffix(k, ".css") {
			if cssEntrypoint == "" {
				cssEntrypoint = v
			} else {
				slog.Warn("multiple CSS entrypoints found", "entrypoint", v)
			}
		}
	}

	// Return with an error if the JS entrypoint is not found.
	if jsEntrypoint == "" {
		errorFrontend(r, nil, "no JS entrypoint found")
		return
	}

	// Return with an error if the CSS entrypoint is not found.
	if cssEntrypoint == "" {
		errorFrontend(r, nil, "no CSS entrypoint found")
		return
	}

	r.StaticFS("/assets", http.FS(f))
	// Handle each route.
	for k, v := range routes {
		routeInfoCpy := v
		r.GET(k, func(c *gin.Context) {
			title := routeInfoCpy.EmbedTitle(c)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Status(http.StatusOK)
			err := tpl.Execute(c.Writer, map[string]string{
				"JavaScriptPath": "/assets" + jsEntrypoint,
				"CSSPath":        "/assets" + cssEntrypoint,
				"EmbedTitle":     title,
				"Regions":        loadRegions(r),
			})
			if err != nil {
				c.Status(http.StatusInternalServerError)
				c.Error(err)
			}
		})
	}
}

// loadRegions loads the regions.yml file from the filesystem and returns
// a JSON blob of the regions.
// Regions are used by /ping, to determine from where to ping the given host.
// Regions are used by /traceroute, to determine from where to traceroute the given host.
// Regions are used by /bgp, to determine from where to get the BGP route for the given host.
func loadRegions(r gin.IRouter) string {
	// Load regions.yml from the filesystem.
	regionsYAMLString, err := os.ReadFile("regions.yml")
	if err != nil {
		if os.IsNotExist(err) {
			regionsYAMLString = []byte(`- id: local
  name: Local
  url: /`)
		} else {
			errorFrontend(r, err, "error reading regions.yml")
			return ""
		}
	}

	// Attempt to unmarshal the YAML.
	var regions []region
	if err := yaml.Unmarshal([]byte(regionsYAMLString), &regions); err != nil {
		errorFrontend(r, err, "error parsing regions.yml")
		return ""
	}
	jBlob, err := json.Marshal(regions)
	if err != nil {
		errorFrontend(r, err, "error marshaling regions.yml")
		return ""
	}

	return string(jBlob)
}
