package controllers

import (
	"bytes"
	"errors"
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

func SellerOverAllSalesReport(c *gin.Context) {
	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "seller not authorized",
		})
		return
	}

	sellerIDStr, ok := sellerID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve seller information",
		})
		return
	}

	var input models.SellerOverallSalesReport

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.StartDate == "" && input.EndDate == "" && input.Limit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide start date and end date, or specify the limit as day, week, month, year"})
		return
	}

	if input.Limit != "" {
		limits := []string{"day", "week", "month", "year"}
		found := false
		for _, l := range limits {
			if input.Limit == l {
				found = true
				break
			}
		}

		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit specified, valid options are: day, week, month, year"})
			return
		}

		var startDate, endDate string
		switch input.Limit {
		case "day":
			startDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		case "week":
			today := time.Now()
			startDate = today.AddDate(0, 0, -int(today.Weekday())).Format("2006-01-02")
			endDate = today.AddDate(0, 0, 7-int(today.Weekday())).Format("2006-01-02")
		case "month":
			today := time.Now()
			firstDayOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
			lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)
			startDate = firstDayOfMonth.Format("2006-01-02")
			endDate = lastDayOfMonth.Format("2006-01-02")
		case "year":
			startDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		}

		result, amount, err := TotalOrders(startDate, endDate, input.PaymentStatus, sellerIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "successfully created sales report",
			"result":  result,
			"amount":  amount,
		})
		return
	}

	if input.StartDate != "" && input.EndDate != "" {
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date provided"})
			return
		}

		endDate, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date provided"})
			return
		}

		if startDate.After(endDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start date cannot be after end date"})
			return
		}
	}

	result, amount, err := TotalOrders(input.StartDate, input.EndDate, input.PaymentStatus, sellerIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully created sales report",
		"result":  result,
		"amount":  amount,
	})
}

func TotalOrders(From string, Till string, PaymentStatus string, SellerID uint) (models.OrderCount, models.AmountInformation, error) {
	var orders []models.Order

	parsedFrom, err := time.Parse("2006-01-02", From)
	if err != nil {
		return models.OrderCount{}, models.AmountInformation{}, fmt.Errorf("error parsing From time: %v", err)
	}

	parsedTill, err := time.Parse("2006-01-02", Till)
	if err != nil {
		return models.OrderCount{}, models.AmountInformation{}, fmt.Errorf("error parsing Till time: %v", err)
	}

	fFrom := time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, time.UTC)
	fTill := time.Date(parsedTill.Year(), parsedTill.Month(), parsedTill.Day(), 23, 59, 59, 999999999, time.UTC)

	startTime := fFrom.Format("2006-01-02T15:04:05Z")
	endDate := fTill.Format("2006-01-02T15:04:05Z")

	if SellerID != 0 {
		if err := database.DB.Where("ordered_at BETWEEN ? AND ? AND payment_status = ? AND seller_id = ?", startTime, endDate, PaymentStatus, SellerID).Find(&orders).Error; err != nil {
			return models.OrderCount{}, models.AmountInformation{}, errors.New("error fetching orders")
		}
	} else {
		if err := database.DB.Where("ordered_at BETWEEN ? AND ? AND payment_status = ?", startTime, endDate, PaymentStatus).Find(&orders).Error; err != nil {
			return models.OrderCount{}, models.AmountInformation{}, errors.New("error fetching orders")
		}
	}

	var orderStatusCounts = map[string]int64{
		models.OrderStatusPending:   0,
		models.OrderStatusConfirmed: 0,
		models.OrderStatusShipped:   0,
		models.OrderStatusDelivered: 0,
		models.OrderStatusCanceled:  0,
		models.OrderStatusReturned:  0,
	}

	var AccountInformation models.AmountInformation

	for _, order := range orders {
		AccountInformation.TotalCouponDeduction += RoundDecimalValue(order.CouponDiscountAmount)
		AccountInformation.TotalReferralDeduction += RoundDecimalValue(order.ReferralDiscountAmount)
		//	AccountInformation.TotalProductOfferDeduction += RoundDecimalValue(order.)
		AccountInformation.TotalAmountBeforeDeduction += RoundDecimalValue(order.TotalAmount)
		AccountInformation.TotalAmountAfterDeduction += RoundDecimalValue(order.FinalAmount)

		for _, status := range []string{
			models.OrderStatusPending,
			models.OrderStatusConfirmed,
			models.OrderStatusShipped,
			models.OrderStatusDelivered,
			models.OrderStatusCanceled,
			models.OrderStatusReturned,
		} {
			var count int64
			if err := database.DB.Model(&models.OrderItem{}).Where("order_id = ? AND status = ?", order.OrderID, status).Count(&count).Error; err != nil {
				return models.OrderCount{}, models.AmountInformation{}, errors.New("failed to query order items")
			}
			orderStatusCounts[status] += count
		}
	}

	var totalCount int64
	for _, count := range orderStatusCounts {
		totalCount += count
	}

	return models.OrderCount{
		TotalOrder:     uint(totalCount),
		TotalPending:   uint(orderStatusCounts[models.OrderStatusPending]),
		TotalConfirmed: uint(orderStatusCounts[models.OrderStatusConfirmed]),
		TotalShipped:   uint(orderStatusCounts[models.OrderStatusShipped]),
		TotalDelivered: uint(orderStatusCounts[models.OrderStatusDelivered]),
		TotalCancelled: uint(orderStatusCounts[models.OrderStatusCanceled]),
		TotalReturned:  uint(orderStatusCounts[models.OrderStatusReturned]),
	}, AccountInformation, nil
}

func GenerateSalesReportPDF(orderCount models.OrderCount, amountInfo models.AmountInformation) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Sales Report")

	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Total Orders: "+strconv.Itoa(int(orderCount.TotalOrder)))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Total Amount Before Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalAmountBeforeDeduction))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Total Coupon Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalCouponDeduction))
	pdf.Ln(10)
	// pdf.Cell(40, 10, "Total Product Offer Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalProductOfferDeduction))
	// pdf.Ln(10)
	pdf.Cell(40, 10, "Total Referral Offer Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalReferralDeduction))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Total Amount After Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalAmountAfterDeduction))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Order Status Summary")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Total Pending Orders: "+strconv.Itoa(int(orderCount.TotalPending)))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Total Confirmed Orders: "+strconv.Itoa(int(orderCount.TotalConfirmed)))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Total Shipped Orders: "+strconv.Itoa(int(orderCount.TotalShipped)))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Total Delivered Orders: "+strconv.Itoa(int(orderCount.TotalDelivered)))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Total Cancelled Orders: "+strconv.Itoa(int(orderCount.TotalCancelled)))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Total Returned Orders: "+strconv.Itoa(int(orderCount.TotalReturned)))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DownloadSalesReportPDF(c *gin.Context) {

	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "seller not authorized"})
		return
	}
	sellerIDStr := sellerID.(uint)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	paymentStatus := c.Query("payment_status")

	orderCount, amountInfo, err := TotalOrders(startDate, endDate, paymentStatus, sellerIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	pdfBytes, err := GenerateSalesReportPDF(orderCount, amountInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=sales_report.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func DownloadSalesReportExcel(c *gin.Context) {

	sellerID, exists := c.Get("sellerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "seller not authorized"})
		return
	}
	sellerIDStr := sellerID.(uint)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	paymentStatus := c.Query("payment_status")

	orderCount, amountInfo, err := TotalOrders(startDate, endDate, paymentStatus, sellerIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	excelBytes, err := GenerateSalesReportExcel(orderCount, amountInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate Excel report"})
		return
	}

	// Set the response headers for Excel download
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=sales_report.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
}

func GenerateSalesReportExcel(orderCount models.OrderCount, amountInfo models.AmountInformation) ([]byte, error) {
	f := excelize.NewFile()

	// Create headers
	f.SetCellValue("Sheet1", "A1", "Sales Report")
	f.SetCellValue("Sheet1", "A3", "Total Orders")
	f.SetCellValue("Sheet1", "B3", strconv.Itoa(int(orderCount.TotalOrder)))

	// Add Amount Information
	f.SetCellValue("Sheet1", "A4", "Total Amount Before Deduction")
	f.SetCellValue("Sheet1", "B4", fmt.Sprintf("%.2f", amountInfo.TotalAmountBeforeDeduction))
	f.SetCellValue("Sheet1", "A5", "Total Coupon Deduction")
	f.SetCellValue("Sheet1", "B5", fmt.Sprintf("%.2f", amountInfo.TotalCouponDeduction))
	f.SetCellValue("Sheet1", "A6", "Total Referral Offer Deduction")
	f.SetCellValue("Sheet1", "B6", fmt.Sprintf("%.2f", amountInfo.TotalReferralDeduction))
	f.SetCellValue("Sheet1", "A7", "Total Amount After Deduction")
	f.SetCellValue("Sheet1", "B7", fmt.Sprintf("%.2f", amountInfo.TotalAmountAfterDeduction))

	// Add Order Status Summary
	f.SetCellValue("Sheet1", "A9", "Order Status Summary")
	f.SetCellValue("Sheet1", "A10", "Total Pending Orders")
	f.SetCellValue("Sheet1", "B10", strconv.Itoa(int(orderCount.TotalPending)))
	f.SetCellValue("Sheet1", "A11", "Total Confirmed Orders")
	f.SetCellValue("Sheet1", "B11", strconv.Itoa(int(orderCount.TotalConfirmed)))
	f.SetCellValue("Sheet1", "A12", "Total Shipped Orders")
	f.SetCellValue("Sheet1", "B12", strconv.Itoa(int(orderCount.TotalShipped)))
	f.SetCellValue("Sheet1", "A13", "Total Delivered Orders")
	f.SetCellValue("Sheet1", "B13", strconv.Itoa(int(orderCount.TotalDelivered)))
	f.SetCellValue("Sheet1", "A14", "Total Cancelled Orders")
	f.SetCellValue("Sheet1", "B14", strconv.Itoa(int(orderCount.TotalCancelled)))
	f.SetCellValue("Sheet1", "A15", "Total Returned Orders")
	f.SetCellValue("Sheet1", "B15", strconv.Itoa(int(orderCount.TotalReturned)))

	// Write the Excel file to a buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
