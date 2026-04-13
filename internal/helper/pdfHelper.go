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

type DadosPdf struct {
	NomeEmpresa     string
	NomeFuncionario string
	Matricula       string
	Setor           string
	Cargo           string
	Assinatura      string
	Epi             []DadosEpiPdf
}

type DadosEpiPdf struct {
	Data       configs.DataBr
	NomeEpi    string
	Ca         string
	Descricao  string
	Quantidade int32
	Tamanho    string
}

type Auditoria struct {
	DadosServidor string
	Ip            string
}

func CreatePdf(Dadosfuncionarios DadosPdf, auditoria Auditoria, responsavel string) (core.Document, error) {

	// ==========================================
	// CONFIGURANDO O MAROTO V2
	// ==========================================
	// A V2 usa um Builder para configurar o tamanho da folha e margens
	cfg := config.NewBuilder().
		WithTopMargin(15).
		WithLeftMargin(10).
		WithRightMargin(10).
		Build()

	m := maroto.New(cfg)

	// ==========================================
	// DESENHANDO AS LINHAS (ROWS E COLS)
	// ==========================================
	m.AddRows(
		// Cabeçalho Principal
		row.New(10).Add(
			text.NewCol(12, "COMPROVANTE DE ENTREGA DE EPI", props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 14}),
		),
		// Linha separadora: Criamos uma coluna de tamanho 12 e jogamos a linha dentro

		// Dados do Funcionário
		row.New(6).Add(text.NewCol(12, "Empresa: "+Dadosfuncionarios.NomeEmpresa, props.Text{Size: 10, Style: fontstyle.Bold})),
		row.New(6).Add(
			text.NewCol(6, "Funcionário: "+Dadosfuncionarios.NomeFuncionario, props.Text{Size: 10}),
			text.NewCol(6, "Matrícula: "+Dadosfuncionarios.Matricula, props.Text{Size: 10}),
		),
		row.New(6).Add(
			text.NewCol(6, "Setor: "+Dadosfuncionarios.Setor, props.Text{Size: 10}),
			text.NewCol(6, "Cargo: "+Dadosfuncionarios.Cargo, props.Text{Size: 10}),
		),
		row.New(8).Add(text.NewCol(12, "Responsável pela impressão: "+ responsavel, props.Text{Size: 10})),
	)
	// 1. Criamos a "caneta" que vai desenhar as caixas ao redor das colunas
	estiloBorda := &props.Cell{
		BorderType:      border.Full, // Borda em cima, embaixo, esquerda e direita
		BorderThickness: 0.2,         // Espessura da linha (fina e elegante)
	}

	// 2. O Cabeçalho da Tabela
	// Colocamos o Align.Center para o título ficar bem no meio da caixinha
	m.AddRows(
		row.New(8).Add(

			text.NewCol(2, "Data", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(2, "Quantidade", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(2, "CA", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(2, "EPI", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(3, "Descrição", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
			text.NewCol(1, "Tamanho", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
		),
	)

	// 3. Os 15 espaços em branco (Agora são caixas reais do grid!)
	for _, epi := range Dadosfuncionarios.Epi {

		dataFormatada := epi.Data.Time().Format("02/01/2006")
		quantidadaFormatada := strconv.Itoa(int(epi.Quantidade))
		m.AddRows(
			row.New(8).Add(

				text.NewCol(2, dataFormatada, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
				text.NewCol(2, quantidadaFormatada, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
				text.NewCol(2, epi.Ca, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
				text.NewCol(2, epi.NomeEpi, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
				text.NewCol(3, epi.Descricao, props.Text{Size: 8, Align: align.Left, Top: 2}).WithStyle(estiloBorda),
				text.NewCol(1, epi.Tamanho, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
			),
		)
	}

	// para o layout não "encolher", você pode adicionar um loop para linhas vazias:
	linhasRestantes := 6 - len(Dadosfuncionarios.Epi)
	for range linhasRestantes {
		m.AddRows(
			row.New(8).Add(
				col.New(2).WithStyle(estiloBorda),
				col.New(2).WithStyle(estiloBorda),
				col.New(2).WithStyle(estiloBorda),
				col.New(2).WithStyle(estiloBorda),
				col.New(3).WithStyle(estiloBorda),
				col.New(1).WithStyle(estiloBorda),
			),
		)
	}

	m.AddRow(5, col.New(12))

	// Espaçamento antes do termo jurídico
	// Respiro antes de começar o termo
	m.AddRow(7, col.New(12))

	// O Título do termo
	m.AddRow(6, text.NewCol(12, "DECLARO QUE:", props.Text{Size: 9, Style: fontstyle.Bold}))

	// Dividimos o texto gigante em 3 variáveis limpas (sem "Enters" no meio do código)
	textoA := "a) Recebi nesta data, da EMPRESA acima identificada, minha empregadora, os equipamentos e materiais supra discriminados, os quais desde já comprometo-me sempre a usar na execução das minhas tarefas, zelando pela sua perfeita guarda, conservação, uso e funcionamento como ora os estou recebendo."
	textoB := "b) Estou ciente e de pleno acordo que o descumprimento das condições estabelecidas na letra A supra, acarretará, além da aplicação de penas disciplinares, inclusive do meu contrato laboral, outras sanções previstas em lei."
	textoC := "c) No caso da perda, dano, extravio ou avaria dos equipamentos e/ou materiais referidos na letra \"A\" favor comunicar imediatamente o departamento de Recursos Humanos."

	// Agora adicionamos uma Linha (Row) para CADA parágrafo.
	// Colocamos tamanhos diferentes (12, 10, 8) porque o texto A é maior e precisa de uma "caixa" mais alta para caber.
	m.AddRows(
		row.New(12).Add(text.NewCol(12, textoA, props.Text{Size: 8, Align: align.Left})),
		row.New(10).Add(text.NewCol(12, textoB, props.Text{Size: 8, Align: align.Left})),
		row.New(8).Add(text.NewCol(12, textoC, props.Text{Size: 8, Align: align.Left})),
	)

	m.AddRow(15, col.New(12)) // Respiro antes de assinar

	// 4. O Bloco da Assinatura Digital
	// 1. Decodifica a string Base64 que veio do banco/struct
	// 1. Tratamento do Base64

	var assinaturaBytes []byte
	assinaturaValida := false
	//caso estiver uma urt com o prefixo "http"
	if Dadosfuncionarios.Assinatura != "" && strings.HasPrefix(Dadosfuncionarios.Assinatura, "https"){

		res, err:= http.Get(Dadosfuncionarios.Assinatura) //baixa a imagem no supabase
		if err == nil  && res.StatusCode == http.StatusOK {

			defer res.Body.Close()

			//transforma em  bytes
			donwload, errResp := io.ReadAll(res.Body)
			if errResp == nil && len(donwload) > 0 {

				assinaturaBytes = donwload
				assinaturaValida = true
			}
		}
	}

	// 2. Linha da Assinatura (Dinâmica)
	if !assinaturaValida{
		// Caso não tenha assinatura: Desenha apenas a linha sólida para assinar à mão
		m.AddRow(20,
			col.New(4),
			col.New(4).Add(line.New(props.Line{Thickness: 0.5})),
			col.New(4),
		)
	} else {
		// Caso tenha assinatura: Desenha a imagem centralizada
		m.AddRow(20,
			col.New(4),
			image.NewFromBytesCol(4, assinaturaBytes, extension.Png, props.Rect{Center: true, Percent: 100}),
			col.New(4),
		)
	}

	// 3. Textos da Assinatura (Sempre aparecem, independente de ter imagem ou não)
	m.AddRows(
		row.New(6).Add(
			text.NewCol(12, "Assinatura do Funcionário", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
		),
		row.New(5).Add(
			text.NewCol(12, Dadosfuncionarios.NomeFuncionario+" - Matrícula: "+Dadosfuncionarios.Matricula, props.Text{Size: 9, Align: align.Center}),
		),
	)

	m.AddRow(10, col.New(12))

	m.AddRow(6, col.New(12))

	// 2. O QR Code Centralizado
	qrContent := fmt.Sprintf("Funcionario: %s | Matricula: %s | Setor: %s | Cargo: %s ",
		Dadosfuncionarios.NomeFuncionario, Dadosfuncionarios.Matricula, Dadosfuncionarios.Setor, Dadosfuncionarios.Cargo)

	m.AddRow(30,
		col.New(4), // Pula 4 colunas
		code.NewQrCol(4, qrContent, props.Rect{Center: true, Percent: 100}), // Desenha o QR no meio
		col.New(4), // Pula mais 4 colunas
	)

	m.AddRow(3, col.New(12))
	// 3. A Legenda do QR Code
	m.AddRows(
		// Título da Legenda (em negrito)
		row.New(5).Add(
			text.NewCol(12, "Autenticação Digital", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center}),
		),
		// Subtítulo explicativo
		row.New(5).Add(
			text.NewCol(12, "Aponte a câmera do celular para validar a assinatura e a integridade deste documento.", props.Text{Size: 8, Align: align.Center}),
		),
	)

	m.AddRow(7, col.New(12))
	// 2. Montando o texto do Carimbo de Tempo
	textoAuditoria := fmt.Sprintf("Registro de Auditoria Digital: Documento gerado em %s | IP de Origem: %s", auditoria.DadosServidor, auditoria.Ip)

	// 3. Estampando no rodapé do documento (Fonte pequena e em itálico)
	m.AddRow(5,
		text.NewCol(12, textoAuditoria, props.Text{
			Size:  7,
			Align: align.Center,
			Style: fontstyle.Italic,
		}),
	)

	// ==========================================
	// COMPILANDO E ENVIANDO PARA O FRONT-END
	// ==========================================
	// A V2 facilita muito aqui. O Generate já resolve tudo:
	return m.Generate()
}
