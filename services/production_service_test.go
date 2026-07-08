package services

import (
	"donapresentes/models"
	"testing"
)

func TestProductionTransitionsByEngraving(t *testing.T) {
	tests := []struct {
		name, from, to        string
		engraving, firstPiece bool
		wantErr               bool
	}{
		{"sem gravação segue expedição", models.ProductionMaterialOK, models.ProductionReadyShipment, false, false, false},
		{"com gravação não pula fluxo", models.ProductionMaterialOK, models.ProductionReadyShipment, true, false, true},
		{"com gravação separa", models.ProductionMaterialOK, models.ProductionPreparingEngraving, true, false, false},
		{"primeira peça obrigatória", models.ProductionSentEngraving, models.ProductionEngraving, true, true, true},
		{"sem primeira peça grava", models.ProductionSentEngraving, models.ProductionEngraving, true, false, false},
		{"não retrocede", models.ProductionShipped, models.ProductionReadyShipment, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidProductionTransition(tt.from, tt.to, tt.engraving, tt.firstPiece)
			if (err != nil) != tt.wantErr {
				t.Fatalf("erro=%v, esperado erro=%v", err, tt.wantErr)
			}
		})
	}
}

func TestProductionAccessProfiles(t *testing.T) {
	if !(models.ProductionAccess{Production: true}).CanOperate() {
		t.Fatal("produção deve operar")
	}
	if !(models.ProductionAccess{Inspection: true}).CanOperate() {
		t.Fatal("conferência deve operar")
	}
	if (models.ProductionAccess{Sales: true}).CanOperate() {
		t.Fatal("vendas não deve operar recebimento")
	}
	if !(models.ProductionAccess{Driver: true}).CanView() {
		t.Fatal("motorista deve visualizar ordens atribuídas")
	}
}

func TestCreateSupplyRequiresAuditedOpeningBalance(t *testing.T) {
	service := NewProductionService(nil)
	_, err := service.CreateSupply(models.ProductionSupplyInput{
		Name:            "Caixa",
		Unit:            "UN",
		CurrentQuantity: 10,
	}, models.ProductionAccess{Admin: true})
	if err == nil {
		t.Fatal("initial balance must be recorded by an entry with supplier and invoice")
	}
}
