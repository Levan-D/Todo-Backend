package report

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"time"
)

type report struct {
	DB *gorm.DB
}

type Report interface {
	GenerateReportPayment(startAt time.Time, endAt time.Time) (url string, err error)
	GenerateReportTenantPayment(from time.Time, to time.Time) (url string, err error)
	GenerateReportTenantInvoice(tenantInvoiceId uuid.UUID) (string, TenantInvoiceData, error)
	GenerateReportTenantComparisonAct(id uuid.UUID) (string, TenantComparisonActData, error)
}

func NewReport(db *gorm.DB) Report {
	return &report{
		DB: db,
	}
}
