package helper

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// DadosDevolucaoPdf representa o escopo do documento
type DadosDevolucaoPdf struct {
	NomeEmpresa     string
	NomeFuncionario string
	Matricula       string
	Setor           string
	Cargo           string
	Assinatura      string
	Itens           ItemDevolvidoPdf
}

// ItemDevolvidoPdf representa a linha da tabela (Devolução e a Troca, se houver)
type ItemDevolvidoPdf struct {
	DataDevolucao       configs.DataBr
	NomeEpi             string
	Tamanho             string
	QuantidadeADevolver int32
	Motivo              string
	HouveTroca          bool
	EpiNovo             string // Só preenchido se HouveTroca == true
	TamanhoNovo         string // Só preenchido se HouveTroca == true
	QuantidadeNova      int32  // Só preenchido se HouveTroca == true
}

func CreatePdfDevolucao(dados DadosDevolucaoPdf, auditoria Auditoria, responsavel string) (core.Document, error) {

	// Configuração base da folha
	cfg := config.NewBuilder().
		WithTopMargin(15).
		WithLeftMargin(10).
		WithRightMargin(10).
		Build()

	m := maroto.New(cfg)

	// ==========================================
	// CABEÇALHO DO DOCUMENTO
	// ==========================================
	m.AddRows(
		row.New(10).Add(
			text.NewCol(12, "COMPROVANTE DE DEVOLUÇÃO / TROCA DE EPI", props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 14}),
		),
		row.New(6).Add(text.NewCol(12, "Empresa: "+dados.NomeEmpresa, props.Text{Size: 10, Style: fontstyle.Bold})),
		row.New(6).Add(
			text.NewCol(6, "Funcionario: "+dados.NomeFuncionario, props.Text{Size: 10}),
			text.NewCol(6, "Matricula: "+dados.Matricula, props.Text{Size: 10}),
		),
		row.New(6).Add(
			text.NewCol(6, "Setor: "+dados.Setor, props.Text{Size: 10}),
			text.NewCol(6, "Cargo: "+dados.Cargo, props.Text{Size: 10}),
		),
		row.New(8).Add(text.NewCol(12, "Responsavel pela impressao: "+responsavel, props.Text{Size: 10})),
	)

	estiloBorda := &props.Cell{
		BorderType:      border.Full,
		BorderThickness: 0.2,
	}

	// ==========================================
	// CABEÇALHO DA TABELA
	// ==========================================
	// Redistribuímos as 12 colunas para focar no que importa para a devolução
	m.AddRows(
		row.New(8).Add(
			text.NewCol(2, "Data", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(3, "Item Devolvido", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(1, "Qtd", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(1, "Tam", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(2, "Motivo", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(3, "Substituição (Troca)", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
		),
	)

// CORPO DA TABELA (Item Único)
	// ==========================================
	// Como agora a struct recebe apenas um item, acessamos ele diretamente.
	// (Certifique-se de que sua struct DadosDevolucaoPdf agora tenha "Item ItemDevolvidoPdf" em vez de "Itens []...")
	item := dados.Itens

	// Montando o texto da troca de forma inteligente
	textoTroca := "---"
	if item.HouveTroca {
		textoTroca = fmt.Sprintf("%s (Tam: %s) x%d", truncarTexto(item.EpiNovo, 18), item.TamanhoNovo, item.QuantidadeNova)
	}

	epiFormatado := truncarTexto(item.NomeEpi, 25)
	motivoFormatado := truncarTexto(item.Motivo, 18)
	qtdFormatada := strconv.Itoa(int(item.QuantidadeADevolver))
	
	// Formata a data no padrão brasileiro (DD/MM/YYYY)
	dataFormatada := item.DataDevolucao.Time().Format("02/01/2006")

	// 1. Imprime a linha com os dados reais
	m.AddRows(
		row.New(8).Add(
			text.NewCol(2, dataFormatada, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
			text.NewCol(3, epiFormatado, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
			text.NewCol(1, qtdFormatada, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
			text.NewCol(1, item.Tamanho, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
			text.NewCol(2, motivoFormatado, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
			text.NewCol(3, textoTroca, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
		),
	)

	// 2. Linhas vazias para manter o layout da tabela (padrão de 3 linhas no total). 
	// Como já preenchemos 1 linha real acima, sobram sempre 2 vazias.
	for range 2 {
		m.AddRows(
			row.New(8).Add(
				col.New(2).WithStyle(estiloBorda),
				col.New(3).WithStyle(estiloBorda),
				col.New(1).WithStyle(estiloBorda),
				col.New(1).WithStyle(estiloBorda),
				col.New(2).WithStyle(estiloBorda),
				col.New(3).WithStyle(estiloBorda),
			),
		)
	}

	m.AddRow(7, col.New(12))

	// ==========================================
	// TERMO JURÍDICO (FOCADO EM DEVOLUÇÃO)
	// ==========================================
	m.AddRow(6, text.NewCol(12, "DECLARO QUE:", props.Text{Size: 9, Style: fontstyle.Bold}))

	textoA := "a) Devolvi à EMPRESA, nesta data, o(s) equipamento(s) e material(is) discriminados acima na coluna 'Item Devolvido', nas condições especificadas pelo motivo do recolhimento."
	textoB := "b) Nos casos em que há registro na coluna 'Substituição (Troca)', atesto que recebi os novos equipamentos em perfeitas condições, comprometendo-me a usá-los e zelar pela sua conservação sob pena de sanções disciplinares."
	textoC := "c) Firmo o presente comprovante para que produza os efeitos legais e administrativos cabíveis perante a Segurança do Trabalho desta Instituição."

	m.AddRows(
		row.New(10).Add(text.NewCol(12, textoA, props.Text{Size: 8, Align: align.Left})),
		row.New(12).Add(text.NewCol(12, textoB, props.Text{Size: 8, Align: align.Left})),
		row.New(8).Add(text.NewCol(12, textoC, props.Text{Size: 8, Align: align.Left})),
	)

	m.AddRow(15, col.New(12)) // Respiro antes da assinatura

	// ==========================================
	// ASSINATURA DIGITAL
	// ==========================================
	var assinaturaBytes []byte
	assinaturaValida := false

	if dados.Assinatura != "" && strings.HasPrefix(dados.Assinatura, "https") {
		res, err := http.Get(dados.Assinatura)
		if err == nil && res.StatusCode == http.StatusOK {
			defer res.Body.Close()

			donwload, errResp := io.ReadAll(res.Body)
			if errResp == nil && len(donwload) > 0 {
				bytesRotacionados, errRot := rotacionar90Graus(donwload)
				if errRot == nil {
					assinaturaBytes = bytesRotacionados
				} else {
					assinaturaBytes = donwload
				}
				assinaturaValida = true
			}
		}
	}

	if !assinaturaValida {
		m.AddRow(20,
			col.New(4),
			col.New(4).Add(line.New(props.Line{Thickness: 0.5})),
			col.New(4),
		)
	} else {
		m.AddRow(20,
			col.New(4),
			image.NewFromBytesCol(4, assinaturaBytes, extension.Png, props.Rect{Center: true, Percent: 100}),
			col.New(4),
		)
	}

	m.AddRows(
		row.New(6).Add(
			text.NewCol(12, "Assinatura do Funcionario", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
		),
		row.New(5).Add(
			text.NewCol(12, dados.NomeFuncionario+" - Matricula: "+dados.Matricula, props.Text{Size: 9, Align: align.Center}),
		),
	)

	m.AddRow(10, col.New(12))

	// ==========================================
	// QR CODE E AUDITORIA
	// ==========================================
	qrContent := fmt.Sprintf("DEVOLUÇÃO | Funcionario: %s | Matricula: %s | Setor: %s",
		dados.NomeFuncionario, dados.Matricula, dados.Setor)

	m.AddRow(30,
		col.New(4),
		code.NewQrCol(4, qrContent, props.Rect{Center: true, Percent: 100}),
		col.New(4),
	)

	m.AddRow(3, col.New(12))
	m.AddRows(
		row.New(5).Add(
			text.NewCol(12, "Autenticação Digital", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center}),
		),
		row.New(5).Add(
			text.NewCol(12, "Aponte a câmera do celular para validar a assinatura e a integridade deste documento.", props.Text{Size: 8, Align: align.Center}),
		),
	)

	m.AddRow(7, col.New(12))

	textoAuditoria := fmt.Sprintf("Registro de Auditoria Digital: Documento gerado em %s | IP de Origem: %s", auditoria.DadosServidor, auditoria.Ip)
	m.AddRow(5,
		text.NewCol(12, textoAuditoria, props.Text{
			Size:  7,
			Align: align.Center,
			Style: fontstyle.Italic,
		}),
	)

	return m.Generate()
}
