package id4client

import "github.com/gin-gonic/gin"

// Authenticate ..
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, code, status := Introspect(c.Request)

		switch {
		case ok && code == 200:
			c.Next()
		case !ok && code == 200:
			c.AbortWithStatusJSON(401, "inactive")
		default:
			c.AbortWithStatusJSON(code, status)
		}
	}
}
