package apihelpers

import (
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/gin-gonic/gin"
)

func WriteRoutesToFile(router *gin.Engine, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		slog.Warn("Error creating file for route infos", slog.String("error", err.Error()))
		return
	}
	defer file.Close()
	routes := router.Routes()
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})

	for _, route := range routes {
		_, err := file.WriteString(fmt.Sprintf("%s %s\n", route.Method, route.Path))
		if err != nil {
			slog.Warn("Error writing route info to file", slog.String("error", err.Error()))
		}
	}
}
