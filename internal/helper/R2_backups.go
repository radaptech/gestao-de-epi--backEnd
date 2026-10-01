package helper

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	r2_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client
var presignClient *s3.PresignClient

func InitR2_cloudflare(ctx context.Context) {

	idconta := os.Getenv("R2_IDCLOUDFLARE")
	secretKey := os.Getenv("R2_SECRETKEY")
	keyid := os.Getenv("R2_KEYID")

	config, err := r2_config.LoadDefaultConfig(ctx,
		r2_config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(keyid, secretKey, "")),
		r2_config.WithRegion("auto"))
	if err != nil {
		slog.ErrorContext(ctx, "carregar configurações da aws", "err", err)
		os.Exit(1)
	}

	s3Client = s3.NewFromConfig(config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", idconta))
	})
	presignClient = s3.NewPresignClient(s3Client)

}

func UploadArquivo(ctx context.Context, bucket, key string, corpo io.Reader, contentType string) error {

	if s3Client == nil {
		return fmt.Errorf("R2 não inicializado")
	}

	if bucket == "" {
		return fmt.Errorf("bucket do R2 não configurado")
	}

	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        corpo,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		slog.ErrorContext(ctx, "salvar arquivo no R2", "bucket", bucket, "key", key, "err", err)
		return fmt.Errorf("erro ao salvar no R2")
	}

	return nil
}
