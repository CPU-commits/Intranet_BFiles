package res

import (
	"github.com/CPU-commits/Intranet_BFiles/models"
)

type OID struct {
	ID string `json:"$oid"`
}

type Date struct {
	Date int `json:"$date"`
}

type FileRes struct {
	ID          OID    `json:"_id"`
	Filename    string `json:"filename"`
	Key         string `json:"key"`
	URL         string `json:"url"`
	User        OID    `json:"user"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Status      bool   `json:"status"`
	Permissions string `json:"permissions"`
	Date        Date   `json:"date"`
}

func WrapFileRes(file models.File) *FileRes {
	return &FileRes{
		ID: OID{
			ID: file.ID.Hex(),
		},
		Filename: file.Filename,
		Key:      file.Key,
		URL:      file.URL,
		User: OID{
			ID: file.User.Hex(),
		},
		Title:       file.Title,
		Type:        file.Type,
		Status:      file.Status,
		Permissions: file.Permissions,
		Date: Date{
			Date: int(file.Date.Time().Unix()),
		},
	}
}

func WrapFilesRes(files []models.File) []*FileRes {
	var filesRes []*FileRes
	for _, file := range files {
		filesRes = append(filesRes, WrapFileRes(file))
	}
	return filesRes
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"body"`
}
