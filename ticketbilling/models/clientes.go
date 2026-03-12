package models

type Cliente struct {
	ClienteID          uint   `gorm:"column:cliente_id;primaryKey;autoIncrement"`
	Nombre             string `gorm:"column:nombre"`
	Email              string `gorm:"column:email"`
	RFC                string `gorm:"column:rfc"`
	RegimenFiscal      string `gorm:"column:regimen_fiscal"`
	DescripcionRegimen string `gorm:"column:descripcion_regimen"`
	Street             string `gorm:"column:street"`
	HouseNumber        string `gorm:"column:house_number"`
	City               string `gorm:"column:city"`
	PostalCode         string `gorm:"column:postal_code"`
	ExternalID         string `gorm:"column:external_id"`
}

func (Cliente) TableName() string {
	return "clientes"
}
