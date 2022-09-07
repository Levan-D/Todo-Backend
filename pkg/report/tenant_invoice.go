package report

import (
	"fmt"
	"github.com/Levan-D/Todo-Backend/pkg/config"
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/Levan-D/Todo-Backend/pkg/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"strings"
	"time"
)

type TenantInvoiceInput struct {
	PeriodFrom         time.Time
	PeriodTo           time.Time
	PaymentCount       int64
	PaymentAmountTotal decimal.Decimal
	TenantName         string
	GeneratedAt        time.Time

	TotalAmount            decimal.Decimal
	TotalUsedPoint         decimal.Decimal
	TotalAcquiredPoint     decimal.Decimal
	TotalAmountDiscounted  decimal.Decimal
	TotalCommissionBank    decimal.Decimal
	TotalCommissionAccount decimal.Decimal

	Payments []TenantInvoicePaymentItem
}

type TenantInvoicePaymentItem struct {
	TenantName             string
	TransactionID          string
	TransactionType        string
	PaymentMethod          string
	CreatedAt              time.Time
	BankCardType           string
	CashbackPercent        decimal.Decimal
	BankCardPan            string
	Total                  decimal.Decimal
	AmountPaid             decimal.Decimal
	PointPaid              decimal.Decimal
	AcquiredPoint          decimal.Decimal
	CommissionBank         decimal.Decimal
	CommissionAccount      decimal.Decimal
	TenantSpentAfterPeriod decimal.Decimal
}

type TenantInvoiceData struct {
	TenantID    uuid.UUID
	GeneratedAt time.Time
	PeriodFrom  time.Time
	PeriodTo    time.Time
}

func (r report) GenerateReportTenantInvoice(tenantInvoiceId uuid.UUID) (string, TenantInvoiceData, error) {
	tenantInvoice, err := r.findTenantInvoiceByID(tenantInvoiceId)
	if err != nil {
		return "", TenantInvoiceData{}, nil
	}

	payments, err := r.findTenantInvoicePaymentByInvoiceID(tenantInvoiceId)
	if err != nil {
		return "", TenantInvoiceData{}, nil
	}

	paymentItems := []TenantInvoicePaymentItem{}

	for _, payment := range payments {
		paymentItems = append(paymentItems, TenantInvoicePaymentItem{
			TenantName:             payment.Tenant.Title,
			TransactionID:          payment.TransactionID,
			TransactionType:        payment.Type,
			PaymentMethod:          payment.PaymentMethodName(),
			CreatedAt:              *payment.CreatedAt,
			BankCardType:           payment.GetBankCardType(),
			CashbackPercent:        payment.CashbackPercent,
			BankCardPan:            payment.GetBankCardMask(), // PAN is full number we haven't
			Total:                  payment.SubTotal,
			AmountPaid:             payment.Amount,
			PointPaid:              payment.UsedPoint,
			AcquiredPoint:          payment.AcquiredPoint,
			CommissionBank:         payment.CommissionBank,
			CommissionAccount:      payment.CommissionAccount,
			TenantSpentAfterPeriod: payment.GetPendingAcquiredPoints(),
		})
	}

	var periodFrom time.Time
	var periodTo time.Time

	if len(payments) > 0 {
		periodFrom = *payments[0].CreatedAt
		periodTo = *payments[len(payments)-1].CreatedAt
	}

	path, err := r.generateReportTenantInvoiceTemplate(TenantInvoiceInput{
		PeriodFrom:             periodFrom,
		PeriodTo:               periodTo,
		PaymentCount:           int64(len(payments)),
		PaymentAmountTotal:     tenantInvoice.TotalSubTotal,
		TenantName:             tenantInvoice.Tenant.Title,
		GeneratedAt:            *tenantInvoice.CreatedAt,
		TotalAmount:            tenantInvoice.TotalAmount,
		TotalUsedPoint:         tenantInvoice.TotalUsedPoint,
		TotalAcquiredPoint:     tenantInvoice.TotalAcquiredPoint,
		TotalAmountDiscounted:  tenantInvoice.TotalAmountDiscounted,
		TotalCommissionBank:    tenantInvoice.TotalCommissionBank,
		TotalCommissionAccount: tenantInvoice.TotalCommissionAccount,
		Payments:               paymentItems,
	})
	if err != nil {
		return "", TenantInvoiceData{}, err
	}

	return path, TenantInvoiceData{
		TenantID:    tenantInvoice.TenantID,
		GeneratedAt: *tenantInvoice.CreatedAt,
		PeriodFrom:  periodFrom,
		PeriodTo:    periodTo,
	}, nil
}

func (r report) generateReportTenantInvoiceTemplate(input TenantInvoiceInput) (path string, err error) {
	sheetName := fmt.Sprintf("INVOICE %s", input.GeneratedAt.Format("2006-01-02"))

	xlsx := excelize.NewFile()
	index := xlsx.NewSheet(sheetName)

	mainBackground, err := xlsx.NewStyle(`{"fill":{"type":"pattern","color":["#1EBECE"],"pattern":1}}`)
	if err != nil {
		return "", fmt.Errorf("cannot set main background, err: %v", err)
	}
	xlsx.SetCellStyle(sheetName, "A5", "O5", mainBackground)

	grayBackground, err := xlsx.NewStyle(`{"fill":{"type":"pattern","color":["#DADADA"],"pattern":1}}`)
	if err != nil {
		return "", fmt.Errorf("cannot set gray background, err: %v", err)
	}
	xlsx.SetCellStyle(sheetName, "A6", "O6", grayBackground)
	xlsx.SetCellStyle(sheetName, "A8", "O8", grayBackground)
	xlsx.SetCellStyle(sheetName, "A10", "O10", grayBackground)
	xlsx.SetCellStyle(sheetName, "A12", "O12", grayBackground)
	xlsx.SetCellStyle(sheetName, "A14", "O14", grayBackground)
	xlsx.SetCellStyle(sheetName, "A16", "O16", grayBackground)

	// Set Column Widths
	xlsx.SetColWidth(sheetName, "A", "A", 50)
	xlsx.SetColWidth(sheetName, "B", "B", 26)
	xlsx.SetColWidth(sheetName, "C", "C", 18)
	xlsx.SetColWidth(sheetName, "D", "D", 18.5)
	xlsx.SetColWidth(sheetName, "E", "E", 15)
	xlsx.SetColWidth(sheetName, "F", "F", 15)
	xlsx.SetColWidth(sheetName, "G", "G", 18)
	xlsx.SetColWidth(sheetName, "H", "H", 15)
	xlsx.SetColWidth(sheetName, "I", "I", 15)
	xlsx.SetColWidth(sheetName, "J", "J", 20)
	xlsx.SetColWidth(sheetName, "K", "K", 20)
	xlsx.SetColWidth(sheetName, "L", "L", 20)
	xlsx.SetColWidth(sheetName, "M", "M", 16)
	xlsx.SetColWidth(sheetName, "N", "N", 24)
	xlsx.SetColWidth(sheetName, "O", "O", 58)

	xlsx.SetCellValue(sheetName, "A1", fmt.Sprintf("%s - რეპორტის დათვლის პერიოდი", input.TenantName))
	xlsx.SetCellValue(sheetName, "A2", fmt.Sprintf("%s - დან %s - მდე", input.PeriodFrom.Format("01/02/2006"), input.PeriodTo.Format("01/02/2006")))

	xlsx.SetCellValue(sheetName, "D1", fmt.Sprintf("გენერაციის თარიღი/დრო: %s", input.GeneratedAt.Format("01/02/2006 15:04:05")))

	xlsx.SetCellValue(sheetName, "A3", "")

	xlsx.SetCellValue(sheetName, "A5", sheetName)

	xlsx.SetCellValue(sheetName, "A6", "ტრანზაქციების რაოდენობა")
	xlsx.SetCellValue(sheetName, "B6", input.PaymentCount)

	xlsx.SetCellValue(sheetName, "A7", "ტრანზაქციების ჯამი")
	xlsx.SetCellValue(sheetName, "B7", input.PaymentAmountTotal)

	xlsx.SetCellValue(sheetName, "A8", "გადახდილი ლარი")
	xlsx.SetCellValue(sheetName, "B8", input.TotalAmount)

	xlsx.SetCellValue(sheetName, "A9", "გადახდილი ლარი (ფასდაკლება)")
	xlsx.SetCellValue(sheetName, "B9", input.TotalAmountDiscounted)

	xlsx.SetCellValue(sheetName, "A10", "გადახდილი ქულები")
	xlsx.SetCellValue(sheetName, "B10", input.TotalUsedPoint)

	xlsx.SetCellValue(sheetName, "A11", "დაგროვილი ქულები")
	xlsx.SetCellValue(sheetName, "B11", input.TotalAcquiredPoint)

	xlsx.SetCellValue(sheetName, "A12", "ბანკის საკომისიო")
	xlsx.SetCellValue(sheetName, "B12", input.TotalCommissionBank)

	xlsx.SetCellValue(sheetName, "A12", "გალერიას საკომისიო")
	xlsx.SetCellValue(sheetName, "B12", input.TotalCommissionAccount)

	xlsx.SetCellValue(sheetName, "A15", "")
	xlsx.SetCellValue(sheetName, "A15", "")

	xlsx.SetCellValue(sheetName, "A16", "ტენანტი")
	xlsx.SetCellValue(sheetName, "B16", "ტრანზაქციის ID")
	xlsx.SetCellValue(sheetName, "C16", "ტრანზაქციის ტიპი")
	xlsx.SetCellValue(sheetName, "D16", "გადახდის მეთოდი")
	xlsx.SetCellValue(sheetName, "E16", "თარიღი")
	xlsx.SetCellValue(sheetName, "F16", "ბარათის ტიპი")
	xlsx.SetCellValue(sheetName, "G16", "ქეშბექის პროცენტი")
	xlsx.SetCellValue(sheetName, "H16", "CARD PAN")
	xlsx.SetCellValue(sheetName, "I16", "სულ ჯამი")
	xlsx.SetCellValue(sheetName, "J16", "გადახდილი თანხა")
	xlsx.SetCellValue(sheetName, "K16", "გადახდილი ქულები")
	xlsx.SetCellValue(sheetName, "L16", "დაგროვილი ქულები")
	xlsx.SetCellValue(sheetName, "M16", "ბანკის საკომისიო")
	xlsx.SetCellValue(sheetName, "N16", "გალერიას საკომისიო")

	paymentRowIndex := 17
	for _, payment := range input.Payments {
		xlsx.SetCellValue(sheetName, fmt.Sprintf("A%d", paymentRowIndex), payment.TenantName)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("B%d", paymentRowIndex), payment.TransactionID)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("C%d", paymentRowIndex), payment.TransactionType)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("D%d", paymentRowIndex), payment.PaymentMethod)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("E%d", paymentRowIndex), payment.CreatedAt)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("F%d", paymentRowIndex), payment.BankCardType)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("G%d", paymentRowIndex), payment.CashbackPercent)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("H%d", paymentRowIndex), payment.BankCardPan)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("I%d", paymentRowIndex), payment.Total)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("J%d", paymentRowIndex), payment.AmountPaid)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("K%d", paymentRowIndex), payment.PointPaid)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("L%d", paymentRowIndex), payment.AcquiredPoint)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("M%d", paymentRowIndex), payment.CommissionBank)
		xlsx.SetCellValue(sheetName, fmt.Sprintf("N%d", paymentRowIndex), payment.CommissionAccount)
		paymentRowIndex++
	}

	xlsx.SetActiveSheet(index)
	xlsx.DeleteSheet("Sheet1")

	tenantCode := strings.ReplaceAll(strings.ToLower(string(input.TenantName)), " ", "-")
	xlsxFilePath := fmt.Sprintf("%s/invoice-%s-%s-%s.xlsx", config.GetDirectoryPath("tmp"), tenantCode, input.PeriodFrom.Format("2006-02-01"), utils.GenerateRandomNumbers(5))
	//zippedFilePath := fmt.Sprintf("%s/invoice-%s-%s.zip", config.GetDirectoryPath("tmp"), strings.ToLower(string(input.TenantName)), input.PeriodFrom.Format("2006-02-01"))

	if err := xlsx.SaveAs(xlsxFilePath); err != nil {
		return "", fmt.Errorf("cannot create xlsx file, err: %v", err)
	}

	//err = files_zip.ZipFile(xlsxFilePath, zippedFilePath)
	//if err != nil {
	//	return "", fmt.Errorf("cannot be zipped report xlsx file, err: %v", err)
	//}
	//defer os.Remove(xlsxFilePath)

	return xlsxFilePath, nil
}

func (r report) findTenantInvoiceByID(id uuid.UUID) (domain.TenantInvoice, error) {
	var invoice domain.TenantInvoice
	err := r.DB.Preload("Tenant").Where("id = ?", id).First(&invoice).Error
	return invoice, err
}

func (r report) findTenantInvoicePaymentByInvoiceID(invoiceId uuid.UUID) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := r.DB.Preload("Tenant").Preload("Gateway").Where("tenant_invoice_id = ?", invoiceId).Find(&payments).Error
	return payments, err
}
