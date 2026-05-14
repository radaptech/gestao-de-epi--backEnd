package helper

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	storage_go "github.com/supabase-community/storage-go"
)

func UploadAssinaturaSupabase(base64str string, token, pasta string) (string, error) {

	if strings.Contains(base64str, ",") {
		base64str = strings.Split(base64str, ",")[1]
	}

	decodificador, err := base64.StdEncoding.DecodeString(base64str)
	if err != nil {

		return "", fmt.Errorf("erro ao decodificar string. %w", err)
	}

	supabaseUrl := os.Getenv("SUPABASE_URL")
	secretKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucket := os.Getenv("SUPABASE_BUCKET")

	cliente := storage_go.NewClient(supabaseUrl+"/storage/v1", secretKey, nil)

	contentType := "image/png"
	arquivo := fmt.Sprintf("%s/%s_%d.png", pasta, token, time.Now().Unix())
	opts := storage_go.FileOptions{
		ContentType: &contentType, // força o formato PNG
	}
	_, errS := cliente.UploadFile(bucket, arquivo, bytes.NewReader(decodificador), opts)
	if errS != nil {

		return "", fmt.Errorf("erro ao enviar o arquivo para o supabase, %v", errS)
	}

	urlPublic := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, bucket, arquivo)

	return urlPublic, err
}
