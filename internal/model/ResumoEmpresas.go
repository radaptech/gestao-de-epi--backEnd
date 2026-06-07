package model



type ResumoDashboard struct {
	TotalEmpresas         int     `json:"totalEmpresas"`
	EmpresasAtivas        int     `json:"empresasAtivas"`
	EmpresasBloqueadas    int     `json:"empresasBloqueadas"`
	EmpresasEmTeste       int     `json:"empresasEmTeste"`
	TotalFuncionarios     int     `json:"totalFuncionarios"`
	TotalEpis             int     `json:"totalEpis"`
	TotalEntregas         int     `json:"totalEntregas"`
	MensalidadesPagas     int     `json:"mensalidadesPagas"`
	MensalidadesAtrasadas int     `json:"mensalidadesAtrasadas"`
	ReceitaMensal         float64 `json:"receitaMensal"`
}