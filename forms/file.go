package forms

type FileForm struct {
	Title string `form:"title" binding:"required,min=3,max=100"`
}

type PermissionsForm struct {
	Permissions string `json:"permissions" binding:"required"`
}
