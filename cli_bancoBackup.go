package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
)

func ExecutarBackupBanco(args []string) {
	conf := configs.NewVariaveisAmbiente()

	bucket := os.Getenv("R2_BUCKET_NAME_BACKUPS")
	if bucket == "" {
		log.Fatal("R2_BUCKET_NAME_BACKUPS não configurado")
	}

	ctx := context.Background()
	helper.InitR2_cloudflare(ctx)

	// Arquivo temporário em vez de pipe direto pro upload: o SDK da AWS
	// calcula checksum/assinatura melhor sobre um io.ReadSeeker, e um dump
	// de tenant novo é pequeno o bastante pra caber tranquilo no disco do
	// Railway. Se o banco crescer a ponto de doer, well: você já vai ter
	// coisa mais importante pra resolver que isso aqui.
	tmp, err := os.CreateTemp("", "backup-*.dump")
	if err != nil {
		log.Fatalf("erro ao criar arquivo temporário: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	cmd := exec.CommandContext(ctx, "pg_dump",
		"--host="+conf.DB_SERVER,
		"--port="+conf.DB_PORT,
		"--username="+conf.DB_USER,
		"--dbname="+conf.DATABASE,
		"--format=custom",
		"--no-owner",
		"--no-privileges",
	)
	cmd.Env = append(os.Environ(),
		"PGPASSWORD="+conf.DB_PASSWORD,
		"PGSSLMODE="+conf.DB_SSLMODE,
	)
	cmd.Stdout = tmp
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("pg_dump falhou: %v", err)
	}

	if _, err := tmp.Seek(0, 0); err != nil {
		log.Fatalf("erro ao rebobinar o dump: %v", err)
	}

	key := fmt.Sprintf("backups/%s.dump", time.Now().UTC().Format("20060102-150405"))
	if err := helper.UploadArquivo(ctx, bucket, key, tmp, "application/octet-stream"); err != nil {
		log.Fatalf("erro ao subir backup pro R2: %v", err)
	}

	fmt.Printf("backup salvo em %s/%s\n", bucket, key)
}
