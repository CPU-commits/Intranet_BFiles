package middlewares

import (
	"net/http"

	"github.com/CPU-commits/Intranet_BFiles/res"
	"github.com/CPU-commits/Intranet_BFiles/services"
	"github.com/gin-gonic/gin"
)

func RolesMiddleware(roles []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, _ := services.NewClaimsFromContext(ctx)
		for _, rol := range roles {
			if rol == claims.UserType {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, &res.Response{
			Success: false,
			Message: "Unauthorized",
		})
	}
}
