# Backup do banco → Cloudflare R2

Dump completo do Postgres (`pg_dump --format=custom`) enviado para um bucket R2 da Cloudflare,
acessado via S3 API com o SDK da AWS (`aws-sdk-go-v2`).

| Arquivo | Papel |
|---|---|
| `cli_bancoBackup.go` | `ExecutarBackupBanco` — roda o `pg_dump` num arquivo temporário e sobe pro R2 |
| `internal/helper/R2_backups.go` | `InitR2_cloudflare` (cliente S3 apontado pro endpoint do R2) e `UploadArquivo` |
| `main.go` | Despacha o subcomando: `./main backup-banco` → `ExecutarBackupBanco`, sem argumento → sobe a API |

---

## Como funciona

1. `configs.NewVariaveisAmbiente()` carrega o `.env` (ou as variáveis de sistema, em produção).
   **Isso precisa acontecer antes do `InitR2_cloudflare`** — o helper lê as credenciais com
   `os.Getenv`, então inverter a ordem entrega chave vazia e o upload volta 401.
2. `pg_dump --format=custom --no-owner --no-privileges` escreve num `os.CreateTemp`
   (`backup-*.dump`), removido no `defer`.
   O arquivo temporário existe em vez de um pipe direto porque o SDK da AWS assina/calcula checksum
   melhor sobre um `io.ReadSeeker`.
3. `helper.UploadArquivo` faz `PutObject` na key `backups/YYYYMMDD-HHMMSS.dump` (UTC), com
   `Content-Type: application/octet-stream`.

Qualquer falha é `log.Fatal` — o processo sai com código ≠ 0, que é o que o agendador precisa ver
para marcar a execução como falha.

## Variáveis de ambiente

| Variável | Uso |
|---|---|
| `R2_IDCLOUDFLARE` | ID da conta Cloudflare; monta o endpoint `https://<id>.r2.cloudflarestorage.com` |
| `R2_KEYID` / `R2_SECRETKEY` | Credenciais do token de API R2 (Access Key ID / Secret) |
| `R2_BUCKET_NAME_BACKUPS` | Nome do bucket de destino |

A região é fixa em `auto` — o R2 não tem regiões no sentido da AWS.

## Rodando

Em **produção** (Railway) é um cron job na mesma imagem que já serve a API:

```bash
./main backup-banco
```

Na sua máquina, se você tiver `pg_dump` no PATH, o atalho é o target do `makefile`:

```bash
make backup-banco    # = go run . backup-banco
```

**Mas o `pg_dump` provavelmente não existe na sua máquina** — ele vem do pacote
`postgresql17-client` instalado no `Dockerfile` de produção. Então teste pela imagem:

```bash
# 1. Compila a imagem de produção (mesmo Dockerfile do deploy).
#    É aqui que se descobre se o Go compila e se o pg_dump entrou na imagem.
docker build -t sgeepi-backup-test .

# 2. Roda só o subcomando de backup e descarta o container.
docker run --rm --network sgeepi-infra_default --env-file .env sgeepi-backup-test ./main backup-banco
```

- `--rm` — comando de uma vez, não é um serviço; some ao terminar.
- `--network sgeepi-infra_default` — mesma rede do `postgres_container`, senão o
  `DB_SERVER=postgres` do `.env` não resolve.
- `--env-file .env` — injeta as variáveis como variáveis de **sistema**; o log vai dizer
  "arquivo .env não encontrado, continuando com variáveis de sistema", que é exatamente o
  comportamento em produção.
- `./main backup-banco` — sobrescreve o `CMD ["./main"]` da imagem. Sem o argumento, sobe a API inteira.

Saída esperada:

```
backup salvo em backups-sge/backups/20260825-004355.dump
```

## Conferindo o que subiu

Sem instalar a CLI da AWS, com `curl` (assina SigV4 sozinho):

```bash
set -a; . ./.env; set +a
curl -I --aws-sigv4 "aws:amz:auto:s3" --user "$R2_KEYID:$R2_SECRETKEY" \
  "https://$R2_IDCLOUDFLARE.r2.cloudflarestorage.com/$R2_BUCKET_NAME_BACKUPS/backups/<arquivo>.dump"
```

`HTTP/1.1 200` + `Content-Length` > 0 significa que o objeto está lá.

## Restaurando

O dump está em formato custom, então é `pg_restore` (não `psql`). Rodando pela mesma imagem:

```bash
docker run --rm --network sgeepi-infra_default --env-file .env \
  -v "$PWD/arquivo.dump:/tmp/arquivo.dump" sgeepi-backup-test \
  sh -c 'PGPASSWORD=$DB_PASSWORD pg_restore --host=$DB_SERVER --username=$DB_USER \
         --dbname=$DATABASE --clean --if-exists --no-owner /tmp/arquivo.dump'
```

> Restauração ainda não foi exercitada de ponta a ponta neste projeto — teste num banco descartável
> antes de precisar dela pra valer.

## Não implementado (de propósito)

- **Agendamento**: nada no código dispara o backup. Quem chama `./main backup-banco` é o cron do Railway.
- **Retenção/expiração**: resolvida por regra de lifecycle no próprio bucket R2, não em código.
