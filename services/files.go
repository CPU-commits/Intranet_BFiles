package services

import (
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/CPU-commits/Intranet_BFiles/db"
	"github.com/CPU-commits/Intranet_BFiles/forms"
	"github.com/CPU-commits/Intranet_BFiles/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var filesService *FilesService

type FilesService struct{}

func (f *FilesService) getMatchFile(idUser primitive.ObjectID) bson.D {
	return bson.D{{
		Key: "$match",
		Value: bson.M{
			"user": idUser,
		},
	}}
}

func (f *FilesService) GetFiles(permissions string, idUser string) ([]models.File, *ErrorRes) {
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	pipeline := mongo.Pipeline{
		f.getMatchFile(idObjUser),
	}
	if permissions != "any" {
		pipeline = append(pipeline, bson.D{{
			Key: "$match",
			Value: bson.M{
				"permissions": permissions,
			},
		}})
	}
	var files []models.File

	cursor, err := filesModel.Use().Aggregate(db.Ctx, pipeline)
	if err != nil {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if err := cursor.All(db.Ctx, &files); err != nil {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	return files, nil
}

func (f *FilesService) getFile(idFile string) (*models.File, *ErrorRes) {
	idObjFile, err := primitive.ObjectIDFromHex(idFile)
	if err != nil {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	var file *models.File
	cursor := filesModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "_id",
		Value: idObjFile,
	}})
	if err := cursor.Decode(&file); err != nil {
		if err.Error() == db.NO_SINGLE_DOCUMENT {
			return nil, &ErrorRes{
				Err:        err,
				StatusCode: http.StatusNotFound,
			}
		}
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	return file, nil
}

func (f *FilesService) GetFile(idFile, idUser string) (string, *ErrorRes) {
	file, err := f.getFile(idFile)
	if err != nil {
		return "", err
	}
	if !file.Status {
		return "", &ErrorRes{
			Err:        errors.New("el archivo está eliminado"),
			StatusCode: http.StatusConflict,
		}
	}
	if idUser != file.User.Hex() && file.Permissions == "private" {
		return "", &ErrorRes{
			Err:        errors.New("el archivo es privado"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	if file.Permissions == "public_classroom" {
		return "", &ErrorRes{
			Err:        errors.New("no se puede determinar si el archivo pertenece a una aula virtual"),
			StatusCode: http.StatusBadRequest,
		}
	}
	urlStr, errRes := aws.GetFileToken(file.Key)
	if errRes != nil {
		return "", &ErrorRes{
			Err:        errRes,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	return urlStr, nil
}

func (f *FilesService) uploadFileDB(
	filename,
	key,
	url,
	title,
	typeFile,
	idUser string,
) (*models.File, *ErrorRes) {
	fileModel, err := filesModel.NewModel(filename, key, url, title, typeFile, idUser, "private")
	if err != nil {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idFile, err := filesModel.Use().InsertOne(db.Ctx, fileModel)
	if err != nil {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return &models.File{
		ID:          idFile.InsertedID.(primitive.ObjectID),
		Filename:    filename,
		Key:         key,
		URL:         url,
		Title:       title,
		Type:        typeFile,
		User:        fileModel.User,
		Status:      fileModel.Status,
		Permissions: fileModel.Permissions,
		Date:        fileModel.Date,
	}, nil
}

func (f *FilesService) UploadFile(
	fileData forms.FileForm,
	file *multipart.FileHeader,
	idUser string,
) (*models.File, *ErrorRes) {
	ext := strings.Split(file.Filename, ".")
	filename := fmt.Sprintf("%s.%s", fileData.Title, ext[len(ext)-1])
	// Check if exists
	var fileCheck *models.File
	cursor := filesModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "filename",
		Value: filename,
	}})
	if err := cursor.Decode(&fileCheck); err != nil && err.Error() != db.NO_SINGLE_DOCUMENT {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if fileCheck != nil {
		return nil, &ErrorRes{
			Err:        errors.New("ya tiene un archivo con este mismo nombre"),
			StatusCode: http.StatusBadRequest,
		}
	}
	// Upload file AWS S3
	out, key, err := aws.UploadFile(file, idUser)
	if err != nil {
		return nil, &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Upload db
	return f.uploadFileDB(
		filename,
		key,
		out.Location,
		fileData.Title,
		mime.TypeByExtension("."+ext[len(ext)-1]),
		idUser,
	)
}

func (f *FilesService) ChangePermissions(idUser, idFile, permissions string) *ErrorRes {
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idObjFile, err := primitive.ObjectIDFromHex(idFile)
	if err != nil {
		return &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	file, errRes := f.getFile(idFile)
	if errRes != nil {
		return errRes
	}
	if file.User != idObjUser {
		return &ErrorRes{
			Err:        errors.New("el archivo le pertenece a otro usuario"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	_, err = filesModel.Use().UpdateByID(db.Ctx, idObjFile, bson.D{{
		Key: "$set",
		Value: bson.M{
			"permissions": permissions,
		},
	}})
	if err != nil {
		return &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	return nil
}

func (f *FilesService) DeleteFile(idFile, idUser string) *ErrorRes {
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idObjFile, err := primitive.ObjectIDFromHex(idFile)
	if err != nil {
		return &ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	file, errRes := f.getFile(idFile)
	if errRes != nil {
		return errRes
	}
	if file.User != idObjUser {
		return &ErrorRes{
			Err:        errors.New("el archivo le pertenece a otro usuario"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	err = filesModel.Transaction(func() error {
		_, err = filesModel.Use().UpdateByID(db.Ctx, idObjFile, bson.D{{
			Key: "$set",
			Value: bson.M{
				"status": false,
			},
		}})
		if err != nil {
			return err
		}

		err = aws.DeleteFile(file.Key)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return &ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return nil
}

func NewFilesService() *FilesService {
	if filesService == nil {
		filesService = &FilesService{}
	}
	return filesService
}
