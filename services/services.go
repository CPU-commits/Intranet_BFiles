package services

import (
	"github.com/CPU-commits/Intranet_BFiles/aws_s3"
	"github.com/CPU-commits/Intranet_BFiles/models"
	"github.com/CPU-commits/Intranet_BFiles/stack"
)

// Models
var filesModel = new(models.FilesModel)

var aws = aws_s3.NewAWSS3()
var nats_service = stack.NewNats()

// Error Response
type ErrorRes struct {
	Err        error
	StatusCode int
}
