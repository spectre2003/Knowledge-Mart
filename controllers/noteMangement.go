package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateCourse(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	var request models.CreateNoteSharing
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid input", "error": err.Error()})
		return
	}

	course := models.Course{
		Name: request.Name,
	}
	if err := database.DB.Create(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Could not create course", "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": course})
}

func EditCourse(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "not authorized"})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "failed to retrieve admin information"})
		return
	}

	courseID := c.Param("id")
	if courseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "course_id is required",
		})
		return
	}

	var course models.Course
	if err := database.DB.First(&course, courseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Course not found"})
		return
	}

	var request models.CreateNoteSharing
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid input", "error": err.Error()})
		return
	}

	course.Name = request.Name
	if err := database.DB.Save(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Could not update course", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": course})
}

func DeleteCourse(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "not authorized"})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "failed to retrieve admin information"})
		return
	}

	courseID := c.Param("id")
	if courseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "course_id is required",
		})
		return
	}

	if err := database.DB.Delete(&models.Course{}, courseID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Could not delete course", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Course deleted successfully"})
}

func CreateSemester(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	courseID := c.Query("course_id")
	if courseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "course_id is required",
		})
		return
	}

	var course models.Course
	if err := database.DB.First(&course, courseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Course not found"})
		return
	}

	var request models.SemesterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid input", "error": err.Error()})
		return
	}

	semester := models.Semester{
		CourseID: course.CourseID,
		Number:   request.Number,
	}

	if err := database.DB.Create(&semester).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to create semester", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": semester})
}

func EditSemester(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	//courseID := c.Query("course_id")
	semesterID := c.Query("semester_id")

	if semesterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "semester_id is required",
		})
		return
	}

	var semester models.Semester
	if err := database.DB.Where("semester_id = ?", semesterID).First(&semester).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Semester not found"})
		return
	}

	var request models.SemesterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid input", "error": err.Error()})
		return
	}

	semester.Number = request.Number
	if err := database.DB.Save(&semester).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Could not update semester", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": semester})
}

func DeleteSemester(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	//courseID := c.Query("course_id")
	semesterID := c.Query("semester_id")

	if semesterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "semester_id is required",
		})
		return
	}

	if err := database.DB.Where("semester_id = ?", semesterID).Delete(&models.Semester{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Could not delete semester", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Semester deleted successfully"})
}

func CreateSubject(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	courseID := c.Query("course_id")
	semesterID := c.Query("semester_id")
	if courseID == "" || semesterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "categoryid and semesterid are required",
		})
		return
	}

	var course models.Course
	if err := database.DB.First(&course, courseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Course not found"})
		return
	}
	var semester models.Semester
	if err := database.DB.Where("semester_id = ? AND course_id = ?", semesterID, courseID).First(&semester).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Semester not found"})
		return
	}

	var request models.CreateNoteSharing

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid input", "error": err.Error()})
		return
	}

	subject := models.Subject{
		CourseID:   course.CourseID,
		SemesterID: semester.SemesterID,
		Name:       request.Name,
	}

	Course := course.Name
	Semester := semester.Number

	if err := database.DB.Create(&subject).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to create subject", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{
		"course":   Course,
		"semester": Semester,
		"details":  subject,
	}})
}

func EditSubject(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	// courseID := c.Query("course_id")
	// semesterID := c.Query("semester_id")
	subjectID := c.Query("subject_id")

	if subjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "subjectid are required",
		})
		return
	}

	var subject models.Subject
	if err := database.DB.Where("subject_id = ? ", subjectID).First(&subject).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Subject not found"})
		return
	}

	var request models.CreateNoteSharing
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid input", "error": err.Error()})
		return
	}

	subject.Name = request.Name
	if err := database.DB.Save(&subject).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Could not update subject", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": subject})
}

func DeleteSubject(c *gin.Context) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "not authorized ",
		})
		return
	}

	_, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve admin information",
		})
		return
	}

	// courseID := c.Query("course_id")
	// semesterID := c.Query("semester_id")
	subjectID := c.Query("subject_id")

	if subjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "categoryid ,semesterid and subjectid are required",
		})
		return
	}

	if err := database.DB.Where("subject_id = ? ", subjectID).Delete(&models.Subject{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Could not delete subject", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Subject deleted successfully"})
}

func UploadNote(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "User not authorized",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve user information",
		})
		return
	}

	var request models.UploadNote
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var course models.Course
	if err := database.DB.First(&course, request.CourseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Course not found"})
		return
	}

	var semester models.Semester
	if err := database.DB.Where("semester_id = ? AND course_id = ?", request.SemesterID, request.CourseID).First(&semester).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Semester not found"})
		return
	}

	var subject models.Subject
	if err := database.DB.Where("subject_id = ? AND semester_id = ? AND course_id = ?", request.SubjectID, request.SemesterID, request.CourseID).First(&subject).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Subject not found"})
		return
	}

	note := models.Note{
		UserID:      userIDUint,
		CourseID:    request.CourseID,
		SemesterID:  request.SemesterID,
		SubjectID:   request.SubjectID,
		Description: request.Description,
		FileURL:     request.FileURL,
	}

	if err := database.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to save note", "error": err.Error()})
		return
	}

	noteResponse := models.NoteResponse{
		NoteID:         note.NoteID,
		UserID:         note.UserID,
		CourseName:     course.Name,
		SemesterNumber: semester.Number,
		SubjectName:    subject.Name,
		Description:    note.Description,
		FileURL:        note.FileURL,
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": noteResponse})
}

func EditNote(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "User not authorized",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve user information",
		})
		return
	}

	noteID := c.Query("note_id")
	if noteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "note_id is required",
		})
		return
	}

	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Note not found"})
		return
	}

	if note.UserID != userIDUint {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "failed",
			"message": "You do not have permission to edit this note",
		})
		return
	}

	var request models.UploadNote
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note.CourseID = request.CourseID
	note.SemesterID = request.SemesterID
	note.Description = request.Description

	if request.FileURL != "" {
		note.FileURL = request.FileURL
	}

	if err := database.DB.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to update note", "error": err.Error()})
		return
	}

	var course models.Course
	if err := database.DB.First(&course, note.CourseID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to retrieve course information"})
		return
	}

	var semester models.Semester
	if err := database.DB.First(&semester, note.SemesterID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to retrieve semester information"})
		return
	}

	var subject models.Subject
	if err := database.DB.First(&subject, note.SubjectID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to retrieve subject information"})
		return
	}

	noteResponse := models.NoteResponse{
		NoteID:         note.NoteID,
		UserID:         note.UserID,
		CourseName:     course.Name,
		SemesterNumber: semester.Number,
		SubjectName:    subject.Name,
		Description:    note.Description,
		FileURL:        note.FileURL,
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": noteResponse})
}

func DeleteNote(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "User not authorized",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve user information",
		})
		return
	}

	noteID := c.Query("note_id")
	if noteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "note_id is required",
		})
		return
	}

	var note models.Note
	if err := database.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Note not found"})
		return
	}

	if note.UserID != userIDUint {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "failed",
			"message": "You do not have permission to delete this note",
		})
		return
	}

	if err := database.DB.Delete(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to delete note", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Note deleted successfully"})
}

func GetAllCoursesWithDetails(c *gin.Context) {
	var courses []models.Course

	if err := database.DB.
		Preload("Semesters", func(db *gorm.DB) *gorm.DB {
			return db.Order("number ASC")
		}).
		Preload("Semesters.Subjects").
		Find(&courses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to retrieve courses", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": courses})
}

func GetAllNotes(c *gin.Context) {
	var notes []struct {
		NoteID         uint   `json:"note_id"`
		UserID         uint   `json:"user_id"`
		CourseName     string `json:"course_name"`
		SemesterNumber string `json:"semester_number"`
		SubjectName    string `json:"subject_name"`
		Description    string `json:"description"`
		FileURL        string `json:"file_url"`
	}

	if err := database.DB.Table("notes").
		Select(`notes.note_id, notes.user_id, courses.name AS course_name, 
				semesters.number AS semester_number, 
				subjects.name AS subject_name, 
				notes.description, notes.file_url`).
		Joins("JOIN courses ON notes.course_id = courses.course_id").
		Joins("JOIN semesters ON notes.semester_id = semesters.semester_id").
		Joins("JOIN subjects ON notes.subject_id = subjects.subject_id").
		Scan(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Could not retrieve notes", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": notes})
}

func GetUserNotes(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "User not authorized",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to retrieve user information",
		})
		return
	}

	var notes []struct {
		NoteID         uint   `json:"note_id"`
		UserID         uint   `json:"user_id"`
		CourseName     string `json:"course_name"`
		SemesterNumber string `json:"semester_number"`
		SubjectName    string `json:"subject_name"`
		Description    string `json:"description"`
		FileURL        string `json:"file_url"`
	}

	if err := database.DB.Table("notes").
		Select(`notes.note_id, notes.user_id, courses.name AS course_name, 
				semesters.number AS semester_number, 
				subjects.name AS subject_name, 
				notes.description, notes.file_url`).
		Joins("JOIN courses ON notes.course_id = courses.course_id").
		Joins("JOIN semesters ON notes.semester_id = semesters.semester_id").
		Joins("JOIN subjects ON notes.subject_id = subjects.subject_id").
		Where("notes.user_id = ?", userIDUint).
		Scan(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Could not retrieve user notes", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": notes})
}
