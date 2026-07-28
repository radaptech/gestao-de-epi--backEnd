## Documentação: Fluxo de Cancelamento de Devolução

### Descrição Geral
Este diagrama ilustra o fluxo transacional do processo de cancelamento de uma devolução no sistema de gestão de EPI. O fluxo inclui validações, tratamento de erros e operações de reversão de estoque.

### Etapas do Fluxo

1. **Início da Transação (A)**
    - Inicia uma transação de banco de dados para garantir consistência

2. **Cancelamento de Devolução (B)**
    - Executa a operação de cancelamento da devolução

3. **Busca da Entrega (C)** - Ponto de Decisão
    - Consulta a entrega usando o identificador `IdTroca`
    - Três caminhos possíveis:

4. **Tratamento de Erros (C)**
    - **pgx.ErrNoRows**: Nenhuma entrega encontrada → Commit da transação (sem alterações)
    - **Outro Erro**: Falha na consulta → Rollback e retorno do erro

5. **Cancelamento de Itens e Reposição (D)**
    - Se a entrega for encontrada:
      - Cancela os itens da devolução
      - Repõe o estoque correspondente
    - Dois caminhos possíveis:

6. **Finalização (F, G, H)**
    - **Sucesso (D)**: Commit da transação → Fim com sucesso
    - **Erro (D)**: Rollback da transação → Retorno do erro

### Palavras-chave
- Transação atômica
- Tratamento de exceções
- Rollback e Commit
- Gestão de estoque
#  Fluxo de Devolução



## Fluxo de Devolução de EPI

Este diagrama representa o processo completo de devolução de Equipamentos de Proteção Individual (EPI), incluindo validações, gerenciamento de estoque e tratamento de transações.

### Etapas Principais

1. **Inicialização da Transação**
    - Inicia uma transação para garantir consistência dos dados
    - Cria um contexto de transação (qtx)

2. **Validação de Troca**
    - Verifica se a devolução é uma troca de EPI
    - Se for troca: valida o novo EPI antes de prosseguir
    - Rollback automático em caso de validação inválida

3. **Lógica de Estoque**
    - Determina o motivo da devolução (descarte ou devolução normal)
    - **Descarte**: item é removido do estoque
    - **Devolução Normal**: item é reintegrado ao estoque

4. **Registro de Entrega (para Trocas)**
    - Se for uma troca, registra a entrega do novo EPI
    - Rollback em caso de erro no registro

5. **Finalização**
    - Commit da transação em caso de sucesso
    - Rollback automático em qualquer erro durante o processo

