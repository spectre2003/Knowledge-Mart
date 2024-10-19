package controllers

import (
	"errors"
	"fmt"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SellerOverAllSalesReport generates an overall sales report for the seller
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

	// Validate input
	if input.StartDate == "" && input.EndDate == "" && input.Limit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide start date and end date, or specify the limit as day, week, month, year"})
		return
	}

	// Validate and process limit
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

		// Determine date range based on limit
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

		// Fetch results based on the limit
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

	// Validate start and end date if limit is not provided
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

	// Fetch results based on date range
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

// TotalOrders calculates the total order count and amount information for the given date range and payment status
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

	// Query orders for the seller within the given time range
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

	// Process each order and calculate amounts and status counts
	for _, order := range orders {
		AccountInformation.TotalCouponDeduction += RoundDecimalValue(order.CouponDiscountAmount)
		AccountInformation.TotalAmountBeforeDeduction += RoundDecimalValue(order.TotalAmount)
		AccountInformation.TotalAmountAfterDeduction += RoundDecimalValue(order.FinalAmount)

		// Count order status occurrences
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

	// Return total order count and amount information
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

// RoundDecimalValue rounds a float64 value to 2 decimal places
func RoundDecimalValue(value float64) float64 {
	multiplier := math.Pow(10, 2)
	return math.Round(value*multiplier) / multiplier
}
