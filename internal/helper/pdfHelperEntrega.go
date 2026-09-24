package helper

import (
	"fmt"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"

	"bytes"
	i90 "image"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/disintegration/imaging"
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
	Cpf             string
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

// rotacionarSeEstiverEmPe gira a assinatura 90° só quando ela veio em pé.
//
// O canvas de assinatura acompanha o formato da tela: no celular em pé ele sai
// alto, e sem girar a firma chega deitada na ficha. No desktop ele já sai
// deitado — girar ali espremeria a assinatura numa tira de ~11mm de largura,
// porque o bloco da ficha tem só 20mm de altura.
func rotacionarSeEstiverEmPe(imgBytes []byte) ([]byte, error) {
	img, _, err := i90.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}

	if limites := img.Bounds(); limites.Dy() <= limites.Dx() {
		return imgBytes, nil
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, imaging.Rotate90(img)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// prepararImagemAssinatura baixa a comprovação da entrega/devolução no bucket e
// devolve os bytes prontos para o PDF, junto do formato real do arquivo.
//
// A assinatura é desenhada deitada na tela do celular, então o PNG do canvas
// precisa girar 90°. A foto da câmera (.jpg) já vem na orientação certa —
// girar ela deixaria o funcionário de lado na ficha.
func prepararImagemAssinatura(url string) ([]byte, extension.Type, bool) {
	if url == "" || !strings.HasPrefix(url, "https") {
		return nil, "", false
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, "", false
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, "", false
	}

	imagem, errResp := io.ReadAll(res.Body)
	if errResp != nil || len(imagem) == 0 {
		return nil, "", false
	}

	// Só a assinatura (.png, desenhada no canvas) pode girar. A foto já vem na
	// orientação em que foi tirada, e uma foto em pé girada deixa o funcionário
	// deitado na ficha.
	if !strings.HasSuffix(strings.ToLower(url), ".png") {
		return imagem, extension.Jpg, true
	}

	rotacionada, errRot := rotacionarSeEstiverEmPe(imagem)
	if errRot != nil {
		// Sem conseguir girar, é melhor a ficha sair com a assinatura deitada
		// do que sem assinatura nenhuma.
		return imagem, extension.Png, true
	}

	return rotacionada, extension.Png, true
}

// adicionarComprovacaoRecebimento desenha o bloco que fecha a ficha: a foto da
// câmera, a assinatura desenhada na tela, ou a linha em branco para assinar à
// mão quando não veio nenhuma das duas.
//
// A foto ganha um bloco maior de propósito. No espaço da assinatura (4 de 12
// colunas por 20mm) um retrato 4:3 encolhe para 27x20mm e não dá para
// reconhecer ninguém — inútil como comprovação de entrega.
func adicionarComprovacaoRecebimento(m core.Maroto, imagem []byte, formato extension.Type, temImagem bool, nome, matricula string) {
	legenda := "Assinatura do Funcionario"

	switch {
	case !temImagem:
		m.AddRow(15, col.New(12)) // respiro antes da linha de assinar
		m.AddRow(20,
			col.New(4),
			col.New(4).Add(line.New(props.Line{Thickness: 0.5})),
			col.New(4),
		)

	case formato == extension.Png:
		m.AddRow(15, col.New(12))
		m.AddRow(20,
			col.New(4),
			image.NewFromBytesCol(4, imagem, formato, props.Rect{Center: true, Percent: 100}),
			col.New(4),
		)

	default:
		// A foto precisa de mais altura que a assinatura para ser reconhecível,
		// mas o respiro encolhe na mesma medida: 5+30 fecha os mesmos 35mm do
		// bloco da assinatura (15+20). Assim a foto nunca empurra o rodapé com
		// o QR code para uma segunda página. Medido: acima de 35mm ele vaza.
		legenda = "Foto de confirmação do recebimento"
		m.AddRow(5, col.New(12))
		m.AddRow(30,
			col.New(3),
			image.NewFromBytesCol(6, imagem, formato, props.Rect{Center: true, Percent: 100}),
			col.New(3),
		)
	}

	m.AddRows(
		row.New(6).Add(
			text.NewCol(12, legenda, props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
		),
		row.New(5).Add(
			text.NewCol(12, nome+" - Matricula: "+matricula, props.Text{Size: 9, Align: align.Center}),
		),
	)
}

func truncarTexto(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Retorna o pedaço da string + reticências
	return s[:maxLen] + "..."
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
            text.NewCol(12, "COMPROVANTE DE ENTREGA DE EPI", props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 12}),
        ),

        // Dados do Funcionário
        row.New(6).Add(text.NewCol(12, "Empresa: "+Dadosfuncionarios.NomeEmpresa, props.Text{Size: 9, Style: fontstyle.Bold})),
        
        // --- LINHA 1: Nome do Funcionário e Matrícula lado a lado ---
        row.New(6).Add(
            text.NewCol(8, "Funcionario: "+Dadosfuncionarios.NomeFuncionario, props.Text{Size: 9}),
            text.NewCol(4, "Matricula: "+Dadosfuncionarios.Matricula, props.Text{Size: 9}),
        ),
        
        // --- LINHA 2: CPF abaixo do nome e Setor abaixo da Matrícula ---
        row.New(6).Add(
            text.NewCol(8, "CPF: "+Dadosfuncionarios.Cpf, props.Text{Size: 9}),
            text.NewCol(4, "Setor: "+Dadosfuncionarios.Setor, props.Text{Size: 9}),
        ),
        
        // --- LINHA 3: Cargo na sequência ---
        row.New(6).Add(
            text.NewCol(12, "Cargo: "+Dadosfuncionarios.Cargo, props.Text{Size: 9}),
        ),
        
        row.New(8).Add(text.NewCol(12, "Responsavel pela impressao: "+responsavel, props.Text{Size: 9})),
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
            text.NewCol(3, "Descricao", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
            text.NewCol(1, "Tamanho", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Align: align.Center}).WithStyle(estiloBorda),
        ),
    )

    // 3. Os 15 espaços em branco (Agora são caixas reais do grid!)
    for _, epi := range Dadosfuncionarios.Epi {

        descricaoCurta := truncarTexto(epi.Descricao, 25)
        dataFormatada := epi.Data.Time().Format("02/01/2006")
        quantidadaFormatada := strconv.Itoa(int(epi.Quantidade))
        m.AddRows(
            row.New(8).Add(
                text.NewCol(2, dataFormatada, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
                text.NewCol(2, quantidadaFormatada, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
                text.NewCol(2, epi.Ca, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
                text.NewCol(2, epi.NomeEpi, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
                text.NewCol(3, descricaoCurta, props.Text{Size: 8, Align: align.Center, Top: 2}).WithStyle(estiloBorda),
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
    m.AddRow(6, text.NewCol(10, "DECLARO QUE:", props.Text{Size: 9, Style: fontstyle.Bold}))

    // Dividimos o texto gigante em 3 variáveis limpas (sem "Enters" no meio do código)
    textoA := "a) Recebi nesta data, da EMPRESA acima identificada, minha empregadora, os equipamentos e materiais supra discriminados, os quais desde ja comprometo-me sempre a usar na execucao das minhas tarefas, zelando pela sua perfeita guarda, conservacao, uso e funcionamento como ora os estou recebendo."
    textoB := "b) Estou ciente e de pleno acordo que o descumprimento das condicoes estabelecidas na letra A supra, acarretara, alem da aplicacao de penas disciplinares, inclusive do meu contrato laboral, outras sancoes previstas em lei."
    textoC := "c) No caso da perda, dano, extravio ou avaria dos equipamentos e/ou materiais referidos na letra \"A\" favor comunicar imediatamente o departamento de Recursos Humanos."

    // Agora adicionamos uma Linha (Row) para CADA parágrafo.
    // Colocamos tamanhos diferentes (12, 10, 8) porque o texto A é maior e precisa de uma "caixa" mais alta para caber.
    m.AddRows(
        row.New(12).Add(text.NewCol(12, textoA, props.Text{Size: 8, Align: align.Left})),
        row.New(10).Add(text.NewCol(12, textoB, props.Text{Size: 8, Align: align.Left})),
        row.New(8).Add(text.NewCol(12, textoC, props.Text{Size: 8, Align: align.Left})),
    )


    // 4. O Bloco da Assinatura Digital
    // 1. Decodifica a string Base64 que veio do banco/struct
    // 1. Tratamento do Base64

    assinaturaBytes, formatoAssinatura, assinaturaValida := prepararImagemAssinatura(Dadosfuncionarios.Assinatura)

    // 2. Bloco de comprovação: foto, assinatura ou linha em branco
    adicionarComprovacaoRecebimento(m, assinaturaBytes, formatoAssinatura, assinaturaValida,
        Dadosfuncionarios.NomeFuncionario, Dadosfuncionarios.Matricula)

    m.AddRow(6, col.New(8))

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
            text.NewCol(12, "Autenticao Digital", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center}),
        ),
        // Subtítulo explicativo
        row.New(5).Add(
            text.NewCol(12, "Aponte a camera do celular para validar a assinatura e a integridade deste documento.", props.Text{Size: 8, Align: align.Center}),
        ),
    )

    m.AddRow(6, col.New(10))
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