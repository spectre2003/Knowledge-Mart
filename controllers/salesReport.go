package controllers

import (
	"bytes"
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
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

		result, amount, _, err := TotalOrders(startDate, endDate, input.PaymentStatus, sellerIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"result":  result,
			"amount":  amount,
			"status":  true,
			"message": "successfully created sales report",
			//"order_counts": orderDateCounts,
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

	result, amount, _, err := TotalOrders(input.StartDate, input.EndDate, input.PaymentStatus, sellerIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully created sales report",
		"result":  result,
		"amount":  amount,
		//"order_counts": orderDateCounts,
	})
}

func TotalOrders(From string, Till string, PaymentStatus string, SellerID uint) (models.OrderCount, models.AmountInformation, []models.OrderDateCount, error) {
	var orders []models.Order

	parsedFrom, err := time.Parse("2006-01-02", From)
	if err != nil {
		return models.OrderCount{}, models.AmountInformation{}, nil, fmt.Errorf("error parsing From time: %v", err)
	}
	parsedTill, err := time.Parse("2006-01-02", Till)
	if err != nil {
		return models.OrderCount{}, models.AmountInformation{}, nil, fmt.Errorf("error parsing Till time: %v", err)
	}

	fFrom := time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, time.UTC)
	fTill := time.Date(parsedTill.Year(), parsedTill.Month(), parsedTill.Day(), 23, 59, 59, 999999999, time.UTC)

	query := database.DB.Where("ordered_at BETWEEN ? AND ? AND payment_status = ?", fFrom, fTill, PaymentStatus)
	if SellerID != 0 {
		query = query.Where("seller_id = ?", SellerID)
	}
	if err := query.Find(&orders).Error; err != nil {
		return models.OrderCount{}, models.AmountInformation{}, nil, fmt.Errorf("error fetching orders: %v", err)
	}

	var AccountInformation models.AmountInformation

	for _, order := range orders {
		AccountInformation.TotalAmountBeforeDeduction += RoundDecimalValue(order.TotalAmount)
		AccountInformation.TotalCouponDeduction += RoundDecimalValue(order.CouponDiscountAmount)
		AccountInformation.TotalProuctOfferDeduction += RoundDecimalValue(order.ProductOfferAmount)
		AccountInformation.TotalDeliveryCharges += RoundDecimalValue(order.DeliveryCharge)
		AccountInformation.TotalCategoryOfferDeduction += RoundDecimalValue(order.CategoryDiscountAmount)
		AccountInformation.TotalAmountAfterDeduction += RoundDecimalValue(order.FinalAmount)
	}

	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	orderIDs := make([]uint, len(orders))
	for i, order := range orders {
		orderIDs[i] = order.OrderID
	}

	if err := database.DB.Model(&models.OrderItem{}).
		Where("order_id IN (?)", orderIDs).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return models.OrderCount{}, models.AmountInformation{}, nil, fmt.Errorf("error counting order items by status: %v", err)
	}

	orderStatusCounts := make(map[string]int64)
	for _, sc := range statusCounts {
		orderStatusCounts[sc.Status] = sc.Count
	}

	var totalCount int64
	for _, count := range orderStatusCounts {
		totalCount += count
	}

	var orderDateCounts []models.OrderDateCount
	if err := database.DB.Model(&models.Order{}).
		Select("DATE(ordered_at) as date, COUNT(*) as count").
		Where("ordered_at BETWEEN ? AND ? AND payment_status = ?", fFrom, fTill, PaymentStatus).
		Group("DATE(ordered_at)").
		Scan(&orderDateCounts).Error; err != nil {
		return models.OrderCount{}, models.AmountInformation{}, nil, fmt.Errorf("error fetching order dates: %v", err)
	}

	return models.OrderCount{
		TotalOrder:     uint(totalCount),
		TotalPending:   uint(orderStatusCounts[models.OrderStatusPending]),
		TotalConfirmed: uint(orderStatusCounts[models.OrderStatusConfirmed]),
		TotalShipped:   uint(orderStatusCounts[models.OrderStatusShipped]),
		TotalDelivered: uint(orderStatusCounts[models.OrderStatusDelivered]),
		TotalCancelled: uint(orderStatusCounts[models.OrderStatusCanceled]),
		TotalReturned:  uint(orderStatusCounts[models.OrderStatusReturned]),
	}, AccountInformation, orderDateCounts, nil
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
		switch limit {
		case "day":
			startDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		case "week":
			startDate = time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Format("2006-01-02")
			endDate = time.Now().AddDate(0, 0, 7-int(time.Now().Weekday())).Format("2006-01-02")
		case "month":
			startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		case "year":
			startDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit specified, valid options are: day, week, month, year"})
			return
		}
	}

	orderCount, amountInfo, orderDateCounts, err := TotalOrders(startDate, endDate, paymentStatus, sellerIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	orderStatusChartPath := "/tmp/order_status_chart.png"
	orderHistoryChartPath := "/tmp/order_history_chart.png"

	err = generateOrderStatusChart(orderStatusChartPath, orderCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate order status chart"})
		return
	}
	err = generateOrderHistoryChart(orderHistoryChartPath, orderDateCounts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate order history chart"})
		return
	}

	pdfBytes, err := GenerateSalesReportPDF(orderCount, amountInfo, startDate, endDate, paymentStatus, orderStatusChartPath, orderHistoryChartPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=sales_report.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func generateOrderStatusChart(outputPath string, orderCount models.OrderCount) error {
	barChart := chart.BarChart{
		Width:  600,
		Height: 400,
		Bars: []chart.Value{
			{Value: float64(orderCount.TotalPending), Label: "Pending"},
			{Value: float64(orderCount.TotalConfirmed), Label: "Confirmed"},
			{Value: float64(orderCount.TotalShipped), Label: "Shipped"},
			{Value: float64(orderCount.TotalDelivered), Label: "Delivered"},
			{Value: float64(orderCount.TotalCancelled), Label: "Cancelled"},
			{Value: float64(orderCount.TotalReturned), Label: "Returned"},
		},
	}

	for i := range barChart.Bars {
		barChart.Bars[i].Style = chart.Style{
			FillColor:   drawing.ColorFromHex("87CEEB"),
			StrokeColor: drawing.ColorFromHex("4682B4"),
			StrokeWidth: 1.0,
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return barChart.Render(chart.PNG, file)
}

func generateOrderHistoryChart(outputPath string, orderDateCounts []models.OrderDateCount) error {
	endDate := time.Now().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -6)

	dateMap := make(map[string]int64)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dateMap[dateStr] = 0
	}

	for _, orderDateCount := range orderDateCounts {
		orderDateStr := orderDateCount.Date.Format("2006-01-02")
		if _, exists := dateMap[orderDateStr]; exists {
			dateMap[orderDateStr] += orderDateCount.Count
		}
	}

	var dates []time.Time
	var values []float64
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dates = append(dates, d)
		values = append(values, float64(dateMap[dateStr]))
	}

	graph := chart.Chart{
		Width:  600,
		Height: 300,
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeValueFormatterWithFormat("2006-01-02"),
		},
		YAxis: chart.YAxis{
			Name:      "Order Count",
			NameStyle: chart.Style{FontSize: 12, StrokeColor: drawing.ColorBlack},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name:    "Order History",
				XValues: dates,
				YValues: values,
				Style: chart.Style{
					StrokeColor: drawing.ColorBlue,
					StrokeWidth: 2.0,
				},
			},
		},
	}

	file, err := os.Create(outputPath)
	if err != nil {
		log.Printf("Failed to create output file: %s, error: %v", outputPath, err)
		return err
	}
	defer file.Close()

	if err := graph.Render(chart.PNG, file); err != nil {
		log.Printf("Failed to render chart: %v", err)
		return err
	}

	return nil
}

func GenerateSalesReportPDF(orderCount models.OrderCount, amountInfo models.AmountInformation, startDate string, endDate string, paymentStatus string, orderStatusChartPath string, orderHistoryChartPath string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Sales Report")
	pdf.Ln(12)

	// Report metadata
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Start Date: "+startDate)
	pdf.Ln(8)
	pdf.Cell(40, 10, "End Date: "+endDate)
	pdf.Ln(8)
	pdf.Cell(40, 10, "Payment Status: "+paymentStatus)
	pdf.Ln(12)

	// Summary Information
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Total Orders: "+strconv.Itoa(int(orderCount.TotalOrder)))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total Amount Before Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalAmountBeforeDeduction))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total Coupon Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalCouponDeduction))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total Category Offer Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalCategoryOfferDeduction))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total Product Offer Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalProuctOfferDeduction))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Delivery Charges: "+fmt.Sprintf("%.2f", amountInfo.TotalDeliveryCharges))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total Amount After Deduction: "+fmt.Sprintf("%.2f", amountInfo.TotalAmountAfterDeduction))
	pdf.Ln(12)

	// Order Status Summary Table
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

	// Adding Order Status Chart
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Order Status Chart")
	pdf.Ln(10)
	pdf.ImageOptions(orderStatusChartPath, 10, pdf.GetY(), 100, 60, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")

	pdf.Ln(70)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Order History Graph")
	pdf.Ln(10)
	pdf.ImageOptions(orderHistoryChartPath, 10, pdf.GetY(), 100, 60, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")

	// Generate PDF output
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

	orderCount, amountInfo, orderDateCounts, err := TotalOrders(startDate, endDate, paymentStatus, sellerIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	excelBytes, err := GenerateSalesReportExcel(orderCount, amountInfo, startDate, endDate, paymentStatus, orderDateCounts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate Excel report"})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=sales_report.xlsx")
	c.Header("Content-Length", strconv.Itoa(len(excelBytes)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
}

func GenerateSalesReportExcel(orderCount models.OrderCount, amountInfo models.AmountInformation, startDate, endDate, paymentStatus string, orderDateCounts []models.OrderDateCount) ([]byte, error) {
	f := excelize.NewFile()

	f.SetCellValue("Sheet1", "A1", "Sales Report")

	f.SetCellValue("Sheet1", "A2", "Start Date")
	f.SetCellValue("Sheet1", "B2", startDate)
	f.SetCellValue("Sheet1", "A3", "End Date")
	f.SetCellValue("Sheet1", "B3", endDate)
	f.SetCellValue("Sheet1", "A4", "Payment Status")
	f.SetCellValue("Sheet1", "B4", paymentStatus)

	// Set up order summary
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

	// Fixed the row number for total delivery charges
	f.SetCellValue("Sheet1", "A11", "Total Delivery Charges")
	f.SetCellValue("Sheet1", "B11", fmt.Sprintf("%.2f", amountInfo.TotalDeliveryCharges))

	f.SetCellValue("Sheet1", "A12", "Total Amount After Deduction")
	f.SetCellValue("Sheet1", "B12", fmt.Sprintf("%.2f", amountInfo.TotalAmountAfterDeduction))

	// Adding Order Status Summary
	f.SetCellValue("Sheet1", "A14", "Order Status Summary")
	f.SetCellValue("Sheet1", "A15", "Total Pending Orders")
	f.SetCellValue("Sheet1", "B15", strconv.Itoa(int(orderCount.TotalPending)))

	f.SetCellValue("Sheet1", "A16", "Total Confirmed Orders")
	f.SetCellValue("Sheet1", "B16", strconv.Itoa(int(orderCount.TotalConfirmed)))

	f.SetCellValue("Sheet1", "A17", "Total Shipped Orders")
	f.SetCellValue("Sheet1", "B17", strconv.Itoa(int(orderCount.TotalShipped)))

	f.SetCellValue("Sheet1", "A18", "Total Delivered Orders")
	f.SetCellValue("Sheet1", "B18", strconv.Itoa(int(orderCount.TotalDelivered)))

	f.SetCellValue("Sheet1", "A19", "Total Cancelled Orders")
	f.SetCellValue("Sheet1", "B19", strconv.Itoa(int(orderCount.TotalCancelled)))

	f.SetCellValue("Sheet1", "A20", "Total Returned Orders")
	f.SetCellValue("Sheet1", "B20", strconv.Itoa(int(orderCount.TotalReturned)))

	// Write to buffer
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

	// User address section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 6, "Ship To:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 6, user.Name)
	pdf.Ln(6)
	printAddress(pdf, billingAddress)
	pdf.Ln(10)

	// Order information section
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

	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(85, 10, "Total", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 10, fmt.Sprintf("%.2f", totalPrice), "1", 0, "C", false, 0, "")
	pdf.CellFormat(24, 10, fmt.Sprintf("%.2f", totalProductOffer), "1", 0, "C", false, 0, "")
	pdf.CellFormat(26, 10, fmt.Sprintf("%.2f", totalCategoryOffer), "1", 0, "C", false, 0, "")
	pdf.CellFormat(22, 10, fmt.Sprintf("%.2f", totalOtherOffers), "1", 0, "C", false, 0, "")
	pdf.CellFormat(23, 10, fmt.Sprintf("%.2f", totalFinalAmount), "1", 0, "C", false, 0, "")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Ln(5)
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
