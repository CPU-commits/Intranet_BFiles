package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/CPU-commits/Intranet_BFiles/db"
	"github.com/CPU-commits/Intranet_BFiles/models"
	"github.com/CPU-commits/Intranet_BFiles/res"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func uploadImage() {
	// Send key
	nats_service.Queue("upload_image", func(m *nats.Msg) {
		var key string

		data, err := nats_service.DecodeDataNest(m.Data)
		if err != nil {
			key = string(m.Data)
			fmt.Printf("key: %v\n", key)
		} else {
			keyFromNest := data["data"].(string)
			key = keyFromNest
		}
		file := strings.Split(key, "/")
		filename := file[len(file)-1]
		fileModel, _ := filesModel.NewModel(
			filename,
			key,
			key,
			filename,
			"image",
			"",
			"public",
		)
		insertedId, err := filesModel.Use().InsertOne(db.Ctx, fileModel)
		if err != nil {
			return
		}
		// Inserted
		fileInserted := &models.File{
			ID:          insertedId.InsertedID.(primitive.ObjectID),
			Filename:    filename,
			Key:         key,
			URL:         key,
			Title:       filename,
			Type:        filename,
			User:        fileModel.User,
			Status:      fileModel.Status,
			Permissions: fileModel.Permissions,
			Date:        fileModel.Date,
		}
		jsonData, _ := json.Marshal(res.WrapFileRes(*fileInserted))
		m.Respond(jsonData)
	})
}

func deleteImage() {
	// AWS Key required
	nats_service.Queue("delete_image", func(m *nats.Msg) {
		var key string

		data, err := nats_service.DecodeDataNest(m.Data)
		if err != nil {
			key = string(m.Data)
		} else {
			keyFromNest := data["data"].(string)
			key = keyFromNest
		}
		if err := aws.DeleteFile(key); err != nil {
			return
		}
		m.Respond([]byte("success"))
	})
}

func deleteAWSFile() {
	nats_service.Queue("delete_aws_file", func(m *nats.Msg) {
		var idFile string

		data, err := nats_service.DecodeDataNest(m.Data)
		if err != nil {
			idFile = string(m.Data)
		} else {
			idFileFromNest := data["data"].(string)
			idFile = idFileFromNest
		}
		file, errRes := filesService.getFile(idFile)
		if errRes != nil {
			return
		}
		errRes = filesService.DeleteFile(file.ID.Hex(), file.User.Hex())
		if errRes != nil {
			return
		}
	})
}

func uploadFileClassroom() {
	nats_service.Queue("upload_files_classroom", func(m *nats.Msg) {
		var file FileNats

		err := json.Unmarshal(m.Data, &file)
		if err != nil {
			return
		}
		fileModel, _ := filesModel.NewModel(
			file.Filename,
			file.Key,
			file.Location,
			file.Filename,
			file.Mimetype,
			"",
			"public_classroom",
		)
		insertedId, err := filesModel.Use().InsertOne(db.Ctx, fileModel)
		if err != nil {
			return
		}
		fileData, errRes := filesService.getFile(
			insertedId.InsertedID.(primitive.ObjectID).Hex(),
		)
		if errRes != nil {
			return
		}
		jsonData, _ := json.Marshal(res.WrapFileRes(*fileData))
		m.Respond(jsonData)
	})
}

func getAWSTokenAccess() {
	nats_service.Queue("get_aws_token_access", func(m *nats.Msg) {
		var filesKeys []string

		data, err := nats_service.DecodeDataNest(m.Data)
		if err != nil {
			err = json.Unmarshal(m.Data, &filesKeys)
			if err != nil {
				return
			}
		} else {
			filesKeysFromNest := data["data"].([]interface{})
			for _, key := range filesKeysFromNest {
				filesKeys = append(filesKeys, fmt.Sprintf("%v", key))
			}
		}
		tokensUrls := make([]string, len(filesKeys))

		var errRes error
		var wg sync.WaitGroup
		c := make(chan (int), 10)
		for i, token := range filesKeys {
			wg.Add(1)
			c <- 1
			go func(index int, token string, wg *sync.WaitGroup, errRes *error) {
				defer wg.Done()

				tokenUrl, err := aws.GetFileToken(token)
				if err != nil {
					*errRes = err
					close(c)
					return
				}
				tokensUrls[index] = tokenUrl
				<-c
			}(i, token, &wg, &errRes)
		}
		wg.Wait()
		if errRes != nil {
			return
		}

		jsonData, _ := json.Marshal(tokensUrls)
		m.Respond(jsonData)
	})
}

func getKeyFromIdFile() {
	nats_service.Queue("get_key_from_id_file", func(m *nats.Msg) {
		var idFile string

		data, err := nats_service.DecodeDataNest(m.Data)
		if err != nil {
			idFile = string(m.Data)
		} else {
			idFileFromNest := data["data"].(string)
			idFile = idFileFromNest
		}
		file, errRes := filesService.getFile(idFile)
		if errRes != nil {
			return
		}

		m.Respond([]byte(file.Key))
	})
}

func getPermissionsFiles() {
	nats_service.Queue("get_permissions_files", func(m *nats.Msg) {
		var dataFile FilePermission

		err := json.Unmarshal(m.Data, &dataFile)
		if err != nil {
			return
		}
		permissions := make([]string, len(dataFile.Files))

		var wg sync.WaitGroup
		var errRes *error
		c := make(chan (int), 10)
		for i, file := range dataFile.Files {
			wg.Add(1)
			c <- 1

			go func(index int, idFile string, wg *sync.WaitGroup, errRes *error) {
				defer wg.Done()
				file, err := filesService.getFile(idFile)
				if err != nil {
					*errRes = err.Err
					close(c)
					return
				}
				permissions[index] = file.Permissions

				<-c
			}(i, file, &wg, errRes)
		}
		wg.Wait()

		jsonRes, _ := json.Marshal(permissions)
		m.Respond(jsonRes)
	})
}

func InitFilesNats() {
	uploadImage()
	deleteImage()
	deleteAWSFile()
	uploadFileClassroom()
	getAWSTokenAccess()
	getKeyFromIdFile()
	getPermissionsFiles()
}
