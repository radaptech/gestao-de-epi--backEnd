package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func criarTabelasPostgres(t *testing.T, pool *pgxpool.Pool) {

	schema := `
-- ==========================================
-- NÍVEL 0: Tabelas sem dependências (Base)
-- ==========================================

CREATE TABLE public.schema_migrations (
  version bigint NOT NULL PRIMARY KEY,
  dirty boolean NOT NULL
);

CREATE TABLE public.planos (
  id SERIAL PRIMARY KEY,
  nome character varying NOT NULL,
  mensalidade numeric NOT NULL,
  limite_funcionarios integer,
  limite_usuarios integer,
  limite_epis integer,
  status character varying DEFAULT 'Ativo'::character varying,
  descricao text NOT NULL,
  criado_em timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.fornecedores (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  razao_social character varying NOT NULL,
  nome_fantasia character varying NOT NULL,
  cnpj character varying NOT NULL,
  inscricao_estadual character varying NOT NULL,
  ativo boolean DEFAULT true,
  cancelado_em timestamp without time zone
);

-- ==========================================
-- NÍVEL 1: Depende apenas de Nível 0
-- ==========================================

CREATE TABLE public.empresas (
  id SERIAL PRIMARY KEY,
  nome_fantasia character varying NOT NULL,
  razao_social character varying NOT NULL,
  cnpj character varying NOT NULL UNIQUE,
  subdominio character varying NOT NULL UNIQUE,
  criado_em timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
  plano_id integer,
  status character varying NOT NULL DEFAULT 'Em teste'::character varying,
  vencimento date,
  observacoes text,
  responsavel character varying,
  email character varying,
  telefone character varying,
  CONSTRAINT empresas_plano_id_fkey FOREIGN KEY (plano_id) REFERENCES public.planos(id)
);

-- ==========================================
-- NÍVEL 2: Depende de Empresas
-- ==========================================

CREATE TABLE public.departamento (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  nome character varying NOT NULL,
  ativo boolean NOT NULL DEFAULT true,
  deletado_em timestamp without time zone,
  CONSTRAINT departamento_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id)
);

CREATE TABLE public.tipo_protecao (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  nome character varying NOT NULL,
  ativo boolean NOT NULL DEFAULT true,
  deletado_em timestamp without time zone,
  CONSTRAINT tipo_protecao_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id)
);

CREATE TABLE public.tamanho (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  tamanho character varying NOT NULL,
  ativo boolean NOT NULL DEFAULT true,
  deletado_em timestamp without time zone,
  CONSTRAINT tamanho_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id)
);

CREATE TABLE public.motivo_devolucao (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  motivo character varying NOT NULL,
  ativo boolean NOT NULL DEFAULT true,
  deletado_em timestamp without time zone,
  gera_descarte boolean NOT NULL DEFAULT false,
  CONSTRAINT motivo_devolucao_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id)
);

CREATE TABLE public.usuarios (
  id SERIAL PRIMARY KEY,
  tenant_id integer,
  nome character varying NOT NULL,
  email character varying NOT NULL,
  senha_hash text NOT NULL,
  ativo boolean DEFAULT true,
  role character varying,
  token_recuperacao_senha text,
  token_expiracao timestamp without time zone,
  ultimo_acesso timestamp without time zone,
  CONSTRAINT usuarios_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id)
);

-- ==========================================
-- NÍVEL 3: Depende do Nível 2
-- ==========================================

CREATE TABLE public.funcao (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  nome character varying NOT NULL,
  iddepartamento integer NOT NULL,
  ativo boolean NOT NULL DEFAULT true,
  deletado_em timestamp without time zone,
  CONSTRAINT funcao_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id),
  CONSTRAINT funcao_iddepartamento_fkey FOREIGN KEY (iddepartamento) REFERENCES public.departamento(id)
);

CREATE TABLE public.epi (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  nome character varying NOT NULL,
  fabricante character varying NOT NULL,
  ca character varying NOT NULL,
  descricao text NOT NULL,
  validade_ca date NOT NULL,
  idtipoprotecao integer NOT NULL,
  alerta_minimo integer NOT NULL,
  ativo boolean NOT NULL DEFAULT true,
  deletado_em timestamp without time zone,
  CONSTRAINT epi_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id),
  CONSTRAINT epi_idtipoprotecao_fkey FOREIGN KEY (idtipoprotecao) REFERENCES public.tipo_protecao(id)
);

CREATE TABLE public.entrada_nf (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  nota_fiscal_numero character varying NOT NULL,
  nota_fiscal_serie character varying DEFAULT '1'::character varying,
  data_emissao date NOT NULL,
  data_registro timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
  ativo boolean NOT NULL DEFAULT true,
  id_usuario_criacao integer NOT NULL,
  id_usuario_cancelamento integer,
  cancelada_em timestamp without time zone,
  idfornecedor integer NOT NULL,
  CONSTRAINT entrada_nf_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id),
  CONSTRAINT fk_usuario_criacao_nf FOREIGN KEY (id_usuario_criacao) REFERENCES public.usuarios(id),
  CONSTRAINT fk_usuario_cancelamento_nf FOREIGN KEY (id_usuario_cancelamento) REFERENCES public.usuarios(id),
  CONSTRAINT fk_entrada_nf_fornecedor FOREIGN KEY (idfornecedor) REFERENCES public.fornecedores(id)
);

-- ==========================================
-- NÍVEL 4: Depende do Nível 3
-- ==========================================

CREATE TABLE public.funcionario (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  nome character varying NOT NULL,
  matricula integer NOT NULL,
  idfuncao integer NOT NULL,
  iddepartamento integer NOT NULL,
  cpf varchar(15) not null,
  ativo boolean NOT NULL DEFAULT true,
  deletado_em timestamp without time zone,
  CONSTRAINT funcionario_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id),
  CONSTRAINT funcionario_idfuncao_fkey FOREIGN KEY (idfuncao) REFERENCES public.funcao(id),
  CONSTRAINT funcionario_iddepartamento_fkey FOREIGN KEY (iddepartamento) REFERENCES public.departamento(id)
);

CREATE TABLE public.tamanhos_epis (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  idepi integer NOT NULL,
  idtamanho integer NOT NULL,
  ativo boolean NOT NULL DEFAULT true,
  deletado_em timestamp without time zone,
  CONSTRAINT tamanhos_epis_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id),
  CONSTRAINT tamanhos_epis_idepi_fkey FOREIGN KEY (idepi) REFERENCES public.epi(id),
  CONSTRAINT tamanhos_epis_idtamanho_fkey FOREIGN KEY (idtamanho) REFERENCES public.tamanho(id)
);

CREATE TABLE public.entrada_epi_item (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  entrada_nf_id integer NOT NULL,
  id_epi integer NOT NULL,
  id_tamanho integer NOT NULL,
  quantidade integer NOT NULL,
  quantidade_atual integer NOT NULL,
  data_fabricacao date NOT NULL,
  data_validade date NOT NULL,
  lote character varying NOT NULL,
  valor_unitario numeric NOT NULL,
  cancelada_em timestamp without time zone,
  ativo boolean NOT NULL DEFAULT true,
  id_usuario_criacao integer NOT NULL,
  id_usuario_cancelamento integer,
  CONSTRAINT entrada_epi_item_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id),
  CONSTRAINT entrada_epi_item_entrada_nf_id_fkey FOREIGN KEY (entrada_nf_id) REFERENCES public.entrada_nf(id),
  CONSTRAINT entrada_epi_item_id_epi_fkey FOREIGN KEY (id_epi) REFERENCES public.epi(id),
  CONSTRAINT entrada_epi_item_id_tamanho_fkey FOREIGN KEY (id_tamanho) REFERENCES public.tamanho(id),
  CONSTRAINT fk_usuario_criacao_item FOREIGN KEY (id_usuario_criacao) REFERENCES public.usuarios(id),
  CONSTRAINT fk_usuario_cancelamento_item FOREIGN KEY (id_usuario_cancelamento) REFERENCES public.usuarios(id)
);

-- ==========================================
-- NÍVEL 5: Transações Operacionais
-- ==========================================

CREATE TABLE public.entrega_epi (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  idfuncionario integer NOT NULL,
  data_entrega date NOT NULL,
  assinatura text NOT NULL,
  idtroca integer,
  cancelada_em timestamp without time zone,
  ativo boolean NOT NULL DEFAULT true,
  token_validacao text,
  id_usuario_entrega integer,
  id_usuario_entrega_cancelamento integer,
  CONSTRAINT entrega_epi_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id),
  CONSTRAINT entrega_epi_idfuncionario_fkey FOREIGN KEY (idfuncionario) REFERENCES public.funcionario(id),
  CONSTRAINT entrega_epi_id_usuario_entrega_fkey FOREIGN KEY (id_usuario_entrega) REFERENCES public.usuarios(id),
  CONSTRAINT entrega_epi_id_usuario_entrega_cancelamento_fkey FOREIGN KEY (id_usuario_entrega_cancelamento) REFERENCES public.usuarios(id)
);

CREATE TABLE public.devolucao (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  idepi integer NOT NULL,
  idfuncionario integer NOT NULL,
  idmotivo integer NOT NULL,
  data_devolucao date NOT NULL,
  idtamanho integer NOT NULL,
  quantidadeadevolver integer NOT NULL,
  idepinovo integer,
  idtamanhonovo integer,
  quantidadenova integer,
  cancelada_em timestamp without time zone,
  assinatura_digital text NOT NULL,
  ativo boolean NOT NULL DEFAULT true,
  id_usuario_cancelamento integer,
  id_usuario_devolucao_cancelamento integer,
  token_validacao text,
  houve_troca boolean DEFAULT false,
  observacao text,
  CONSTRAINT devolucao_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id),
  CONSTRAINT devolucao_idepi_fkey FOREIGN KEY (idepi) REFERENCES public.epi(id),
  CONSTRAINT devolucao_idfuncionario_fkey FOREIGN KEY (idfuncionario) REFERENCES public.funcionario(id),
  CONSTRAINT devolucao_idmotivo_fkey FOREIGN KEY (idmotivo) REFERENCES public.motivo_devolucao(id),
  CONSTRAINT devolucao_idepinovo_fkey FOREIGN KEY (idepinovo) REFERENCES public.epi(id),
  CONSTRAINT devolucao_idtamanhonovo_fkey FOREIGN KEY (idtamanhonovo) REFERENCES public.tamanho(id),
  CONSTRAINT devolucao_idtamanho_fkey FOREIGN KEY (idtamanho) REFERENCES public.tamanho(id),
  CONSTRAINT devolucao_id_usuario_cancelamento_fkey FOREIGN KEY (id_usuario_cancelamento) REFERENCES public.usuarios(id),
  CONSTRAINT devolucao_id_usuario_devolucao_cancelamento_fkey FOREIGN KEY (id_usuario_devolucao_cancelamento) REFERENCES public.usuarios(id)
);

-- ==========================================
-- NÍVEL 6: Depende das Transações
-- ==========================================

CREATE TABLE public.epis_entregues (
  id SERIAL PRIMARY KEY,
  tenant_id integer NOT NULL,
  id_entrega_cabecalho integer NOT NULL,
  id_entrada_item integer NOT NULL,
  id_epi integer NOT NULL,
  id_tamanho integer NOT NULL,
  quantidade integer NOT NULL,
  ativo boolean NOT NULL DEFAULT true,
  deletado_em timestamp without time zone,
  CONSTRAINT epis_entregues_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.empresas(id),
  CONSTRAINT epis_entregues_id_epi_fkey FOREIGN KEY (id_epi) REFERENCES public.epi(id),
  CONSTRAINT epis_entregues_id_entrada_item_fkey FOREIGN KEY (id_entrada_item) REFERENCES public.entrada_epi_item(id),
  CONSTRAINT fk_epis_entregues_cabecalho FOREIGN KEY (id_entrega_cabecalho) REFERENCES public.entrega_epi(id)
);
`

	_, err := pool.Exec(context.Background(), schema)
	if err != nil {
		t.Fatalf("Erro ao criar tabelas no banco de testes: %v", err)
	}
}
