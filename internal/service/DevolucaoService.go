package service

import (
	"context"
	"fmt"
	"log"

	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DevolucaoRepository define o contrato para persistência de devoluções
type DevolucaoRepository interface {
	AdicionarTroca(ctx context.Context, qtx *repository.Queries, arg repository.AddTrocaEpiParams) (int32, error)
	Cancelar(ctx context.Context, qtx *repository.Queries, arg repository.CancelarDevolucaoParams) (int32, error)
	 Listar(ctx context.Context, tenantId int32) ([]repository.ListarDevolucoesRow, error)
}

type DevolucaoService struct {
	repo            DevolucaoRepository
	db              *pgxpool.Pool
	queries         *repository.Queries
	repoEntrega     EntregaService // Service de entrega para automatizar a troca
	MotivoDevolucao MotivoDevolucaoRepository
}

func NewDevolucaoService(d DevolucaoRepository, db *pgxpool.Pool, repoEntregaEpi EntregaService, motivoDev MotivoDevolucaoRepository) *DevolucaoService {
	return &DevolucaoService{
		repo:            d,
		db:              db,
		queries:         repository.New(db),
		repoEntrega:     repoEntregaEpi,
		MotivoDevolucao: motivoDev,
	}
}

func (d *DevolucaoService) TokenDevolucao(ctx context.Context, tenantId, Idfuncionario int32) (string, error) {

	fmt.Printf("🔍 [TOKEN] Buscando Funcionario %d | Tenant %d\n", Idfuncionario, tenantId)
	funcionario, err := d.queries.BuscaFuncionarioPorId(ctx, repository.BuscaFuncionarioPorIdParams{
		ID:       int32(Idfuncionario),
		TenantID: tenantId,
	})
	if err != nil {
		fmt.Printf("❌ [TOKEN] Erro na query: %v\n", err)
		if err == pgx.ErrNoRows {
			return "", helper.ErrNaoEncontrado
		}
		return "", err
	}

	token := helper.GerarTokenDevolucao(funcionario.Nome, funcionario.FuncaoNome,
		funcionario.DepartamentoNome, time.Now())

	return token, nil
}

// SalvarDevolucao orquestra a devolução de um EPI e opcionalmente realiza uma nova entrega (Troca)
func (d *DevolucaoService) SalvarDevolucao(ctx context.Context, modelDevolucao model.DevolucaoInserir, tenantId int32, token string) error {
	// Inicia a transação atômica
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := d.queries.WithTx(tx)

	//valida o saldo

	saldoAtual, err := d.MotivoDevolucao.ConsultaSaldo(ctx, repository.ConsultarSaldoEpiFuncionarioParams{
		Idfuncionario: int32(modelDevolucao.IdFuncionario),
		IDEpi:         int32(modelDevolucao.IdEpi),
		IDTamanho:     int32(modelDevolucao.IdTamanho),
		TenantID:      tenantId,
	})

	if err != nil {

		return fmt.Errorf("erro ao consultar saldo do funcionario: %w", err)
	}

	if int32(modelDevolucao.QuantidadeADevolver) > saldoAtual {

		log.Printf("erro: %v", err)
		return fmt.Errorf("operação bloqueada: o funcionário possui apenas %d unidade(s) deste EPI em mãos, mas tentou devolver %d", saldoAtual, modelDevolucao.QuantidadeADevolver)
	}
	// Prepara variáveis para caso de troca (EPI novo)
	var idEpiNovo, idTamanhoNovo, idQuantidadeNova pgtype.Int4
	if modelDevolucao.Troca {
		if modelDevolucao.IdEpiNovo == nil {
			return fmt.Errorf("id do novo epi é obrigatório para trocas")
		}
		idEpiNovo = pgtype.Int4{Int32: int32(*modelDevolucao.IdEpiNovo), Valid: true}
		idTamanhoNovo = pgtype.Int4{Int32: int32(*modelDevolucao.IdTamanhoNovo), Valid: true}
		idQuantidadeNova = pgtype.Int4{Int32: int32(*modelDevolucao.NovaQuantidade), Valid: true}
	}

	// Regra de Descarte: Motivos 1 (Desgaste), 2 (Dano) ou 3 (Vencimento) não voltam ao estoque
	ehDescarte, err := d.MotivoDevolucao.Descarte(ctx, repository.EhDescarteParams{
		ID:       int32(modelDevolucao.IdMotivo),
		TenantID: tenantId,
	})
	if err != nil {

		return err
	}

	// Se não for descarte, devolve a quantidade ao estoque no lote mais recente
	if !ehDescarte {
		qtdRestanteParaDevolver := int32(modelDevolucao.QuantidadeADevolver)

		lotes, err := qtx.ListarLotesParaRepor(ctx, repository.ListarLotesParaReporParams{
			TenantID:  tenantId,
			IDEpi:     int32(modelDevolucao.IdEpi),
			IDTamanho: int32(modelDevolucao.IdTamanho),
		})
		if err != nil {
			log.Printf("erro ao buscar lotes para repor: %v", err)
			return fmt.Errorf("erro ao buscar lotes para reposição")
		}

		for _, lote := range lotes {

			if qtdRestanteParaDevolver <= 0 {
				break
			}

			espaçoNoLote := lote.Quantidade - lote.QuantidadeAtual
			qtdParaDeixarNoLote := min(qtdRestanteParaDevolver, espaçoNoLote)

			novoSaldoLote := lote.QuantidadeAtual + qtdParaDeixarNoLote

			err := qtx.AtualizarSaldoLote(ctx, repository.AtualizarSaldoLoteParams{
				ID:              lote.ID,
				QuantidadeAtual: novoSaldoLote,
				TenantID:        tenantId,
			})
			if err != nil {

				return fmt.Errorf("erro ao atualizar saldo do lote %d: %w", lote.ID, err)
			}

			qtdRestanteParaDevolver -= qtdParaDeixarNoLote
		}

		if qtdRestanteParaDevolver > 0 {
			return fmt.Errorf("anomalia de estoque: impossível repor %d itens. A quantidade devolvida excede o total comprado na história deste EPI", qtdRestanteParaDevolver)
		}

	}

	// Registra a devolução/troca na tabela principal
	arg := repository.AddTrocaEpiParams{
		TenantID:              tenantId,
		Idfuncionario:         int32(modelDevolucao.IdFuncionario),
		Idepi:                 int32(modelDevolucao.IdEpi),
		Idmotivo:              int32(modelDevolucao.IdMotivo),
		DataDevolucao:         pgtype.Date{Time: modelDevolucao.DataDevolucao.Time(), Valid: true},
		Idtamanho:             int32(modelDevolucao.IdTamanho),
		Quantidadeadevolver:   int32(modelDevolucao.QuantidadeADevolver),
		Idepinovo:             idEpiNovo,
		Idtamanhonovo:         idTamanhoNovo,
		Quantidadenova:        idQuantidadeNova,
		AssinaturaDigital:     modelDevolucao.AssinaturaDigital,
		IDUsuarioCancelamento: pgtype.Int4{Int32: int32(modelDevolucao.Iduser), Valid: true},
		TokenValidacao:        pgtype.Text{String: token, Valid: true},
		HouveTroca: pgtype.Bool{Bool: modelDevolucao.Troca, Valid: true},
		Observacao: pgtype.Text{String: *modelDevolucao.Observacao, Valid: *modelDevolucao.Observacao == ""},
	}

	idDevolucao, err := d.repo.AdicionarTroca(ctx, qtx, arg)
	if err != nil {
		return fmt.Errorf("erro ao registrar devolução: %w", err)
	}

	// Se for uma troca, dispara automaticamente o processo de nova entrega
	if modelDevolucao.Troca {
		idTrocaInt := int32(idDevolucao)
		modelEntrega := model.EntregaParaInserir{
			ID_funcionario:     int32(arg.Idfuncionario),
			Id_user:            int32(modelDevolucao.Iduser),
			Data_entrega:       modelDevolucao.DataDevolucao,
			IdTroca:            &idTrocaInt,
			Assinatura_Digital: arg.AssinaturaDigital,
			Itens: []model.ItemParaInserir{
				{
					ID_epi:     int32(*modelDevolucao.IdEpiNovo),
					ID_tamanho: int32(*modelDevolucao.IdTamanhoNovo),
					Quantidade: int32(*modelDevolucao.NovaQuantidade),
				},
			},
		}
		// Chama o service de entrega dentro da mesma transação (qtx)
		if err := d.repoEntrega.RegistrarEntrega(ctx, qtx, modelEntrega, tenantId, token); err != nil {
			return fmt.Errorf("erro ao realizar nova entrega da troca: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// CancelarDevolucao reverte uma devolução, cancelando entregas vinculadas e repondo estoque
func (d *DevolucaoService) CancelarDevolucao(ctx context.Context, id, iduser, tenantId int) error {
	if id <= 0 {
		return helper.ErrId
	}

	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := d.queries.WithTx(tx)

	// Cancela o registro de devolução e obtém o ID para rastreio
	arg := repository.CancelarDevolucaoParams{
		ID:                             int32(id),
		IDUsuarioDevolucaoCancelamento: pgtype.Int4{Int32: int32(iduser), Valid: true},
		TenantID:                       int32(tenantId),
	}

	idDevolucao, err := d.repo.Cancelar(ctx, qtx, arg)
	if err != nil {
		return fmt.Errorf("falha ao cancelar devolução: %w", err)
	}

	// Tenta cancelar a entrega gerada por essa troca (se houver uma)
	idEntrega, err := qtx.CancelaEntregaPorIdTroca(ctx, repository.CancelaEntregaPorIdTrocaParams{
		Idtroca:                      pgtype.Int4{Int32: int32(idDevolucao), Valid: true},
		IDUsuarioEntregaCancelamento: arg.IDUsuarioDevolucaoCancelamento,
		TenantID:                     arg.TenantID,
	})

	// Se existia uma entrega vinculada, cancelamos os itens e devolvemos ao estoque
	if err == nil {
		itensCancelados, err := qtx.CancelaItemEntregue(ctx, repository.CancelaItemEntregueParams{
			IDEntregaCabecalho: idEntrega,
			TenantID:           arg.TenantID,
		})
		if err != nil {
			return err
		}

		for _, item := range itensCancelados {
			// Devolve a quantidade ao lote original de entrada
			linhas, err := qtx.ReporEstoqueLote(ctx, repository.ReporEstoqueLoteParams{
				QuantidadeAtual: item.Quantidade,
				ID:              item.IDEntradaItem,
				TenantID:        arg.TenantID,
			})
			if err != nil || linhas == 0 {
				return fmt.Errorf("erro ao repor estoque do lote %d", item.IDEntradaItem)
			}
		}
	} else if err != pgx.ErrNoRows {
		// Se o erro não for "sem linhas", é um erro real de banco
		return fmt.Errorf("erro crítico ao buscar entrega vinculada: %w", err)
	}

	return tx.Commit(ctx)
}



// ListarDevolucoes gerencia a busca paginada com filtros
func (d *DevolucaoService) ListarDevolucoes(ctx context.Context,tenantId int32) ([]model.DevolucaoResponse, error) {


	devolucoes, err := d.repo.Listar(ctx, tenantId)
	if err != nil {
		return []model.DevolucaoResponse{}, helper.TraduzErroPostgres(err)
	}

	dto := make([]model.DevolucaoResponse, 0, len(devolucoes))

	for _, dev := range devolucoes {
	
		obs:= dev.Observacao.String
		assinaturta:= dev.AssinaturaDigital
		token:=dev.TokenValidacao.String
		devo:= model.DevolucaoResponse{
			ID: dev.ID,
			DataDevolucao: *configs.NewDataBrPtr(dev.DataDevolucao.Time),
			IDFuncionario: dev.Idfuncionario,
			FuncionarioNome: dev.Funcionarionome,
			FuncionarioMatricula: dev.Matricula,
			EpiNome: dev.Epinome,
			TamanhoNome: dev.Tamanho,
			QuantidadeADevolver: dev.Quantidadeadevolver,
			MotivoNome: dev.Motivo,
			HouveTroca: dev.HouveTroca.Bool,
			Observacao: &obs,
			AssinaturaDigital: &assinaturta,
			TokenValidacao: &token,
		}

		if devo.HouveTroca == true{

			epinovo:= dev.Epinovo.String
			tamanhoNovo:= dev.Tamanhonovo.String
			quantidadeNova:=dev.Quantidadenova.Int32
			devo.EpiNovoNome = &epinovo
			devo.TamanhoNovoNome = &tamanhoNovo
			devo.QuantidadeNova = &quantidadeNova
		}

		dto = append(dto, devo)
	}
		
	return dto,nil

}
