package aws_s3

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/CPU-commits/Intranet_BFiles/settings"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

type AWSS3 struct {
	sess *session.Session
}

var settingsData = settings.GetSettings()

func NewAWSS3() *AWSS3 {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(settingsData.AWS_REGION),
	}))
	return &AWSS3{
		sess: sess,
	}
}

func (aws_s3 *AWSS3) GetFileToken(key string) (string, error) {
	svc := s3.New(aws_s3.sess)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(settingsData.AWS_BUCKET),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", err
	}
	return urlStr, nil
}

func (aws_s3 *AWSS3) DeleteFile(key string) error {
	svc := s3.New(aws_s3.sess)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(settingsData.AWS_BUCKET),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(settingsData.AWS_BUCKET),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

func (aws_s3 *AWSS3) UploadFile(file *multipart.FileHeader, idUser string) (*s3manager.UploadOutput, string, error) {
	ext := strings.Split(file.Filename, ".")
	uploader := s3manager.NewUploader(aws_s3.sess)
	// To buffer
	openFile, err := file.Open()
	if err != nil {
		return nil, "", err
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, openFile); err != nil {
		return nil, "", err
	}
	fileName := uuid.New()
	key := fmt.Sprintf("user_files/%s/%s.%s", idUser, fileName.String(), ext[len(ext)-1])
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(settingsData.AWS_BUCKET),
		Key:    aws.String(key),
		Body:   buf,
	})
	return result, key, err
}
