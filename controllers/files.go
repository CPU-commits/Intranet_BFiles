package controllers

import (
	"net/http"

	"github.com/CPU-commits/Intranet_BFiles/forms"
	"github.com/CPU-commits/Intranet_BFiles/res"
	"github.com/CPU-commits/Intranet_BFiles/services"
	"github.com/gin-gonic/gin"
)

// Services
var filesService = services.NewFilesService()

type FilesController struct{}

func (f *FilesController) GetFiles(c *gin.Context) {
	permissions := c.DefaultQuery("permissions", "any")
	if permissions != "private" && permissions != "public" && permissions != "public_classroom" && permissions != "any" {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Success: false,
			Message: "Query permissions must be private, public or public_classroom",
		})
		return
	}

	claims, _ := services.NewClaimsFromContext(c)
	files, err := filesService.GetFiles(permissions, claims.ID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
			Success: false,
		})
		return
	}

	c.JSON(200, &res.Response{
		Success: true,
		Data:    res.WrapFilesRes(files),
	})
}

func (f *FilesController) GetFile(c *gin.Context) {
	idFile := c.Param("idFile")
	claims, _ := services.NewClaimsFromContext(c)

	file, err := filesService.GetFile(idFile, claims.ID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Message: err.Err.Error(),
			Success: false,
		})
		return
	}
	// Response
	response := make(map[string]interface{})
	response["token"] = file

	c.JSON(200, &res.Response{
		Success: true,
		Data:    response,
	})
}

func (f *FilesController) UploadFile(c *gin.Context) {
	var fileData forms.FileForm
	if err := c.ShouldBind(&fileData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Message: err.Error(),
			Success: false,
		})
		return
	}
	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, res.Response{
			Success: false,
			Message: "Ha ocurrido un error tratando de leer el archivo",
		})
		return
	}
	claims, _ := services.NewClaimsFromContext(c)
	// Upload file
	newFile, errRes := filesService.UploadFile(fileData, file, claims.ID)
	if errRes != nil {
		c.AbortWithStatusJSON(errRes.StatusCode, &res.Response{
			Success: false,
			Message: errRes.Err.Error(),
		})
		return
	}

	c.JSON(201, &res.Response{
		Success: true,
		Data:    res.WrapFileRes(*newFile),
	})
}

func (f *FilesController) ChangePermissions(c *gin.Context) {
	var permissions forms.PermissionsForm
	if err := c.BindJSON(&permissions); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if permissions.Permissions != "private" && permissions.Permissions != "public" && permissions.Permissions != "public_classroom" {
		c.AbortWithStatusJSON(http.StatusBadRequest, &res.Response{
			Success: false,
			Message: "permissions must be private, public or public_classroom",
		})
		return
	}
	idFile := c.Param("idFile")
	claims, _ := services.NewClaimsFromContext(c)
	// Change permissions
	err := filesService.ChangePermissions(claims.ID, idFile, permissions.Permissions)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Success: false,
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(200, &res.Response{
		Success: true,
	})
}

func (f *FilesController) DeleteFile(c *gin.Context) {
	idFile := c.Param("idFile")
	claims, _ := services.NewClaimsFromContext(c)

	err := filesService.DeleteFile(idFile, claims.ID)
	if err != nil {
		c.AbortWithStatusJSON(err.StatusCode, &res.Response{
			Success: false,
			Message: err.Err.Error(),
		})
		return
	}

	c.JSON(200, &res.Response{
		Success: true,
	})
}
