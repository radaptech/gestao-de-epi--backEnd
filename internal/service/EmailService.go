package service

import (
	"fmt"
	"net/smtp"
	"os"
)

// Estrutura do seu serviço de e-mail
type EmailService struct {
	Host string
	Port string
	// Em produção, você adicionaria User e Password aqui
}


func NewEmail(host, port string) *EmailService{

	return &EmailService{
		Host: host,
		Port: port,
	}
}
// Função responsável por montar e enviar o e-mail de recuperação
func (e *EmailService) EnviarEmailRecuperacao(emailDestino string, token string, slugEmpresa string) error {
	// 1. Monta o link dinâmico
	
	URL:= os.Getenv("URL_FRONTEND_FORMATO")

	if URL == "" {
		URL = "http://%s:3000"
	}
	linkBase := fmt.Sprintf(URL, slugEmpresa)
	link:= fmt.Sprintf("%s/redefinir-senha?token=%s", linkBase, token)

	// 2. Configura os cabeçalhos do e-mail (Assunto e tipo HTML)
	remetente := "nao-responda@radaptech.com.br"
	assunto := "Subject: Recuperação de Senha - SGE\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	// 3. Monta o corpo do e-mail em HTML simples
	corpo := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>SGE - Sistema de Gestão de EPIs</h2>
			<p>Olá,</p>
			<p>Recebemos um pedido para redefinir a senha da sua conta.</p>
			<p>Clique no botão abaixo para criar uma nova senha:</p>
			<a href="%s" style="background-color: #007BFF; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block;">Redefinir Senha</a>
			<br><br>
			<p>Se você não fez essa solicitação, pode ignorar este e-mail em segurança.</p>
		</div>
	`, link)

	// Junta tudo em um array de bytes
	mensagem := []byte(assunto + mime + corpo)

	// 4. Conecta no servidor SMTP e dispara
	enderecoSMTP := fmt.Sprintf("%s:%s", e.Host, e.Port)
	
	// Como estamos usando o Mailpit local (porta 1025), a autenticação (nil) não é necessária.
	// Quando for pra produção, você usará smtp.PlainAuth() aqui.
	err := smtp.SendMail(enderecoSMTP, nil, remetente, []string{emailDestino}, mensagem)
	
	return err
}