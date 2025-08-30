package routes

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	uploadDir = "./upload"
	region    = "us-east-1"
)

func createS3Session() (*s3.S3, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(region), /*
				LogLevel: aws.LogLevel(aws.LogDebugWithHTTPBody), */
		},
		/*			SharedConfigState: session.SharedConfigEnable, */

		/* 	Profile:           "manager-sa", */ // especifica el perfil
		/* SharedConfigState: session.SharedConfigEnable, */ // habilita la configuración compartida
	})
	if err != nil {
		return nil, err
	}
	return s3.New(sess), nil
}

func uploadFileToS3(file multipart.File, fileName string) (string, error) {
	s3Client, err := createS3Session()
	if err != nil {
		return "", fmt.Errorf("Error creando la sesión de S3: %v", err)
	}

	// Lee el contenido del archivo a un buffer
	var buffer bytes.Buffer
	size, err := io.Copy(&buffer, file)
	if err != nil {
		return "", fmt.Errorf("Error al leer el archivo: %v", err)
	}

	// Intento de subida a S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(cfg.BucketNameStorage()),
		Key:           aws.String("AUTENTIA/" + fileName),
		Body:          bytes.NewReader(buffer.Bytes()),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("application/octet-stream"),
	})
	if err != nil {
		return "", fmt.Errorf("Error al subir el archivo a S3: %v", err)
	}

	url := fmt.Sprintf(fileName)
	return url, nil
}

func downloadFileFromS3(s3Client *s3.S3, s3Key, localFilePath string) error {

	output, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(cfg.BucketNameStorage()),
		Key:    aws.String("AUTENTIA/" + s3Key),
	})

	if err != nil {
		return fmt.Errorf("error al obtener el archivo de S3: %v", err)
	}
	defer output.Body.Close()

	// Crear el archivo localmente
	localFile, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Errorf("error al crear archivo local: %v", err)
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, output.Body)
	if err != nil {
		return fmt.Errorf("error al escribir archivo local: %v", err)
	}

	return nil
}
