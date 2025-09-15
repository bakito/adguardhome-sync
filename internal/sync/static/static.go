package static

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed index.html
	index string

	//go:embed favicon.ico
	favicon []byte

	//go:embed logo.svg
	logo []byte

	//go:embed bootstrap.min-5.3.3.js
	bootstrapJS []byte

	// https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css
	//go:embed bootstrap.min-5.3.3.css
	bootstrapCSS []byte

	// https://bootswatch.com/5/darkly/bootstrap.min.css
	//go:embed bootstrap.min-darkly-5.3.css
	bootstrapDarkCSS []byte

	// https://code.jquery.com/jquery-3.7.1.min.js
	//go:embed jquery-3.7.1.min.js
	jquery []byte

	// https://cdn.jsdelivr.net/npm/@popperjs/core@2.9.2/dist/umd/popper.min.js
	//go:embed popper.min-2.9.2.js
	popper []byte

	// https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js
	//go:embed chart.umd.min-4.4.7.js
	chart []byte
)

func handleFavicon(c *gin.Context) {
	c.Data(http.StatusOK, "image/x-icon", favicon)
}

func handleLogo(c *gin.Context) {
	c.Data(http.StatusOK, "image/svg+xml", logo)
}

func Index() string {
	return index
}

func HandleResources(group gin.IRouter, dark bool) {
	group.GET("/favicon.ico", handleFavicon)
	group.GET("/logo.svg", handleLogo)
	group.GET("/lib/jquery.js", handleJS(jquery))
	group.GET("/lib/popper.js", handleJS(popper))
	group.GET("/lib/chart.js", handleJS(chart))
	group.GET("/lib/bootstrap.js", handleJS(bootstrapJS))
	if dark {
		group.GET("/lib/bootstrap.css", handleCSS(bootstrapDarkCSS))
	} else {
		group.GET("/lib/bootstrap.css", handleCSS(bootstrapCSS))
	}
}

func handleJS(bytes []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "application/javascript", bytes)
	}
}

func handleCSS(bytes []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "text/css", bytes)
	}
}
