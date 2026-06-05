# Endpoints da API

Lista de todos os endpoints da API, agrupados por categoria.

## Grupo: painel

| Método | Caminho | Controller | Método do Controller | Descrição (implícita) |
|--------|---------|------------|----------------------|-----------------------|
| POST | `/cadastrar-planos` | - | - | Endpoint |
| GET | `/planos` | - | - | Endpoint |
| PUT | `planos/:id` | - | - | Endpoint |
| PATCH | `planos/:id/status` | - | - | Endpoint |
| POST | `/cadastrar-empresa` | - | - | Endpoint |

## Grupo: api

| Método | Caminho | Controller | Método do Controller | Descrição (implícita) |
|--------|---------|------------|----------------------|-----------------------|
| POST | `/login` | - | - | Endpoint |
| POST | `/cadastro` | - | - | Endpoint |
| POST | `/esqueci-minha-senha` | - | - | Endpoint |
| POST | `/redefinir-senha` | - | - | Endpoint |
| GET | `/me` | - | - | Endpoint |
| GET | `/departamentos` | - | - | Endpoint |
| GET | `/quantidade-epi` | - | - | Endpoint |
| GET | `/saldo-epi` | - | - | Endpoint |
| GET | `/funcoes` | - | - | Endpoint |
| GET | `/funcionarios` | - | - | Endpoint |
| GET | `/funcionario/:matricula` | - | - | Endpoint |
| GET | `/funcionarios-dashbord` | - | - | Endpoint |
| GET | `/funcionario-estoque` | - | - | Endpoint |
| GET | `/tamanhos` | - | - | Endpoint |
| GET | `/tamanhos-id-epi/:id` | - | - | Endpoint |
| GET | `/tamanho/:id` | - | - | Endpoint |
| GET | `/protecoes` | - | - | Endpoint |
| GET | `/protecao/:id` | - | - | Endpoint |
| GET | `/epis` | - | - | Endpoint |
| GET | `/epi/:id` | - | - | Endpoint |
| GET | `/epis-dashbord` | - | - | Endpoint |
| GET | `/funcionarios/:id/epis` | - | - | Endpoint |
| GET | `/entradas` | - | - | Endpoint |
| GET | `/entradas-dashbord` | - | - | Endpoint |
| GET | `/entradas-estoque` | - | - | Endpoint |
| GET | `/fornecedores` | - | - | Endpoint |
| GET | `/entregas` | - | - | Endpoint |
| POST | `/cadastro-entregas` | - | - | Endpoint |
| GET | `/entregas-dashbord` | - | - | Endpoint |
| GET | `/entrega-itens-dashbord` | - | - | Endpoint |
| POST | `/cadastro-devolucao` | - | - | Endpoint |
| GET | `/devolucao` | - | - | Endpoint |
| POST | `/cadastrar-motivo-devolucao` | - | - | Endpoint |
| GET | `/motivos` | - | - | Endpoint |

## Grupo: rotasAdm

| Método | Caminho | Controller | Método do Controller | Descrição (implícita) |
|--------|---------|------------|----------------------|-----------------------|
| DELETE | `/departamento/:id` | - | - | Endpoint |
| PUT | `/departamento/:id` | - | - | Endpoint |
| POST | `/cadastro-departamento` | - | - | Endpoint |
| DELETE | `/funcao/:id` | - | - | Endpoint |
| PUT | `/funcao/:id` | - | - | Endpoint |
| POST | `/cadastro-funcao` | - | - | Endpoint |
| DELETE | `/funcionario/:id` | - | - | Endpoint |
| PATCH | `/funcionario/:id` | - | - | Endpoint |
| POST | `/cadastro-funcionario` | - | - | Endpoint |
| POST | `/cadastro-tamanho` | - | - | Endpoint |
| DELETE | `/tamanho/:id` | - | - | Endpoint |
| POST | `/cadastro-protecao` | - | - | Endpoint |
| DELETE | `/protecao/:id` | - | - | Endpoint |
| DELETE | `/epi/:id` | - | - | Endpoint |
| PATCH | `/epi/:id` | - | - | Endpoint |
| POST | `/cadastro-epi` | - | - | Endpoint |
| POST | `/cadastrar-entrada` | - | - | Endpoint |
| DELETE | `/entrada/:id` | - | - | Endpoint |
| POST | `/cadastro-fornecedores` | - | - | Endpoint |
| DELETE | `/fornecedor/:id` | - | - | Endpoint |
| PATCH | `/fornecedor/:id` | - | - | Endpoint |
| DELETE | `/entrega/:id` | - | - | Endpoint |
| GET | `/:matricula/ficha-pdf/:id` | - | - | Endpoint |
| GET | `/usuarios` | - | - | Endpoint |
| GET | `/devolucoes/:id/pdf` | - | - | Endpoint |


## Notas sobre Payloads e Respostas

- Os payloads de entrada (JSON) são definidos pelas structs de entrada (ex: `FuncionarioInserir`, `EntradaEpiInserir`) presentes nos arquivos em `internal/model/`;
- As respostas são geralmente JSON contendo os DTOs (ex: `Funcionario_Dto`, `EntradaEpiDto`) ou mensagens de sucesso;
- Para detalhes exatos de cada campo, consulte o documento `Modelos_de_Dados.md`;
- Autenticação: a maioria dos endpoints requer token JWT no header `Authorization: Bearer <token>`;
- Multitenancy: todos os endpoints (exceto os de painel e autenticação pública) requerem o header `X-tenant-ID` contendo o subdomínio da empresa.