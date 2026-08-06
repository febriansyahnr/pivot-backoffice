package unifiedPaymentService

import (
	"bytes"
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"

	"github.com/xuri/excelize/v2"
)

func (s *UnifiedPaymentService) ExportToExcelChargeHistories(ctx context.Context, request *unifiedPaymentModel.FilterChargeRequest, charges []unifiedPaymentModel.ChargeResponse) (*bytes.Buffer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/ExportToExcelChargeHistories")
	defer segment.End()

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sw, err := f.NewStreamWriter(constant.DefaultSheetName)
	if err != nil {
		return nil, err
	}
	loc := util.GetTimeLocationFromContext(ctx)

	startDate := request.StartCreatedAt.In(loc)
	endDate := request.EndCreatedAt.In(loc)

	_ = sw.MergeCell("A2", "C2")
	_ = sw.MergeCell("A3", "C3")
	// Set column widths for all 22 columns
	for i := 1; i <= 22; i++ {
		_ = sw.SetColWidth(i, i, 18)
	}

	styleSubTitle := xlsx.StyleSubTitle(f)
	styleHeaderColumn := xlsx.StyleHeaderColumn(f)
	styleValDatetime12Hour := xlsx.StyleDatetime12HoursRow(f)
	styleValString := xlsx.StyleNormalRow(f)
	styleValCurrency := xlsx.StyleCurrencyRowWithCustomFormat(f, xlsx.CustomNormalCurrencyFmt)

	_ = sw.SetRow("A2", []any{excelize.Cell{StyleID: xlsx.StyleTitleWithAlignment(f), Value: "Charge History"}}, excelize.RowOpts{
		Height: 20,
	})
	_ = sw.SetRow("A3", []any{
		excelize.Cell{StyleID: styleSubTitle, Value: util.DateStrMonthYear(startDate) + " - " + util.DateStrMonthYear(endDate)},
	}, xlsx.DefaultRowOpt())

	rows := 3
	if request.Status != "" {
		rows++
		_ = sw.SetRow(fmt.Sprintf("A%d", rows), []any{
			excelize.Cell{StyleID: styleSubTitle, Value: "Charge Status:"},
			excelize.Cell{StyleID: styleValString, Value: util.ToTitle(request.Status)},
		}, xlsx.DefaultRowOpt())
	}
	if request.UUID != "" {
		rows++
		_ = sw.SetRow(fmt.Sprintf("A%d", rows), []any{
			excelize.Cell{StyleID: styleSubTitle, Value: "Charge Id:"},
			excelize.Cell{StyleID: styleValString, Value: request.UUID},
		}, xlsx.DefaultRowOpt())
	}

	headerNames := []string{
		"Created Date", "Method", "Channel", "Payment ID", "Reference ID", "Amount", "Charge ID", "Status", "Failure Reason", "Payment Date",
		"Bank Reference ID", "Expiry Time", "Total Authorized Amount", "Total Captured Amount", "Statement Descriptor",
		"Network Response Code", "Acquiring Bank", "Bank Merchant ID (MID)", "Issuer Bank", "Card Number", "VA Name", "VA Number", "Merchant Name",
	}

	headers := make([]any, len(headerNames))
	for i := range headerNames {
		headers[i] = excelize.Cell{StyleID: styleHeaderColumn, Value: headerNames[i]}
	}
	rows += 2
	_ = sw.SetRow(fmt.Sprintf("A%d", rows), headers)

	for i, charge := range charges {

		// Derive method and channel based on payment method type
		method := ""
		paymentMethodType := ""
		channel := ""
		bankReferenceID := ""
		expiryTime := ""
		networkResponseCode := ""
		issuerBank := ""
		cardNumber := ""
		vaName := ""
		vaNumber := ""
		merchantName := charge.MerchantName

		// Determine method and channel
		if charge.Card != nil {
			method = "Card"
			paymentMethodType = paymentConstant.PAYMENT_METHOD_CREDIT_CARD
			channel = charge.Card.BinInformations.Brand
			if charge.Card.AuthorizationResult != nil {
				bankReferenceID = charge.Card.AuthorizationResult.AcquirerReferenceNumber
				networkResponseCode = charge.Card.AuthorizationResult.IssuerAuthorizationCode
			}
			issuerBank = charge.Card.BinInformations.IssuingBank
			cardNumber = fmt.Sprintf("%s****%s", charge.Card.First6, charge.Card.Last4)
			expiryTime = charge.ExpiredAt.In(loc).Format("2006-01-02 15:04:05")
		} else if charge.VirtualAccount != nil {
			method = "Virtual Account"
			paymentMethodType = paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT
			channel = charge.VirtualAccount.Channel
			expiryTime = charge.VirtualAccount.ExpiryAt.In(loc).Format("2006-01-02 15:04:05")
			vaName = charge.VirtualAccount.VirtualAccountName
			vaNumber = charge.VirtualAccount.VirtualAccountNumber
			bankReferenceID = charge.VirtualAccount.BankReferenceNo
		} else if charge.Ewallet != nil {
			method = "E-WALLET"
			paymentMethodType = paymentConstant.PAYMENT_METHOD_EWALLET
			channel = charge.Ewallet.Channel
			bankReferenceID = charge.Ewallet.ReferenceNo
		} else if charge.Qr != nil {
			method = "QRIS"
			paymentMethodType = paymentConstant.PAYMENT_METHOD_QRIS
			channel = charge.Qr.Acquirer
			bankReferenceID = charge.Qr.RetrievalReferenceNumber
			expiryTime = charge.Qr.ExpiryAt.In(loc).Format("2006-01-02 15:04:05")
			if merchantName == "" {
				merchantName = charge.StatementDescriptor
			}
		}

		// Acquiring Bank and MID - only show for Direct PSP (MIDType == "DIRECT")
		acquiringBank := ""
		mid := ""
		if charge.Card != nil && charge.Card.MIDInfo != nil && charge.Card.MIDInfo.Type == constant.PaymentMethodChannelTypeDirect {
			acquiringBank = charge.Card.MIDInfo.Acquirer
			mid = charge.Card.MIDInfo.MID
		}

		// Format payment date
		paymentDate := ""
		if charge.PaidAt != nil {
			paymentDate = charge.PaidAt.In(loc).Format("2006-01-02 15:04:05")
		}

		_ = sw.SetRow(
			fmt.Sprintf("A%d", (rows+1)+i), []any{
				excelize.Cell{StyleID: styleValDatetime12Hour, Value: charge.CreatedAt.In(loc)},
				excelize.Cell{StyleID: styleValString, Value: method},
				excelize.Cell{StyleID: styleValString, Value: channel},
				excelize.Cell{StyleID: styleValString, Value: charge.PaymentSessionID},
				excelize.Cell{StyleID: styleValString, Value: charge.PaymentSessionClientReferenceID},
				excelize.Cell{StyleID: styleValCurrency, Value: charge.Amount.Value},
				excelize.Cell{StyleID: styleValString, Value: charge.ID},
				excelize.Cell{StyleID: styleValString, Value: util.ToTitle(charge.Status)},
				excelize.Cell{StyleID: styleValString, Value: charge.ChargePaymentMethodDetails.GetNaturalPaymentFailureMessage(paymentMethodType, charge.FailureCode)},
				excelize.Cell{StyleID: styleValString, Value: paymentDate},
				excelize.Cell{StyleID: styleValString, Value: bankReferenceID},
				excelize.Cell{StyleID: styleValString, Value: expiryTime},
				excelize.Cell{StyleID: styleValCurrency, Value: util.ValueOfPtr(charge.AuthorizedAmount).Value},
				excelize.Cell{StyleID: styleValCurrency, Value: util.ValueOfPtr(charge.CapturedAmount).Value},
				excelize.Cell{StyleID: styleValString, Value: charge.StatementDescriptor},
				excelize.Cell{StyleID: styleValString, Value: networkResponseCode},
				excelize.Cell{StyleID: styleValString, Value: acquiringBank},
				excelize.Cell{StyleID: styleValString, Value: mid},
				excelize.Cell{StyleID: styleValString, Value: issuerBank},
				excelize.Cell{StyleID: styleValString, Value: cardNumber},
				excelize.Cell{StyleID: styleValString, Value: vaName},
				excelize.Cell{StyleID: styleValString, Value: vaNumber},
				excelize.Cell{StyleID: styleValString, Value: merchantName},
			},
		)
	}
	if err := sw.Flush(); err != nil {
		return nil, err
	}
	return f.WriteToBuffer()
}
