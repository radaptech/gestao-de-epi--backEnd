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

// formatosAceitos mapeia o mime declarado no data URL para a extensão do arquivo.
// Só entra aqui o que o canvas do front gera: PNG para a assinatura desenhada e
// JPEG para a foto da câmera. É uma lista fechada de propósito — o content-type
// vai junto para uma URL pública, então aceitar o que o cliente mandar
// permitiria servir HTML pelo nosso domínio.
var formatosAceitos = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
}

// separarDataURL divide "data:image/jpeg;base64,AAAA" no mime e no conteúdo.
// Sem o prefixo, assume PNG — é o formato das assinaturas antigas.
func separarDataURL(dataURL string) (string, string) {
	prefixo, conteudo, temPrefixo := strings.Cut(dataURL, ",")
	if !temPrefixo {
		return "image/png", dataURL
	}

	mime := strings.TrimPrefix(prefixo, "data:")
	mime, _, _ = strings.Cut(mime, ";")

	return strings.ToLower(strings.TrimSpace(mime)), conteudo
}

func UploadAssinaturaSupabase(base64str string, token, pasta string) (string, error) {

	contentType, conteudo := separarDataURL(base64str)

	extensao, aceito := formatosAceitos[contentType]
	if !aceito {
		return "", fmt.Errorf("formato de imagem não suportado: %q", contentType)
	}

	decodificador, err := base64.StdEncoding.DecodeString(conteudo)
	if err != nil {

		return "", fmt.Errorf("erro ao decodificar string. %w", err)
	}

	supabaseUrl := os.Getenv("SUPABASE_URL")
	secretKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucket := os.Getenv("SUPABASE_BUCKET")

	cliente := storage_go.NewClient(supabaseUrl+"/storage/v1", secretKey, nil)

	arquivo := fmt.Sprintf("%s/%s_%d.%s", pasta, token, time.Now().Unix(), extensao)
	opts := storage_go.FileOptions{
		ContentType: &contentType,
	}
	_, errS := cliente.UploadFile(bucket, arquivo, bytes.NewReader(decodificador), opts)
	if errS != nil {

		return "", fmt.Errorf("erro ao enviar o arquivo para o supabase, %v", errS)
	}

	urlPublic := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, bucket, arquivo)

	return urlPublic, nil
}
