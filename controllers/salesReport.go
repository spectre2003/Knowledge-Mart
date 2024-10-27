package controllers

import (
	"bytes"
	"errors"
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"log"
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
		AccountInformation.TotalAmountBeforeDeduction += RoundDecimalValue(order.TotalAmount)
		AccountInformation.TotalCouponDeduction += RoundDecimalValue(order.CouponDiscountAmount)
		AccountInformation.TotalProuctOfferDeduction += RoundDecimalValue(order.ProductOfferAmount)
		AccountInformation.TotalDeliveryCharges += RoundDecimalValue(order.DeliveryCharge)
		AccountInformation.TotalCategoryOfferDeduction += RoundDecimalValue(order.CategoryDiscountAmount)
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
	limit := c.Query("limit")

	if limit == "" && (startDate == "" || endDate == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide either limit or start_date and end_date"})
		return
	}

	if limit != "" {
		limits := []string{"day", "week", "month", "year"}
		found := false
		for _, l := range limits {
			if limit == l {
				found = true
				break
			}
		}

		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit specified, valid options are: day, week, month, year"})
			return
		}

		switch limit {
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
	}

	orderCount, amountInfo, err := TotalOrders(startDate, endDate, paymentStatus, sellerIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	pdfBytes, err := GenerateSalesReportPDF(orderCount, amountInfo, startDate, endDate, paymentStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=sales_report.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func GenerateSalesReportPDF(orderCount models.OrderCount, amountInfo models.AmountInformation, startDate string, endDate string, paymentStatus string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Sales Report")

	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Start Date: "+startDate)
	pdf.Ln(8)
	pdf.Cell(40, 10, "End Date: "+endDate)
	pdf.Ln(8)
	pdf.Cell(40, 10, "Payment Status: "+paymentStatus)
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Total Orders: "+strconv.Itoa(int(orderCount.TotalOrder)))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total Amount Before Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalAmountBeforeDeduction))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total Coupon Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalCouponDeduction))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total Category Offer Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalCategoryOfferDeduction))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total product Offer Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalProuctOfferDeduction))
	pdf.Ln(8)
	pdf.Cell(40, 10, "delivery charges: "+fmt.Sprintf("%.2f", amountInfo.TotalDeliveryCharges))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total Amount After Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalAmountAfterDeduction))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Order Status Summary")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(60, 10, "Order Status", "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 10, "Total Count", "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 12)

	orderStatuses := map[string]uint{
		"Pending Orders":   orderCount.TotalPending,
		"Confirmed Orders": orderCount.TotalConfirmed,
		"Shipped Orders":   orderCount.TotalShipped,
		"Delivered Orders": orderCount.TotalDelivered,
		"Cancelled Orders": orderCount.TotalCancelled,
		"Returned Orders":  orderCount.TotalReturned,
	}

	for status, count := range orderStatuses {
		pdf.CellFormat(60, 10, status, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 10, strconv.Itoa(int(count)), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
	limit := c.Query("limit")

	if limit == "" && (startDate == "" || endDate == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide either limit or start_date and end_date"})
		return
	}

	if limit != "" {
		limits := []string{"day", "week", "month", "year"}
		found := false
		for _, l := range limits {
			if limit == l {
				found = true
				break
			}
		}

		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit specified, valid options are: day, week, month, year"})
			return
		}

		switch limit {
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
	}

	orderCount, amountInfo, err := TotalOrders(startDate, endDate, paymentStatus, sellerIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	excelBytes, err := GenerateSalesReportExcel(orderCount, amountInfo, startDate, endDate, paymentStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate Excel report"})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=sales_report.xlsx")
	c.Header("Content-Length", strconv.Itoa(len(excelBytes))) // Ensure Content-Length is added
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
}

func GenerateSalesReportExcel(orderCount models.OrderCount, amountInfo models.AmountInformation, startDate, endDate, paymentStatus string) ([]byte, error) {
	f := excelize.NewFile()

	f.SetCellValue("Sheet1", "A1", "Sales Report")

	f.SetCellValue("Sheet1", "A2", "Start Date")
	f.SetCellValue("Sheet1", "B2", startDate)
	f.SetCellValue("Sheet1", "A3", "End Date")
	f.SetCellValue("Sheet1", "B3", endDate)
	f.SetCellValue("Sheet1", "A4", "Payment Status")
	f.SetCellValue("Sheet1", "B4", paymentStatus)

	f.SetCellValue("Sheet1", "A6", "Total Orders")
	f.SetCellValue("Sheet1", "B6", strconv.Itoa(int(orderCount.TotalOrder)))

	f.SetCellValue("Sheet1", "A7", "Total Amount Before Deduction")
	f.SetCellValue("Sheet1", "B7", fmt.Sprintf("%.2f", amountInfo.TotalAmountBeforeDeduction))

	f.SetCellValue("Sheet1", "A8", "Total Coupon Deduction")
	f.SetCellValue("Sheet1", "B8", fmt.Sprintf("%.2f", amountInfo.TotalCouponDeduction))

	f.SetCellValue("Sheet1", "A9", "Total Category Offer Deduction")
	f.SetCellValue("Sheet1", "B9", fmt.Sprintf("%.2f", amountInfo.TotalCategoryOfferDeduction))

	f.SetCellValue("Sheet1", "A10", "Total Product Offer Deduction")
	f.SetCellValue("Sheet1", "B10", fmt.Sprintf("%.2f", amountInfo.TotalProuctOfferDeduction))

	f.SetCellValue("Sheet1", "A10", "Total Delivery charges")
	f.SetCellValue("Sheet1", "B10", fmt.Sprintf("%.2f", amountInfo.TotalDeliveryCharges))

	f.SetCellValue("Sheet1", "A11", "Total Amount After Deduction")
	f.SetCellValue("Sheet1", "B11", fmt.Sprintf("%.2f", amountInfo.TotalAmountAfterDeduction))

	// Adding Order Status Summary
	f.SetCellValue("Sheet1", "A13", "Order Status Summary")
	f.SetCellValue("Sheet1", "A14", "Total Pending Orders")
	f.SetCellValue("Sheet1", "B14", strconv.Itoa(int(orderCount.TotalPending)))

	f.SetCellValue("Sheet1", "A15", "Total Confirmed Orders")
	f.SetCellValue("Sheet1", "B15", strconv.Itoa(int(orderCount.TotalConfirmed)))

	f.SetCellValue("Sheet1", "A16", "Total Shipped Orders")
	f.SetCellValue("Sheet1", "B16", strconv.Itoa(int(orderCount.TotalShipped)))

	f.SetCellValue("Sheet1", "A17", "Total Delivered Orders")
	f.SetCellValue("Sheet1", "B17", strconv.Itoa(int(orderCount.TotalDelivered)))

	f.SetCellValue("Sheet1", "A18", "Total Cancelled Orders")
	f.SetCellValue("Sheet1", "B18", strconv.Itoa(int(orderCount.TotalCancelled)))

	f.SetCellValue("Sheet1", "A19", "Total Returned Orders")
	f.SetCellValue("Sheet1", "B19", strconv.Itoa(int(orderCount.TotalReturned)))

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func OrderInvoice(c *gin.Context) {
	orderID := c.Query("order_id")

	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Order ID is required",
		})
		return
	}

	var order models.Order
	if err := database.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "Failed to fetch order information",
		})
		return
	}

	var seller models.Seller
	if err := database.DB.Where("id = ?", order.SellerID).First(&seller).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "Failed to fetch seller information",
		})
		return
	}

	if order.PaymentStatus != models.PaymentStatusPaid {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "only paid order invoice will get",
		})
		return
	}

	var orderItems []models.OrderItem
	if err := database.DB.Preload("Product").Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "Failed to fetch order item information",
		})
		return
	}

	var payment models.Payment
	if err := database.DB.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "Failed to fetch payment information",
		})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", order.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "Failed to fetch user information",
		})
		return
	}

	billingAddress := order.ShippingAddress

	pdfBytes, err := GeneratePDFInvoice(order, orderItems, user, billingAddress, payment, seller.UserName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to generate PDF",
			"error":   err.Error(),
		})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/pdf")
	c.Writer.Header().Set("Content-Disposition", "inline; filename=invoice.pdf")
	c.Writer.Write(pdfBytes)
}

func GeneratePDFInvoice(order models.Order, orderItems []models.OrderItem, user models.User, billingAddress models.ShippingAddress, payment models.Payment, sellerName string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetMargins(10, 10, 20) // Added right margin for more space
	pdf.SetFont("Arial", "B", 16)

	pdf.CellFormat(0, 10, "Tax Invoice", "", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(95, 6, fmt.Sprintf("Sold By: %s", sellerName), "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Contact: support@knowledgemart.com", "", 0, "R", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(95, 6, "Ship-from Address: Sy no 18/2, Mananthavady, Kerala", "", 0, "L", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 6, "Ship To:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 6, user.Name)
	pdf.Ln(6)
	printAddress(pdf, billingAddress)
	pdf.Ln(10)

	pdf.CellFormat(95, 6, fmt.Sprintf("Order ID: %d", order.OrderID), "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Order Date: %s", order.OrderedAt.Format("02-Jan-2006")), "", 0, "R", false, 0, "")
	pdf.Ln(6)
	pdf.CellFormat(95, 6, fmt.Sprintf("Payment ID: %s", payment.RazorpayPaymentID), "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Payment Method: %s", order.PaymentMethod), "", 0, "R", false, 0, "")
	pdf.Ln(12)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	tableHeaders := []string{"Product", "Description", "Price", "Product Offer", "Category Offer", "Other Offers", "Final Amount"}
	headerWidths := []float64{25, 60, 20, 24, 26, 22, 23}

	for i, header := range tableHeaders {
		pdf.CellFormat(headerWidths[i], 10, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(10)

	// Table content
	pdf.SetFont("Arial", "", 10)
	var totalFinalAmount, totalProductOffer, totalCategoryOffer, totalOtherOffers, totalPrice float64
	for _, item := range orderItems {
		pdf.CellFormat(25, 10, item.Product.Name, "1", 0, "C", false, 0, "")

		// x, y := pdf.GetXY()
		// pdf.MultiCell(50, 10, item.Product.Description, "1", "L", false)
		// pdf.SetXY(x+50, y)
		pdf.CellFormat(60, 10, item.Product.Description, "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 10, fmt.Sprintf("%.2f", item.Price), "1", 0, "C", false, 0, "")
		pdf.CellFormat(24, 10, fmt.Sprintf("%.2f", item.ProductOfferAmount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(26, 10, fmt.Sprintf("%.2f", item.CategoryOfferAmount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(22, 10, fmt.Sprintf("%.2f", item.OtherOffers), "1", 0, "C", false, 0, "")
		pdf.CellFormat(23, 10, fmt.Sprintf("%.2f", item.FinalAmount), "1", 0, "C", false, 0, "")
		pdf.Ln(10)

		// Accumulate totals for final row
		totalFinalAmount += item.FinalAmount
		totalProductOffer += item.ProductOfferAmount
		totalCategoryOffer += item.CategoryOfferAmount
		totalOtherOffers += item.OtherOffers
		totalPrice += item.Price
	}

	// Add space before the Total row
	pdf.Ln(5)

	// Total row with padding and centered alignment
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(85, 10, "Total", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 10, fmt.Sprintf("%.2f", totalPrice), "1", 0, "C", false, 0, "")
	pdf.CellFormat(24, 10, fmt.Sprintf("%.2f", totalProductOffer), "1", 0, "C", false, 0, "")
	pdf.CellFormat(26, 10, fmt.Sprintf("%.2f", totalCategoryOffer), "1", 0, "C", false, 0, "")
	pdf.CellFormat(22, 10, fmt.Sprintf("%.2f", totalOtherOffers), "1", 0, "C", false, 0, "")
	pdf.CellFormat(23, 10, fmt.Sprintf("%.2f", totalFinalAmount), "1", 0, "C", false, 0, "")
	pdf.Ln(12)

	// Thank you note, centered and with more space below the table
	pdf.SetFont("Arial", "B", 12)
	pdf.Ln(5) // Additional space from the total row
	pdf.CellFormat(0, 10, "Thank you for shopping with Knowledge Mart!", "", 1, "C", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(0, 10, "Visit us at: www.knowledgemart.com", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		log.Printf("Failed to generate PDF: %v", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func printAddress(pdf *gofpdf.Fpdf, address models.ShippingAddress) {
	pdf.Cell(95, 6, fmt.Sprintf("%s, %s", address.StreetName, address.City))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("%s, %s, %s", address.State, address.PinCode, address.PhoneNumber))
	pdf.Ln(6)
}
