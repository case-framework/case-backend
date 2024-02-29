package apihelpers

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/gin-gonic/gin"
)

func WriteRoutesToFile(router *gin.Engine, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	routes := router.Routes()
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})

	for _, route := range routes {
		_, err := file.WriteString(fmt.Sprintf("%s\t%s\n", route.Method, route.Path))
		if err != nil {
			log.Fatal(err)
		}
	}

}
