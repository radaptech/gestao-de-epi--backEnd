package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSepararDataURL(t *testing.T) {
	testCases := []struct {
		nome             string
		dataURL          string
		mimeEsperado     string
		conteudoEsperado string
	}{
		{"PNG da assinatura", "data:image/png;base64,QUJD", "image/png", "QUJD"},
		{"JPEG da foto", "data:image/jpeg;base64,QUJD", "image/jpeg", "QUJD"},
		{"Mime em maiúsculo", "data:IMAGE/JPEG;base64,QUJD", "image/jpeg", "QUJD"},
		{"Sem o parâmetro base64", "data:image/png,QUJD", "image/png", "QUJD"},
		{"Base64 puro, sem prefixo", "QUJD", "image/png", "QUJD"},
		{"String vazia", "", "image/png", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.nome, func(t *testing.T) {
			mime, conteudo := separarDataURL(tc.dataURL)
			assert.Equal(t, tc.mimeEsperado, mime)
			assert.Equal(t, tc.conteudoEsperado, conteudo)
		})
	}
}

func TestFormatosAceitos(t *testing.T) {
	// A lista é fechada de propósito: o content-type é servido em URL pública.
	assert.Equal(t, "png", formatosAceitos["image/png"])
	assert.Equal(t, "jpg", formatosAceitos["image/jpeg"])

	_, aceito := formatosAceitos["text/html"]
	assert.False(t, aceito, "text/html não pode ser aceito no bucket público")
}

func TestUploadAssinaturaRejeitaFormatoNaoSuportado(t *testing.T) {
	_, err := UploadAssinaturaSupabase("data:text/html;base64,PHNjcmlwdD4=", "ENT-123", "entregas")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "formato de imagem não suportado")
}
