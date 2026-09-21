package helper

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// imagemDeTeste gera uma imagem no formato e nas dimensões pedidas.
func imagemDeTeste(t *testing.T, formato string, largura, altura int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, largura, altura))
	img.Set(0, 0, color.Black)

	var buf bytes.Buffer
	var err error
	if formato == "png" {
		err = png.Encode(&buf, img)
	} else {
		err = jpeg.Encode(&buf, img, nil)
	}
	require.NoError(t, err)

	return buf.Bytes()
}

// servidorDeImagem sobe um bucket falso em HTTPS e aponta o http.DefaultClient
// para ele — prepararImagemAssinatura só aceita URL https.
func servidorDeImagem(t *testing.T, caminho string, conteudo []byte, status int) string {
	t.Helper()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Write(conteudo)
	}))
	t.Cleanup(srv.Close)

	clienteOriginal := http.DefaultClient
	http.DefaultClient = srv.Client()
	t.Cleanup(func() { http.DefaultClient = clienteOriginal })

	return srv.URL + caminho
}

func TestPrepararImagemAssinatura(t *testing.T) {
	t.Run("Assinatura do celular, em pé, gira 90 graus", func(t *testing.T) {
		original := imagemDeTeste(t, "png", 10, 20)
		url := servidorDeImagem(t, "/entregas/ENT-ABC_1.png", original, http.StatusOK)

		imagem, formato, ok := prepararImagemAssinatura(url)

		require.True(t, ok)
		assert.Equal(t, extension.Png, formato)

		cfg, _, err := image.DecodeConfig(bytes.NewReader(imagem))
		require.NoError(t, err)
		assert.Equal(t, 20, cfg.Width, "a largura devia virar a altura original")
		assert.Equal(t, 10, cfg.Height, "a altura devia virar a largura original")
	})

	t.Run("Assinatura do desktop, deitada, passa intacta", func(t *testing.T) {
		original := imagemDeTeste(t, "png", 20, 10)
		url := servidorDeImagem(t, "/entregas/ENT-ABC_1.png", original, http.StatusOK)

		imagem, formato, ok := prepararImagemAssinatura(url)

		require.True(t, ok)
		assert.Equal(t, extension.Png, formato)
		assert.True(t, bytes.Equal(original, imagem), "girar aqui espremeria a assinatura na ficha")
	})

	t.Run("Foto deitada passa intacta", func(t *testing.T) {
		original := imagemDeTeste(t, "jpg", 20, 10)
		url := servidorDeImagem(t, "/entregas/ENT-ABC_1.jpg", original, http.StatusOK)

		imagem, formato, ok := prepararImagemAssinatura(url)

		require.True(t, ok)
		assert.Equal(t, extension.Jpg, formato)
		assert.True(t, bytes.Equal(original, imagem), "a foto não pode ser rotacionada nem reencodada")
	})

	t.Run("Foto em pé também passa intacta", func(t *testing.T) {
		// A regra de proporção vale só para a assinatura: girar uma foto em pé
		// deitaria o funcionário na ficha.
		original := imagemDeTeste(t, "jpg", 10, 20)
		url := servidorDeImagem(t, "/entregas/ENT-ABC_1.jpg", original, http.StatusOK)

		imagem, formato, ok := prepararImagemAssinatura(url)

		require.True(t, ok)
		assert.Equal(t, extension.Jpg, formato)
		assert.True(t, bytes.Equal(original, imagem), "foto em pé não pode girar")
	})
}

func TestPrepararImagemAssinaturaRecusaEntradaInvalida(t *testing.T) {
	t.Run("URL vazia", func(t *testing.T) {
		_, _, ok := prepararImagemAssinatura("")
		assert.False(t, ok)
	})

	t.Run("URL sem https", func(t *testing.T) {
		_, _, ok := prepararImagemAssinatura("http://bucket.local/entregas/ENT-ABC_1.png")
		assert.False(t, ok)
	})

	t.Run("Arquivo sumiu do bucket", func(t *testing.T) {
		url := servidorDeImagem(t, "/entregas/ENT-ABC_1.png", nil, http.StatusNotFound)

		_, _, ok := prepararImagemAssinatura(url)
		assert.False(t, ok)
	})
}
