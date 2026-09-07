package http

import (
	"fmt"
)

func renderQuickRevokeHTML(message string) string {
	msg := message
	if msg == "" {
		msg = "Sessão revogada com sucesso."
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>KeepGuard - Dispositivo Revogado</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
        .card { background: #1e293b; padding: 2.5rem; border-radius: 12px; box-shadow: 0 10px 25px -5px rgba(0,0,0,0.5); max-width: 480px; text-align: center; border: 1px solid #334155; }
        .icon { font-size: 3.5rem; margin-bottom: 1rem; color: #ef4444; }
        h1 { font-size: 1.5rem; margin-bottom: 0.75rem; color: #ffffff; }
        p { color: #94a3b8; line-height: 1.5; font-size: 0.95rem; margin-bottom: 1.5rem; }
        .badge { display: inline-block; padding: 0.35rem 0.75rem; background: #334155; color: #38bdf8; border-radius: 9999px; font-size: 0.85rem; font-weight: 600; margin-bottom: 1.5rem; }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">&#128737;</div>
        <h1>Acesso Revogado com Sucesso</h1>
        <div class="badge">Dispositivo Bloqueado</div>
        <p>%s O acesso desta sessão foi encerrado e este dispositivo foi adicionado à sua lista de bloqueios.</p>
        <p style="font-size:0.85rem; color:#64748b;">Sua conta permanece protegida.</p>
    </div>
</body>
</html>`, msg)
}
