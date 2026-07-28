# Modelos de Dados (Structs)

Este documento lista todas as structs de entrada/saída usadas na API, com seus campos e tags de validação.

## AtualizarPlanoParams

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Nome | *string | json:nome |
| Mensalidade | *decimal.Decimal | json:mensalidade |
| Descricao | *string | json:descricao |
| LimiteFuncionarios | *int32 | json:limite_funcionarios |
| LimiteUsuarios | *int32 | json:limite_usuarios |
| LimiteEpis | *int32 | json:limite_epis |
| Status | *string | json:status |
| ID | int32 | json:id |

---

## Departamento

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Departamento | string | json:departamento, binding:required,max=50 |

---

## DepartamentoDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| Departamento | string | json:departamento |

---

## DevolucaoDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int | json:id |
| Funcionario | Funcionario_Dto | json:funcionario |
| Epi | EpiDto | json:epi |
| Motivo | MotivoDevolucaoEpiDto | json:motivo |
| DataDevolucao | configs.DataBr | json:data_devolucao |
| QuantidadeADevolver | int | json:quantidade_a_devolver |
| AssinaturaDigital | string | json:assinatura_digital |
| EpiNovo | *EpiDto | json:epi_novo |
| Tamanho | *TamanhoDto | json:tamanho |
| TamanhoNovo | *TamanhoDto | json:tamanho_novo |
| NovaQuantidade | *int | json:quantidade_nova |

---

## DevolucaoInserir

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| IdFuncionario | int | json:idFuncionario, binding:required |
| IdEpi | int | json:idEpi, binding:required |
| IdMotivo | int | json:idMotivo, binding:required |
| IdTamanho | int | json:idTamanho, binding:required |
| DataDevolucao | configs.DataBr | json:data_devolucao, binding:required |
| QuantidadeADevolver | int | json:quantidadeADevolver, binding:required,numeric,gt=0 |
| Iduser | int | json:- |
| Troca | bool | json:houve_troca |
| IdEpiNovo | *int | json:idEpiNovo |
| IdTamanhoNovo | *int | json:idTamanhoNovo |
| NovaQuantidade | *int | json:quantidadeNova |
| AssinaturaDigital | string | json:assinatura_digital, binding:required |
| TokenValidacao | string | json:- |
| Observacao | *string | json:observacao |

---

## DevolucaoResponse

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int32 | json:id |
| DataDevolucao | configs.DataBr | json:data_devolucao |
| IDFuncionario | int32 | json:idFuncionario |
| FuncionarioNome | string | json:funcionarioNome |
| FuncionarioMatricula | int32 | json:funcionarioMatricula |
| EpiNome | string | json:epiNome |
| TamanhoNome | string | json:tamanhoNome |
| QuantidadeADevolver | int32 | json:quantidadeADevolver |
| MotivoNome | string | json:motivoNome |
| HouveTroca | bool | json:houveTroca |
| EpiNovoNome | *string | json:epiNovoNome |
| TamanhoNovoNome | *string | json:tamanhoNovoNome |
| QuantidadeNova | *int32 | json:quantidadeNova |
| Observacao | *string | json:observacao |
| AssinaturaDigital | *string | json:assinatura_digital |
| TokenValidacao | *string | json:token_validacao |

---

## EmpresaInserir

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| NomeFantasia | string | json:nome_fantasia, binding:required |
| Cnpj | string | json:cnpj, binding:cnpj,required,max=40 |
| Responsavel | string | json:responsavel, binding:required,max=40 |
| Email | string | json:email, binding:required,max=40 |
| Telefone | string | json:telefone, binding:max=40 |
| Plano | string | json:plano, binding:required |
| Status | string | json:status, binding:required |
| Mensalidade | decimal.Decimal | json:mensalidade, binding:required |
| Vencimento | configs.DataBr | json:vencimento, binding:required |
| Observacoes | string | json:observacoes, binding:lte=150 |
| Subdominio | string | json:- |

---

## EntradaDashbord

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int |  |
| IdEpi | int |  |
| IdTamanho | int |  |
| QuantidadeAtual | int |  |
| ValorUnitario | decimal.Decimal |  |
| Quantidade | int |  |
| DataEntrada | configs.DataBr |  |
| Lote | string |  |

---

## EntradaEpiDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| IDEpi | int | json:id_epi |
| IDTamanho | int | json:id_tamanho |
| IDFornecedor | int | json:id_fornecedor |
| DataEntrada | configs.DataBr | json:data_entrada |
| Quantidade | int | json:quantidade |
| QuantidadeAtual | int | json:quantidade_atual |
| ValorUnitario | decimal.Decimal | json:valor_unitario |
| Lote | string | json:lote |
| NotaFiscalNumero | string | json:nota_fiscal_numero |
| NotaFiscalSerie | string | json:nota_fiscal_serie |
| UsuarioCriacao | string | json:usuario |
| Epi | EpiSimples | json:epi |
| Tamanho | TamanhoSimples | json:tamanho |
| Fornecedor | FornecedorSimples | json:fornecedor |

---

## EntradaEpiInserir

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Idfornecedor | int32 | json:idfornecedor, binding:required,gt=0 |
| Nota_fiscal_numero | string | json:nota_fiscal_numero, binding:required,max=50 |
| Nota_fiscal_serie | string | json:nota_fiscal_serie, binding:required,max=20 |
| Data_emissao | configs.DataBr | json:data_emissao, binding:required |
| Id_user | int32 | json:- |
| Itens | []EntradaEpiItemInserir | json:itens, binding:required,dive |

---

## EntradaEpiItemInserir

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID_epi | int | json:id_epi, binding:required,numeric |
| Id_tamanho | int | json:id_tamanho, binding:required,numeric |
| Quantidade | int | json:quantidade, binding:required,numeric,gt=0 |
| DataFabricacao | configs.DataBr | json:data_fabricacao, binding:required |
| DataValidade | configs.DataBr | json:data_validade, binding:required |
| Lote | string | json:lote, binding:required,max=50 |
| ValorUnitario | decimal.Decimal | json:valor_unitario, binding:required |

---

## EntradaEstoqueDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int | json:id |
| Lote | string | json:lote |
| Quantidade | int | json:quantidade_inicial |
| QuantidadeAtual | int | json:quantidade_atual |
| ValorUnitario | decimal.Decimal | json:valor_unitario |
| DataValidade | configs.DataBr | json:data_validade |
| Tamanho | TamanhoDto | json:tamanho |
| Epi | EpiDtoEstoque | json:epi |
| DataEntrada | configs.DataBr | json:data_entrada |

---

## EntregaDashbord

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| IdFuncionario | int32 | json:id_funcionario |
| Data_entrega | configs.DataBr | json:data_entrega |
| Assinatura | string | json:assinatura |
| TokenValidacao | string | json:token_validacao |

---

## EntregaDoFuncionarioDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int64 | json:id |
| Data_entrega | configs.DataBr | json:data_entrega |
| Assinatura_Digital | string | json:assinatura_digital |
| Itens | []ItemEntregueDto | json:itens,omitempty |

---

## EntregaDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| Id_user | int32 | json:id_user |
| Funcionario | Funcionario_Dto | json:funcionario |
| Data_entrega | configs.DataBr | json:data_entrega |
| Assinatura_Digital | string | json:assinatura_digital |
| Token_validacao | string | json:token_validacao |
| Itens | []ItemEntregueDto | json:itens,omitempty |

---

## EntregaItensDashBord

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| IdEntregaCabecalho | int32 | json:id_entrega_cabecalho |
| IdEpi | int32 | json:id_epi |
| IdTamanho | int32 | json:id_tamanho |
| Quantidade | int32 | json:quantidade |

---

## EntregaParaInserir

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID_funcionario | int32 | json:id_funcionario, binding:required |
| Id_user | int32 | json:id_user |
| Data_entrega | configs.DataBr | json:data_entrega, binding:required |
| IdTroca | *int32 | json:id_troca |
| Assinatura_Digital | string | json:assinatura_digital, binding:required |
| Itens | []ItemParaInserir | json:itens, binding:required,min=1,dive |

---

## EpiDashBord

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| Nome | string | json:nome |
| AlertaMinimo | int32 | json:alerta_minimo |

---

## EpiDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| Nome | string | json:nome |
| Fabricante | string | json:fabricante |
| CA | string | json:ca |
| Tamanhos | []TamanhoDto | json:tamanhos |
| Descricao | string | json:descricao |
| DataValidadeCa | configs.DataBr | json:validade_ca |
| Protecao | TipoProtecaoDto | json:protecao |
| AlertaMinimo | int32 | json:alerta_minimo |

---

## EpiDtoDevolucao

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| Nome | string | json:nome |
| Fabricante | string | json:fabricante |
| CA | string | json:ca |
| Tamanhos | []TamanhoDto | json:tamanhos |
| Descricao | string | json:descricao |
| DataValidadeCa | configs.DataBr | json:validade_ca |
| Protecao | TipoProtecaoDto | json:protecao |
| AlertaMinimo | int32 | json:alerta_minimo |
| SaldoAtual | int32 | json:saldoAtual |

---

## EpiDtoEntrega

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| Nome | string | json:nome |
| Fabricante | string | json:fabricante |
| CA | string | json:ca |
| Tamanhos | []TamanhoEntregaDto | json:tamanhos |
| Descricao | string | json:descricao |
| DataValidadeCa | configs.DataBr | json:validade_ca |
| Protecao | TipoProtecaoDto | json:protecao |
| AlertaMinimo | int32 | json:alerta_minimo |

---

## EpiDtoEstoque

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| Nome | string | json:nome |
| Fabricante | string | json:fabricante |
| Descricao | string | json:descricao |
| DataValidadeCa | configs.DataBr | json:validadeCa |
| Ca | string | json:ca |
| AlertaMinimo | int | json:alertaMinimo |
| Protecao | TipoProtecaoDto | json:protecao |

---

## EpiInserir

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Nome | string | json:nome, binding:required |
| Fabricante | string | json:fabricante, binding:required,max=100 |
| CA | string | json:ca, binding:required,max=20 |
| Descricao | string | json:descricao, binding:lte=250 |
| DataValidadeCa | configs.DataBr | json:data_validade_ca, binding:required |
| IdTamanho | []int32 | json:id_tamanho, binding:required,min=1 |
| IDProtecao | int32 | json:id_protecao, binding:required,numeric |
| AlertaMinimo | int32 | json:alerta_minimo, binding:required,gte=0 |

---

## EpiResponse

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| Nome | string | json:nome |
| Fabricante | string | json:fabricante |
| CA | string | json:ca |
| Descricao | string | json:descricao |
| DataValidadeCa | configs.DataBr | json:validade_ca |
| Protecao | TipoProtecaoDto | json:protecao |
| AlertaMinimo | int | json:alerta_minimo |

---

## EpiSimples

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| Nome | string | json:nome |
| Fabricante | string | json:fabricante |
| CA | string | json:ca |

---

## EstoquePorTamanhoDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| IDEpi | int32 | json:id_epi |
| IDTamanho | int32 | json:id_tamanho |
| TamanhoNome | string | json:tamanho_nome |
| QuantidadeAtual | int32 | json:quantidade_atual |

---

## EstoqueSaldoTotalDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| IDEpi | int32 | json:id_epi |
| NomeEpi | string | json:nome_epi |
| QuantidadeAtual | int32 | json:quantidade_atual |
| SaldoTotal | decimal.Decimal | json:saldo_total |

---

## EstoqueTotalDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| IDEpi | int32 | json:id_epi |
| NomeEpi | string | json:nome_epi |
| QuantidadeTotal | int64 | json:quantidade_total |

---

## Fornecedor

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| TenantID | int | json:- |
| RazaoSocial | string | json:razao_social |
| NomeFantasia | string | json:nome_fantasia |
| CNPJ | string | json:cnpj |
| InscricaoEstadual | string | json:inscricao_estadual |
| Ativo | bool | json:ativo |

---

## FornecedorDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| RazaoSocial | string | json:razao_social |
| NomeFantasia | string | json:nome_fantasia |
| CNPJ | string | json:cnpj |
| InscricaoEstadual | string | json:inscricao_estadual |

---

## FornecedorInserir

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| RazaoSocial | string | json:razao_social, binding:required,max=100 |
| NomeFantasia | string | json:nome_fantasia, binding:required,max=100 |
| CNPJ | string | json:cnpj, binding:required,cnpj |
| InscricaoEstadual | string | json:inscricao_estadual, binding:required |

---

## FornecedorSimples

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| NomeFantasia | string | json:nome_fantasia |
| RazaoSocial | string | json:razao_social |

---

## FornecedorUpdate

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| RazaoSocial | *string | json:razao_social |
| NomeFantasia | *string | json:nome_fantasia |
| CNPJ | *string | json:cnpj, binding:cnpj |
| InscricaoEstadual | *string | json:inscricao_estadual |

---

## Funcao

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Funcao | string`json:"funcao" | binding:required,max=50 |
| IdDepartamento | int | json:id_departamento, binding:required,min=1 |

---

## FuncaoDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| Funcao | string | json:cargo |
| Departamento | DepartamentoDto | json:departamento |

---

## FuncionarioCompletoDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int32 | json:id |
| Nome | string | json:nome |
| Matricula | string | json:matricula |
| Funcao | FuncaoDto | json:funcao |
| Entregas | []EntregaDoFuncionarioDto | json:entregas |

---

## FuncionarioDashbord

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| Nome | string | json:nome |
| Matricula | string | json:matricula |

---

## FuncionarioInserir

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Nome | string | json:nome, binding:required,min=3,max=150 |
| ID_departamento | int32 | json:id_departamento, binding:required,min=1 |
| ID_funcao | int32 | json:id_funcao, binding:required,min=1 |

---

## Funcionario_Dto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int32 | json:id |
| Nome | string | json:nome |
| Matricula | string | json:matricula |
| Funcao | FuncaoDto | json:funcao |

---

## ItemEntregueDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int32 | json:id |
| Quantidade | int32 | json:quantidade |
| Epi | EpiResponse | json:epi |
| Tamanho | TamanhoDto | json:tamanho |

---

## ItemParaInserir

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID_epi | int32 | json:id_epi, binding:required |
| ID_tamanho | int32 | json:id_tamanho, binding:required |
| ID_entrada_item | int32 | json:id_entrada_item |
| Quantidade | int32 | json:quantidade, binding:required,gt=0 |

---

## LoginInput

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Email | string | json:email, binding:required,email |
| Senha | string | json:senha, binding:required |

---

## LoginResponse

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Token | string | json:token |
| User | Usuario | json:user |

---

## MotivoDevolucao

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Motivo | string | json:motivo, binding:required |
| Descaste | bool | json:gera_descarte |

---

## MotivoDevolucaoEpiDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int | json:id |
| Motivo | string | json:motivo |

---

## PaginacaoParams

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Pagina | int32 | binding:min=1 |
| Limite | int32 | binding:min=1,max=100 |

---

## Plano

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| Nome | string | json:nome, binding:required |
| Mensalidade | decimal.Decimal | json:mensalidade, binding:required |
| LimiteFuncionarios | *int32 | json:limite_funcionarios |
| LimiteUsuarios | *int32 | json:limite_usuarios |
| LimiteEpis | *int32 | json:limite_epis |
| Status | string | json:status, binding:required |
| Descricao | string | json:descricao, binding:required |

---

## PlanoNome

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int |  |
| Nome | string |  |

---

## RecuperaLogin

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Empresa | string | json:empresa, binding:required |
| TenantId | int | json:- |
| Email | string | json:email, binding:required |

---

## RecuperaUser

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int | json:id |
| Nome | string | json:nome |
| Email | string | json:email |
| Role | string | json:role |

---

## RecuperaUserEntrada

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Id | int | json:id |
| Nome | string | json:nome |

---

## RedefinirSenha

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Token | string | json:token, binding:required |
| NovaSenha | string | json:senha_nova, binding:required,min=6 |
| TenantId | int | json:- |

---

## TamanhoDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| Tamanho | string | json:tamanho |

---

## TamanhoEntregaDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| Tamanho | string | json:tamanho |
| Id_epi | int32 | json:id_epi |
| QuantidadeAtual | int32 | json:quantidade_atual |

---

## TamanhoSimples

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| Tamanho | string | json:tamanho |

---

## Tamanhos

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Tamanho | string | json:tamanho, binding:required |

---

## TipoProtecao

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Nome | string | json:nome, binding:required,max=50 |

---

## TipoProtecaoDto

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int64 | json:id |
| Nome | string | json:nome |

---

## UpdateEpiInput

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Nome | *string | json:nome |
| Fabricante | *string | json:fabricante |
| CA | *string | json:ca |
| Descricao | *string | json:descricao |
| DataValidadeCa | *configs.DataBr | json:validade_ca |
| IdProtecao | *int32 | json:id_protecao |
| AlertaMinimo | *int32 | json:alerta_minimo |
| Tamanhos | []int32 | json:tamanhos |

---

## UpdateFuncionarioRequest

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Nome | *string | json:nome |
| IdDepartamento | *int32 | json:id_departamento |
| IdFuncao | *int32 | json:id_funcao |

---

## Usuario

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| Nome | string | json:nome, binding:required,min=3,max=50 |
| Email | string | json:email, binding:required,email |
| Senha | string | json:senha, binding:required,max=10 |
| Role | string | json:cargo, binding:required |

---

## UsuarioResponse

| Campo | Tipo | Tags JSON/Binding |
|-------|------|-------------------|
| ID | int | json:id |
| Nome | string | json:nome |
| Email | string | json:email |
| Cargo | string | json:cargo |

---
