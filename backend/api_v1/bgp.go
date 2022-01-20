package api_v1

import (
	birdsocket "github.com/czerwonk/bird_socket"
	"github.com/gin-gonic/gin"
	"os"
)

func bgp(g *gin.RouterGroup) {
	g.GET("/:ip", func(context *gin.Context) {
		// Get the IP address.
		ip := context.Param("ip")

		conn := birdsocket.NewSocket(os.Getenv("BIRD_SOCKET"))
		defer conn.Close()
		//conn.Query("show route " + prefix + " all")

		println(ip)
		// TODO
	})
}
