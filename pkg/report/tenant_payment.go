package report

import (
	"fmt"
	"github.com/Levan-D/Todo-Backend/pkg/config"
	"github.com/Levan-D/Todo-Backend/pkg/database/postgres"
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/Levan-D/Todo-Backend/pkg/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"sort"
	"strings"
	"time"
)

type TenantPaymentStatistic struct {
	PaymentsAmount decimal.Decimal `json:"payments_amount" gorm:"column:payments_amount"`
	UsedPoints     decimal.Decimal `json:"used_points" gorm:"column:used_points"`
	AcquiredPoints decimal.Decimal `json:"acquired_points" gorm:"column:acquired_points"`
}

type TenantPaymentInput struct {
	PeriodFrom         time.Time
	PeriodTo           time.Time
	PaymentCount       int64
	PaymentAmountTotal decimal.Decimal
	GeneratedAt        time.Time

	TransactionCount             int64
	TotalSubTotal                decimal.Decimal
	TotalAmount                  decimal.Decimal
	TotalAmountDiscounted        decimal.Decimal
	TotalUsedPoint               decimal.Decimal
	TotalAcquiredPoint           decimal.Decimal
	TotalAcquiredPointDiscounted decimal.Decimal
	TotalCommissionBank          decimal.Decimal
	TotalCommissionAccount       decimal.Decimal
	TenantUnpaidTotal            decimal.Decimal
	TenantPaidTotal              decimal.Decimal
	ParkingTransactionCount      int64
	ParkingTransactionTotal      decimal.Decimal

	ReturnedUsedPoint        decimal.Decimal
	ReturnedAmount           decimal.Decimal
	ReturnedAmountDiscounted decimal.Decimal

	VisitorFemaleCount       int64
	VisitorFemaleAmountSpent decimal.Decimal
	VisitorMaleCount         int64
	VisitorMaleAmountSpent   decimal.Decimal

	GiftClaimedCount   int64
	GiftUsedCount      int64
	TicketClaimedCount int64
	TicketUsedCount    int64

	Payments []TenantPaymentItem
}

type TenantPaymentItem struct {
	TenantName                string
	TransactionID             string
	SourceRRN                 string
	SourceTransactionID       string
	TransactionType           string
	PaymentMethod             string
	CreatedAt                 time.Time
	PaidAt                    time.Time
	ReturnExpiredAt           time.Time
	ReturnExpiredDuration     int64
	BankCardType              string
	AmountDiscounted          decimal.Decimal
	CashbackPercent           decimal.Decimal
	CashbackDiscountedPercent decimal.Decimal
	BankCardPan               string
	Status                    string
	InvoiceStatus             string
	Total                     decimal.Decimal
	Amount                    decimal.Decimal
	UsedPoint                 decimal.Decimal
	AcquiredPoint             decimal.Decimal
	AcquiredPointDiscounted   decimal.Decimal
	AcquiredPointStatus       domain.PaymentAcquiredPointStatusType
	CommissionBank            decimal.Decimal
	CommissionBankPercent     decimal.Decimal
	CommissionAccount         decimal.Decimal
	CommissionAccountPercent  decimal.Decimal
	ReturnExpiredStatus       string
	TenantSpentAmount         decimal.Decimal
	MainTransactionID         string
	ReturnTransactionID       string
}

func (r report) GenerateReportTenantPayment(from time.Time, to time.Time) (url string, err error) {
	generatedAt := time.Now()

	var payments []domain.Payment
	err = postgres.GetDB().
		Preload("Tenant").Preload("User").Preload("Gateway").Preload("BankCard").Preload("ParentPayment").Preload("ChildPayment").
		Where("type = ? OR type = ? OR type = ?", domain.PaymentTypeTenantPayment, domain.PaymentTypeTenantReturn, domain.PaymentTypeParkingPayment).
		Where("status = ? OR status = ?", domain.PaymentStatusPaid, domain.PaymentStatusReturned).
		Where("created_at >= ? AND created_at <= ?", from, to).
		Order("created_at asc").
		Find(&payments).Error
	if err != nil {
		return "", err
	}

	var eventVouchers []domain.EventVoucher
	err = postgres.GetDB().Preload("Event").Where("updated_at >= ? AND updated_at <= ?", from, to).Find(&eventVouchers).Error
	if err != nil {
		return "", err
	}

	var totalStats TenantPaymentStatistic
	err = postgres.GetDB().Model(&domain.Payment{}).
		Select("SUM(CASE WHEN (type = 'TENANT_PAYMENT' OR type = 'TENANT_RETURN' OR type = 'PARKING_PAYMENT') THEN amount ELSE -amount END) as payments_amount, "+
			"SUM(CASE WHEN (type = 'TENANT_PAYMENT' OR type = 'TENANT_RETURN' OR type = 'PARKING_PAYMENT') THEN used_point ELSE -used_point END) as used_points, "+
			"SUM(CASE WHEN (type = 'TENANT_PAYMENT' OR type = 'TENANT_RETURN' OR type = 'PARKING_PAYMENT') THEN acquired_point ELSE -acquired_point END) as acquired_points").
		Where("(status = 'PAID' OR status = 'RETURNED') AND (type = 'TENANT_PAYMENT' OR type = 'TENANT_RETURN' OR type = 'PARKING_PAYMENT')").
		Where("created_at >= ? AND created_at <= ?", from, to).
		Find(&totalStats).Error

	if err != nil {
		return "", err
	}

	var paidStats TenantPaymentStatistic
	err = postgres.GetDB().Model(&domain.TenantInvoice{}).
		Select("SUM(total_amount) as payments_amount, SUM(total_used_point) as used_points, SUM(total_acquired_point) as acquired_points").
		Where("status = ?", domain.InvoiceStatusPaid).
		Where("created_at >= ? AND created_at <= ?", from, to).
		Find(&paidStats).Error

	if err != nil {
		return "", err
	}

	tenantPayments := make(map[uuid.UUID]decimal.Decimal)
	tenantNames := make(map[uuid.UUID]string)
	maleSpent := make(map[uuid.UUID]decimal.Decimal)
	femaleSpent := make(map[uuid.UUID]decimal.Decimal)

	paymentList := make([]TenantPaymentItem, 0, len(payments))

	paymentSubTotalSum := decimal.New(0, 0)
	paymentAmountSum := decimal.New(0, 0)
	paymentAmountDiscountedSum := decimal.New(0, 0)
	paymentUsedPointSum := decimal.New(0, 0)
	paymentAcquiredPointSum := decimal.New(0, 0)
	paymentAcquiredPointDiscountedSum := decimal.New(0, 0)
	returnedAmountSum := decimal.New(0, 0)
	returnedAmountDiscountedSum := decimal.New(0, 0)
	returnedUsedPoints := decimal.New(0, 0)
	var parkingTransactionsCount int64
	parkingTransactionsAmount := decimal.New(0, 0)
	commissionBank := decimal.New(0, 0)
	commissionAccount := decimal.New(0, 0)
	totalTenantPaid := decimal.New(0, 0)
	totalTenantUnPaid := decimal.New(0, 0)

	for _, payment := range payments {
		if payment.Tenant != nil {
			tenantPayments[payment.Tenant.ID] = tenantPayments[payment.Tenant.ID].Add(payment.SubTotal)
			tenantNames[payment.Tenant.ID] = payment.Tenant.Title
		}

		if (payment.Status == domain.PaymentStatusPaid || payment.Status == domain.PaymentStatusReturned) && payment.Type != domain.PaymentTypeCardAdd {

			tenantName := "UNDEFINED"
			if payment.Tenant != nil {
				tenantName = payment.Tenant.Title
			}

			returnExpiredStatus := "-"
			if payment.ReturnExpiredAt != nil {
				if time.Now().Before(*payment.ReturnExpiredAt) {
					returnExpiredStatus = "PENDING"
				} else {
					returnExpiredStatus = "VALID"
				}
			}

			parentPayment := ""
			childPayment := ""

			if payment.ParentPayment != nil {
				parentPayment = payment.ParentPayment.TransactionID
			}
			if payment.ChildPayment != nil {
				childPayment = payment.ChildPayment.TransactionID
			}

			returnExpiredAt := time.Time{}
			if payment.ReturnExpiredAt != nil {
				returnExpiredAt = *payment.ReturnExpiredAt
			}

			paidAt := time.Time{}
			if payment.PaidAt != nil {
				paidAt = *payment.PaidAt
			}

			paymentList = append(paymentList, TenantPaymentItem{
				TenantName:                tenantName,
				TransactionID:             payment.TransactionID,
				SourceRRN:                 payment.SourceRRN,
				SourceTransactionID:       payment.SourceTransactionID,
				TransactionType:           payment.Type,
				PaymentMethod:             payment.PaymentMethodName(),
				CreatedAt:                 *payment.CreatedAt,
				BankCardType:              payment.GetBankCardType(),
				CashbackPercent:           payment.CashbackPercent,
				CashbackDiscountedPercent: payment.CashbackDiscountedPercent,
				BankCardPan:               payment.GetBankCardMask(), // PAN is full number we haven't
				UsedPoint:                 payment.UsedPoint,
				Amount:                    payment.Amount,
				AmountDiscounted:          payment.AmountDiscounted,
				AcquiredPoint:             payment.AcquiredPoint,
				AcquiredPointDiscounted:   payment.AcquiredPointDiscounted,
				AcquiredPointStatus:       payment.AcquiredPointStatus,
				CommissionBank:            payment.CommissionBank,
				CommissionBankPercent:     payment.CommissionBankPercent,
				CommissionAccount:         payment.CommissionAccount,
				CommissionAccountPercent:  payment.CommissionAccountPercent,
				Total:                     payment.SubTotal,
				Status:                    payment.Status,
				InvoiceStatus:             string(payment.TenantInvoiceStatus),
				TenantSpentAmount:         payment.SubTotalTenantSpent,
				ReturnExpiredStatus:       returnExpiredStatus,
				PaidAt:                    paidAt,
				ReturnExpiredAt:           returnExpiredAt,
				ReturnExpiredDuration:     payment.ReturnExpiredDuration,
				MainTransactionID:         parentPayment,
				ReturnTransactionID:       childPayment,
			})
		}

		switch payment.User.Gender {
		default:
			// todo: log unknown gender?
		case domain.UserGenderFemale:
			femaleSpent[*payment.UserID] = femaleSpent[*payment.UserID].Add(payment.Amount.Add(payment.AmountDiscounted))
		case domain.UserGenderMale:
			maleSpent[*payment.UserID] = maleSpent[*payment.UserID].Add(payment.Amount.Add(payment.AmountDiscounted))
		}

		switch payment.Type {
		case domain.PaymentTypeTenantPayment:
			paymentSubTotalSum = paymentSubTotalSum.Add(payment.SubTotal)
			paymentAmountSum = paymentAmountSum.Add(payment.Amount)
			paymentAmountDiscountedSum = paymentAmountDiscountedSum.Add(payment.AmountDiscounted)
			paymentUsedPointSum = paymentUsedPointSum.Add(payment.UsedPoint)
			paymentAcquiredPointSum = paymentAcquiredPointSum.Add(payment.AcquiredPoint)
			paymentAcquiredPointDiscountedSum = paymentAcquiredPointDiscountedSum.Add(payment.AcquiredPointDiscounted)
		case domain.PaymentTypeTenantReturn:
			returnedAmountSum = returnedAmountSum.Add(payment.Amount)
			returnedAmountDiscountedSum = returnedAmountDiscountedSum.Add(payment.AmountDiscounted)
			returnedUsedPoints = returnedUsedPoints.Add(payment.UsedPoint)
		case domain.PaymentTypeParkingPayment:
			parkingTransactionsCount += 1
			parkingTransactionsAmount = parkingTransactionsAmount.Add(payment.Amount.Add(payment.AmountDiscounted))
		}

		commissionBank = commissionBank.Add(payment.CommissionBank)
		commissionAccount = commissionAccount.Add(payment.CommissionAccount)

		if payment.TenantInvoiceID != nil {
			totalTenantPaid = totalTenantPaid.Add(payment.SubTotalTenantSpent)
		} else {
			totalTenantUnPaid = totalTenantUnPaid.Add(payment.SubTotalTenantSpent)
		}
	}

	var (
		giftClaimed, giftUsed     int64
		ticketClaimed, ticketUsed int64
	)

	for _, e := range eventVouchers {
		switch e.Event.Type {
		case domain.EventTypeGift:
			if e.Status == domain.EventVoucherStatusUsed {
				giftUsed++
			}

			if e.UserID != nil {
				giftClaimed++
			}
		case domain.EventTypeTicket:
			if e.Status == domain.EventVoucherStatusUsed {
				ticketUsed++
			}

			if e.UserID != nil {
				ticketClaimed++
			}
		}
	}

	path, err := r.generateReportTenantPaymentTemplate(TenantPaymentInput{
		PeriodFrom:                   from,
		PeriodTo:                     to,
		PaymentCount:                 int64(len(payments)),
		PaymentAmountTotal:           paymentSubTotalSum,
		GeneratedAt:                  generatedAt,
		TransactionCount:             int64(len(payments)),
		TotalSubTotal:                paymentSubTotalSum,
		TotalAmount:                  paymentAmountSum,
		TotalAmountDiscounted:        paymentAmountDiscountedSum,
		TotalUsedPoint:               paymentUsedPointSum,
		TotalAcquiredPoint:           paymentAcquiredPointSum,
		TotalAcquiredPointDiscounted: paymentAcquiredPointDiscountedSum,
		ReturnedAmount:               returnedAmountSum,
		ReturnedAmountDiscounted:     decimal.New(0, 0),
		ReturnedUsedPoint:            returnedUsedPoints,
		VisitorFemaleCount:           int64(len(femaleSpent)),
		VisitorFemaleAmountSpent:     r.getReportPaymentTotalSpent(femaleSpent),
		VisitorMaleCount:             int64(len(maleSpent)),
		VisitorMaleAmountSpent:       r.getReportPaymentTotalSpent(maleSpent),
		GiftClaimedCount:             giftClaimed,
		GiftUsedCount:                giftClaimed,
		TicketClaimedCount:           ticketClaimed,
		TicketUsedCount:              ticketUsed,
		TotalCommissionBank:          commissionBank,
		TotalCommissionAccount:       commissionAccount,
		TenantPaidTotal:              totalTenantPaid,
		TenantUnpaidTotal:            totalTenantUnPaid,
		ParkingTransactionCount:      parkingTransactionsCount,
		ParkingTransactionTotal:      parkingTransactionsAmount,
		Payments:                     paymentList,
	})
	if err != nil {
		return "", err
	}

	return path, nil
}

func (r report) generateReportTenantPaymentTemplate(input TenantPaymentInput) (path string, err error) {
	summarySheetName := "Summary"
	paymentSheetName := "Payments"

	xlsx := excelize.NewFile()
	index := xlsx.NewSheet(summarySheetName)
	xlsx.NewSheet(paymentSheetName)

	mainBackground, err := xlsx.NewStyle(`{"fill":{"type":"pattern","color":["#1EBECE"],"pattern":1}}`)
	if err != nil {
		return "", fmt.Errorf("cannot set main background, err: %v", err)
	}
	xlsx.SetCellStyle(summarySheetName, "A4", "D4", mainBackground)

	grayBackground, err := xlsx.NewStyle(`{"fill":{"type":"pattern","color":["#DADADA"],"pattern":1}}`)
	if err != nil {
		return "", fmt.Errorf("cannot set gray background, err: %v", err)
	}
	xlsx.SetCellStyle(summarySheetName, "A6", "D6", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A8", "D8", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A10", "D10", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A12", "D12", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A14", "D14", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A16", "D16", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A18", "D18", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A20", "D20", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A22", "D22", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A24", "D24", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A26", "D26", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A28", "D28", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A30", "D30", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A32", "D32", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A34", "D34", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A37", "D37", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A38", "D38", grayBackground)
	xlsx.SetCellStyle(summarySheetName, "A39", "D39", grayBackground)

	xlsx.SetColWidth(summarySheetName, "A", "A", 42)
	xlsx.SetColWidth(summarySheetName, "B", "B", 16)
	xlsx.SetColWidth(summarySheetName, "C", "C", 16)
	xlsx.SetColWidth(summarySheetName, "D", "D", 44)

	// Set Column Widths

	xlsx.SetCellStyle(paymentSheetName, "A1", "AB1", grayBackground)

	xlsx.SetColWidth(paymentSheetName, "A", "A", 16)
	xlsx.SetColWidth(paymentSheetName, "B", "B", 26)
	xlsx.SetColWidth(paymentSheetName, "C", "C", 23)
	xlsx.SetColWidth(paymentSheetName, "D", "D", 24)
	xlsx.SetColWidth(paymentSheetName, "E", "E", 18)
	xlsx.SetColWidth(paymentSheetName, "F", "F", 17)
	xlsx.SetColWidth(paymentSheetName, "G", "G", 18)
	xlsx.SetColWidth(paymentSheetName, "H", "H", 18)
	xlsx.SetColWidth(paymentSheetName, "I", "I", 15)
	xlsx.SetColWidth(paymentSheetName, "J", "J", 20)
	xlsx.SetColWidth(paymentSheetName, "K", "K", 20)
	xlsx.SetColWidth(paymentSheetName, "L", "L", 33)
	xlsx.SetColWidth(paymentSheetName, "M", "M", 32)
	xlsx.SetColWidth(paymentSheetName, "N", "N", 15)
	xlsx.SetColWidth(paymentSheetName, "O", "O", 19)
	xlsx.SetColWidth(paymentSheetName, "P", "P", 33.5)
	xlsx.SetColWidth(paymentSheetName, "Q", "Q", 32)
	xlsx.SetColWidth(paymentSheetName, "R", "R", 25)
	xlsx.SetColWidth(paymentSheetName, "S", "S", 37)
	xlsx.SetColWidth(paymentSheetName, "T", "T", 36)
	xlsx.SetColWidth(paymentSheetName, "U", "U", 27)
	xlsx.SetColWidth(paymentSheetName, "V", "V", 28.5)
	xlsx.SetColWidth(paymentSheetName, "W", "W", 23)
	xlsx.SetColWidth(paymentSheetName, "X", "X", 23)
	xlsx.SetColWidth(paymentSheetName, "Y", "Y", 23)
	xlsx.SetColWidth(paymentSheetName, "Z", "Z", 23)
	xlsx.SetColWidth(paymentSheetName, "AA", "AA", 23)
	xlsx.SetColWidth(paymentSheetName, "AB", "AB", 28)

	xlsx.SetCellValue(summarySheetName, "A1", "რეპორტის დათვლის პერიოდი")
	xlsx.SetCellValue(summarySheetName, "A2", fmt.Sprintf("%s - დან %s - მდე", input.PeriodFrom.Format("01/02/2006"), input.PeriodTo.Format("01/02/2006")))

	xlsx.SetCellValue(summarySheetName, "B1", fmt.Sprintf("ტრანზაქციების რაოდენობა: %d", input.PaymentCount))
	xlsx.SetCellValue(summarySheetName, "B2", fmt.Sprintf("ტრანზაქციების მოცულობა: %s", input.PaymentAmountTotal.Round(2).String()))

	xlsx.SetCellValue(summarySheetName, "D1", fmt.Sprintf("გენერაციის თარიღი/დრო: %s", input.GeneratedAt.Format("01/02/2006 15:04:05")))

	xlsx.SetCellValue(summarySheetName, "A3", "")

	xlsx.SetCellValue(summarySheetName, "A4", "REPORT")

	xlsx.SetCellValue(summarySheetName, "A5", "ტრანზაქციების რაოდენობა")
	xlsx.SetCellValue(summarySheetName, "B5", input.TransactionCount)

	xlsx.SetCellValue(summarySheetName, "A6", "ტრანზაქციების ჯამი")
	xlsx.SetCellValue(summarySheetName, "B6", TenantPaymentBeautyDecimalForExcel(input.TotalSubTotal))

	xlsx.SetCellValue(summarySheetName, "A7", "გადახდილი ქულები")
	xlsx.SetCellValue(summarySheetName, "B7", TenantPaymentBeautyDecimalForExcel(input.TotalUsedPoint))

	xlsx.SetCellValue(summarySheetName, "A8", "გადახდილი ლარი (ფასდაუკლებელი)")
	xlsx.SetCellValue(summarySheetName, "B8", TenantPaymentBeautyDecimalForExcel(input.TotalAmount))

	xlsx.SetCellValue(summarySheetName, "A9", "გადახდილი ლარი (ფასდაკლებული)")
	xlsx.SetCellValue(summarySheetName, "B9", TenantPaymentBeautyDecimalForExcel(input.TotalAmountDiscounted))

	xlsx.SetCellValue(summarySheetName, "A10", "დაგროვილი ქულები (ფასდაუკლებელი)")
	xlsx.SetCellValue(summarySheetName, "B10", TenantPaymentBeautyDecimalForExcel(input.TotalAcquiredPoint))

	xlsx.SetCellValue(summarySheetName, "A11", "დაგროვილი ქულები (ფასდაკლებული)")
	xlsx.SetCellValue(summarySheetName, "B11", TenantPaymentBeautyDecimalForExcel(input.TotalAcquiredPointDiscounted))

	xlsx.SetCellValue(summarySheetName, "A12", "")

	xlsx.SetCellValue(summarySheetName, "A13", "დაბრუნებული ქულები")
	xlsx.SetCellValue(summarySheetName, "B13", TenantPaymentBeautyDecimalForExcel(input.ReturnedUsedPoint))

	xlsx.SetCellValue(summarySheetName, "A14", "დაბრუნებული ლარი (ფასდაუკლებელი)")
	xlsx.SetCellValue(summarySheetName, "B14", TenantPaymentBeautyDecimalForExcel(input.ReturnedAmount))

	xlsx.SetCellValue(summarySheetName, "A15", "დაბრუნებული ლარი (ფასდაკლებული)")
	xlsx.SetCellValue(summarySheetName, "B15", TenantPaymentBeautyDecimalForExcel(input.ReturnedAmountDiscounted))

	xlsx.SetCellValue(summarySheetName, "A16", "")

	xlsx.SetCellValue(summarySheetName, "A17", "ბანკის საკომისიო სულ:")
	xlsx.SetCellValue(summarySheetName, "B17", TenantPaymentBeautyDecimalForExcel(input.TotalCommissionBank))

	xlsx.SetCellValue(summarySheetName, "A18", "გალერიას საკომისიო სულ:")
	xlsx.SetCellValue(summarySheetName, "B18", TenantPaymentBeautyDecimalForExcel(input.TotalCommissionAccount))

	xlsx.SetCellValue(summarySheetName, "A19", "")

	xlsx.SetCellValue(summarySheetName, "A20", "ტენანტებთან გადასარიცხი თანხის ჯამი")
	xlsx.SetCellValue(summarySheetName, "B20", TenantPaymentBeautyDecimalForExcel(input.TenantUnpaidTotal))

	xlsx.SetCellValue(summarySheetName, "A21", "ტენანტებთან გადარიცხული თანხის ჯამი")
	xlsx.SetCellValue(summarySheetName, "B21", TenantPaymentBeautyDecimalForExcel(input.TenantPaidTotal))

	xlsx.SetCellValue(summarySheetName, "A22", "")

	xlsx.SetCellValue(summarySheetName, "A23", "პარკინგზე ტრანზაქციები")
	xlsx.SetCellValue(summarySheetName, "B23", input.ParkingTransactionCount)

	xlsx.SetCellValue(summarySheetName, "A24", "პარკინგის თანხა")
	xlsx.SetCellValue(summarySheetName, "B24", TenantPaymentBeautyDecimalForExcel(input.ParkingTransactionTotal))

	xlsx.SetCellValue(summarySheetName, "A25", "")

	xlsx.SetCellValue(summarySheetName, "A26", "ვიზიტორების სტატისტიკა")

	xlsx.SetCellValue(summarySheetName, "A27", "ქალი")
	xlsx.SetCellValue(summarySheetName, "B27", input.VisitorFemaleCount)
	xlsx.SetCellValue(summarySheetName, "A28", "დახარჯული თანხა ***")
	xlsx.SetCellValue(summarySheetName, "B28", TenantPaymentBeautyDecimalForExcel(input.VisitorFemaleAmountSpent))
	xlsx.SetCellValue(summarySheetName, "A29", "მამაკაცი")
	xlsx.SetCellValue(summarySheetName, "B29", input.VisitorMaleCount)
	xlsx.SetCellValue(summarySheetName, "A30", "დახარჯული თანხა ***")
	xlsx.SetCellValue(summarySheetName, "B30", TenantPaymentBeautyDecimalForExcel(input.VisitorMaleAmountSpent))

	xlsx.SetCellValue(summarySheetName, "A31", "")

	xlsx.SetCellValue(summarySheetName, "A32", "ივენთები")

	xlsx.SetCellValue(summarySheetName, "A33", "საჩუქარი აიღო")
	xlsx.SetCellValue(summarySheetName, "B33", input.GiftClaimedCount)

	xlsx.SetCellValue(summarySheetName, "A34", "ბილეთი აიღო")
	xlsx.SetCellValue(summarySheetName, "B34", input.TicketClaimedCount)

	xlsx.SetCellValue(summarySheetName, "A35", "გამოიყენა საჩუქარი")
	xlsx.SetCellValue(summarySheetName, "B35", input.GiftUsedCount)

	xlsx.SetCellValue(summarySheetName, "A36", "გამოიყენა ბილეთი")
	xlsx.SetCellValue(summarySheetName, "B36", input.TicketUsedCount)

	xlsx.SetCellValue(summarySheetName, "A37", "")
	xlsx.SetCellValue(summarySheetName, "A38", "ტოპ 10 ტენანტი")
	xlsx.SetCellValue(summarySheetName, "A39", "")

	// Payments Sheet
	xlsx.SetCellValue(paymentSheetName, "A1", "თარიღი")
	xlsx.SetCellValue(paymentSheetName, "B1", "ტენანტი")
	xlsx.SetCellValue(paymentSheetName, "C1", "ტრანზაქციის ID")
	xlsx.SetCellValue(paymentSheetName, "D1", "თიბისის RRN")
	xlsx.SetCellValue(paymentSheetName, "E1", "თიბისის ტრანზაქციის ID")
	xlsx.SetCellValue(paymentSheetName, "F1", "ტრანზაქციის ტიპი")
	xlsx.SetCellValue(paymentSheetName, "G1", "გადახდის მეთოდი")
	xlsx.SetCellValue(paymentSheetName, "H1", "ბარათის ტიპი")
	xlsx.SetCellValue(paymentSheetName, "I1", "CARD PAN")
	xlsx.SetCellValue(paymentSheetName, "J1", "სტატუსი")
	xlsx.SetCellValue(paymentSheetName, "K1", "სულ ჯამი")
	xlsx.SetCellValue(paymentSheetName, "L1", "გადახდილი ქულები")
	xlsx.SetCellValue(paymentSheetName, "M1", "გადახდილი თანხა (ფასდაუკლებელი)")
	xlsx.SetCellValue(paymentSheetName, "N1", "გადახდილი თანხა (ფასდაკლებული)")
	xlsx.SetCellValue(paymentSheetName, "O1", "ბანკის საკომისიო")
	xlsx.SetCellValue(paymentSheetName, "P1", "გალერიას საკომისიო")
	xlsx.SetCellValue(paymentSheetName, "Q1", "ქეშბექის პროცენტი (ფასდაუკლებელი)")
	xlsx.SetCellValue(paymentSheetName, "R1", "ქეშბექის პროცენტი (ფასდაკლებული)")
	xlsx.SetCellValue(paymentSheetName, "S1", "ფასდაუკლებელი საქონელი")
	xlsx.SetCellValue(paymentSheetName, "T1", "დაგროვილი ქულები (ფასდაუკლებელზე)")
	xlsx.SetCellValue(paymentSheetName, "U1", "დაგროვილი ქულები (ფასდაკლებულზე)")
	xlsx.SetCellValue(paymentSheetName, "V1", "დაგროვილი ქულები (სტატუსი)")
	xlsx.SetCellValue(paymentSheetName, "W1", "ტენანტზე დასაბრუნებელი თანხა")
	xlsx.SetCellValue(paymentSheetName, "X1", "სტატუსი (ინვოისი)")
	xlsx.SetCellValue(paymentSheetName, "Y1", "გადახდის თარიღი")
	xlsx.SetCellValue(paymentSheetName, "Z1", "დაბრუნების თარიღი")
	xlsx.SetCellValue(paymentSheetName, "AA1", "დაბრუნების ვადა")
	xlsx.SetCellValue(paymentSheetName, "AB1", "მთავარი ტრანზაქციის ID")
	xlsx.SetCellValue(paymentSheetName, "AC1", "დაბრუნებული ტრანზაქციის ID")

	paymentRowIndex := 2
	for _, payment := range input.Payments {
		paymentMethod := payment.PaymentMethod
		if payment.PaymentMethod == "" {
			paymentMethod = "Points"
		}

		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("A%d", paymentRowIndex), payment.CreatedAt)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("B%d", paymentRowIndex), payment.TenantName)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("C%d", paymentRowIndex), payment.TransactionID)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("D%d", paymentRowIndex), payment.SourceRRN)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("E%d", paymentRowIndex), payment.SourceTransactionID)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("F%d", paymentRowIndex), payment.TransactionType)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("G%d", paymentRowIndex), paymentMethod)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("H%d", paymentRowIndex), payment.BankCardType)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("I%d", paymentRowIndex), payment.BankCardPan)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("J%d", paymentRowIndex), payment.Status)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("K%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.Total))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("L%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.UsedPoint))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("M%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.Amount))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("N%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.AmountDiscounted))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("O%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.CommissionBank))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("P%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.CommissionAccount))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("Q%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.CashbackPercent))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("R%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.CashbackDiscountedPercent))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("S%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.Total.Sub(payment.AmountDiscounted)))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("T%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.AcquiredPoint))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("U%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.AcquiredPointDiscounted))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("V%d", paymentRowIndex), payment.AcquiredPointStatus)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("W%d", paymentRowIndex), TenantPaymentBeautyDecimalForExcel(payment.TenantSpentAmount))
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("X%d", paymentRowIndex), payment.InvoiceStatus)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("Y%d", paymentRowIndex), payment.PaidAt)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("Z%d", paymentRowIndex), payment.ReturnExpiredAt)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("AA%d", paymentRowIndex), payment.ReturnExpiredDuration)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("AB%d", paymentRowIndex), payment.MainTransactionID)
		xlsx.SetCellValue(paymentSheetName, fmt.Sprintf("AC%d", paymentRowIndex), payment.ReturnTransactionID)

		paymentRowIndex++
	}

	xlsx.SetActiveSheet(index)
	xlsx.DeleteSheet("Sheet1")

	randomString := utils.GenerateRandomString(5)

	xlsxFilePath := fmt.Sprintf("%s/loyalty-payments-report-%s-%s-%s.xlsx", config.GetDirectoryPath("tmp"), input.PeriodFrom.Format("2006-02-01"), input.PeriodTo.Format("2006-02-01"), randomString)
	//zippedFilePath := fmt.Sprintf("%s/loyalty-payments-report-%s-%s-%s.zip", config.GetDirectoryPath("tmp"), input.PeriodFrom.Format("2006.02.01"), input.PeriodTo.Format("2006.02.01"), randomString)

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

func TenantPaymentBeautyDecimalForExcel(val decimal.Decimal) string {
	return strings.ReplaceAll(val.Round(2).String(), ".", ",")
}

func (r report) getReportTenantPaymentTotalSpent(genderSpent map[uuid.UUID]decimal.Decimal) decimal.Decimal {
	var total decimal.Decimal

	for _, amount := range genderSpent {
		total = total.Add(amount)
	}

	return total
}

func (r report) getReportTenantPaymentTopTenantsByAmount(payments map[uuid.UUID]decimal.Decimal, names map[uuid.UUID]string, n int) []TenantItem {
	tenants := make([]TenantItem, 0, len(payments))

	for tenantID, amount := range payments {
		tenants = append(tenants, TenantItem{
			Name:  names[tenantID],
			Value: amount,
		})
	}

	sort.Slice(tenants, func(i, j int) bool {
		return tenants[i].Value.GreaterThan(tenants[j].Value)
	})

	if len(tenants) < n {
		n = len(tenants)
	}

	return tenants[:n]
}
