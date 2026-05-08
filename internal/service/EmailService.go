package service

import (
	"fmt"
	
	"os"
	"github.com/resend/resend-go/v2"
)

// Estrutura do seu serviço de e-mail
type EmailService struct {

}


func NewEmail() *EmailService{

	return &EmailService{}
}
// Função responsável por montar e enviar o e-mail de recuperação
func (e *EmailService) EnviarEmailRecuperacao(emailDestino string, token string, slugEmpresa string) error {
	// 1. Monta o link dinâmico
	
	URL:= os.Getenv("URL_FRONTEND_FORMATO")
	apikey:= os.Getenv("RESEND_API_KEY")
	client:= resend.NewClient(apikey)

	if URL == "" {
		URL = "http://%s:3000"
	}
	linkBase := fmt.Sprintf(URL, slugEmpresa)
	link:= fmt.Sprintf("%s/redefinir-senha?token=%s", linkBase, token)

	// 2. Configura os cabeçalhos do e-mail (Assunto e tipo HTML)
	//emetente := "nao-responda@radaptech.com.br"
	assunto := "	Recuperação de Senha - SGE"
	//mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	// 3. Monta o corpo do e-mail em HTML simples
	corpo := fmt.Sprintf(`
		<div style="background-color: #f8fafc; padding: 40px 20px; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;">
			
			<div style="max-width: 500px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);">
				
				<!-- Barra superior de destaque (Mesmo efeito do React bg-slate-500) -->
				<div style="height: 6px; background-color: #64748b; width: 100%%;"></div>

				<div style="padding: 40px 30px;">
					<!-- Header do Email -->
					<div style="text-align: center; margin-bottom: 30px;">
						<h1 style="margin: 0; color: #1e293b; font-size: 28px; font-weight: 800; letter-spacing: -0.5px;">SGEPI</h1>
						<p style="margin: 4px 0 0 0; color: #94a3b8; font-size: 12px; text-transform: uppercase; letter-spacing: 1.5px; font-weight: bold;">Recuperação de Senha</p>
					</div>

					<!-- Conteúdo -->
					<p style="color: #334155; font-size: 16px; line-height: 24px; margin: 0 0 20px 0;">Olá,</p>
					<p style="color: #334155; font-size: 16px; line-height: 24px; margin: 0 0 30px 0;">Recebemos uma solicitação para redefinir a senha da sua conta no sistema de gestão. Para criar uma nova senha, clique no botão abaixo:</p>

					<!-- Botão Principal (Mesma cor do bg-slate-800) -->
					<div style="text-align: center; margin-bottom: 30px;">
						<a href="%s" style="background-color: #1e293b; color: #ffffff; padding: 14px 28px; text-decoration: none; border-radius: 8px; display: inline-block; font-weight: bold; font-size: 16px;">Criar nova senha</a>
					</div>

					<!-- Link de Backup (Boa prática de UX para e-mails corporativos) -->
					<p style="color: #64748b; font-size: 14px; line-height: 22px; margin: 0 0 10px 0;">Ou copie e cole o link abaixo no seu navegador:</p>
					<p style="margin: 0; word-break: break-all;"><a href="%s" style="color: #3b82f6; font-size: 13px; text-decoration: underline;">%s</a></p>

					<!-- Rodapé Interno do Card -->
					<div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #e2e8f0;">
						<p style="color: #94a3b8; font-size: 12px; line-height: 18px; margin: 0;">Se você não solicitou esta alteração, nenhuma ação é necessária. Sua senha continuará a mesma em segurança.</p>
					</div>
				</div>
			</div>
			
			<!-- Rodapé Externo -->
			<div style="text-align: center; margin-top: 20px;">
				<p style="color: #94a3b8; font-size: 11px; font-weight: bold; text-transform: uppercase;">SGEPI - Gestão de Estoque © 2026</p>
			</div>
			
		</div>
	`, link, link, link)



	params:= &resend.SendEmailRequest{
		From: "SGEPI <onboarding@resend.dev>",
		To: []string{emailDestino},
		Subject: assunto,
		Html: corpo,
	}

	_, err:= client.Emails.Send(params)
	
	return err
}