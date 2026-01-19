package handlers

import (
	"database/sql"
	"math/rand"
	"path/filepath"
	"time"

	"hyw-webpics/database"
	"hyw-webpics/models"
	"hyw-webpics/utils"

	"github.com/gofiber/fiber/v2"
)

func UploadImage(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No image file provided",
		})
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowedExts[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Only JPG, PNG, and GIF files are allowed",
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read file",
		})
	}
	defer src.Close()

	// Convert to WebP
	filename, err := utils.ConvertToWebP(src, file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process image: " + err.Error(),
		})
	}

	// Actually we need to parse it to int64
	// But Fiber's FormValue returns string.
	// Let's use BodyParser for cleaner code if possible, but this is multipart.
	// So just standard parsing.

	// Save to database
	query := "INSERT INTO images (filename, original_name, uploader_id) VALUES (?, ?, ?)"
	result, err := database.DB.Exec(query, filename, file.Filename, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save image record: " + err.Error(),
		})
	}

	id, _ := result.LastInsertId()
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Image uploaded successfully, pending review",
		"image": fiber.Map{
			"id":       id,
			"filename": filename,
		},
	})
}

func GetApprovedImages(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	offset := (page - 1) * limit

	categoryID := c.Query("category_id")

	baseQuery := "SELECT id, filename, original_name, uploader_id, category_id, status, created_at, approved_at FROM images WHERE status = 'approved'"
	countQuery := "SELECT COUNT(*) FROM images WHERE status = 'approved'"
	var args []interface{}

	if categoryID != "" {
		baseQuery += " AND category_id = ?"
		countQuery += " AND category_id = ?"
		args = append(args, categoryID)
	}

	baseQuery += " ORDER BY approved_at DESC LIMIT ? OFFSET ?"

	// Append pagination args
	queryArgs := append(args, limit, offset)

	rows, err := database.DB.Query(baseQuery, queryArgs...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch images",
		})
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		var approvedAt sql.NullTime
		err := rows.Scan(&img.ID, &img.Filename, &img.OriginalName, &img.UploaderID, &img.CategoryID, &img.Status, &img.CreatedAt, &approvedAt)
		if err != nil {
			continue
		}
		if approvedAt.Valid {
			img.ApprovedAt = &approvedAt.Time
		}
		images = append(images, img)
	}

	// Get total count
	var total int
	database.DB.QueryRow(countQuery, args...).Scan(&total)

	return c.JSON(fiber.Map{
		"images": images,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func GetRandomImage(c *fiber.Ctx) error {
	categoryID := c.Query("category_id")

	countQuery := "SELECT COUNT(*) FROM images WHERE status = 'approved'"
	baseQuery := "SELECT id, filename, original_name, uploader_id, category_id, status, created_at, approved_at FROM images WHERE status = 'approved'"
	var args []interface{}

	if categoryID != "" {
		countQuery += " AND category_id = ?"
		baseQuery += " AND category_id = ?"
		args = append(args, categoryID)
	}

	var count int
	err := database.DB.QueryRow(countQuery, args...).Scan(&count)
	if err != nil || count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No approved images found",
		})
	}

	rand.Seed(time.Now().UnixNano())
	offset := rand.Intn(count)

	baseQuery += " LIMIT 1 OFFSET ?"
	args = append(args, offset)

	var img models.Image
	var approvedAt sql.NullTime
	err = database.DB.QueryRow(baseQuery, args...).Scan(&img.ID, &img.Filename, &img.OriginalName, &img.UploaderID, &img.CategoryID, &img.Status, &img.CreatedAt, &approvedAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch random image",
		})
	}

	if approvedAt.Valid {
		img.ApprovedAt = &approvedAt.Time
	}

	return c.JSON(img)
}
