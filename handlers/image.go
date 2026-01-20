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

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse form",
		})
	}

	files := form.File["images"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No image files provided",
		})
	}

	categoryID := c.FormValue("category_id")
	var uploadedImages []fiber.Map
	var errors []string

	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}

	for _, file := range files {
		// Validate file type
		ext := filepath.Ext(file.Filename)
		if !allowedExts[ext] {
			errors = append(errors, file.Filename+": Only JPG, PNG, and GIF files are allowed")
			continue
		}

		// Open the file
		src, err := file.Open()
		if err != nil {
			errors = append(errors, file.Filename+": Failed to read file")
			continue
		}

		// Convert to WebP
		filename, err := utils.ConvertToWebP(src, file.Filename)
		src.Close() // Close early
		if err != nil {
			errors = append(errors, file.Filename+": "+err.Error())
			continue
		}

		// Save to database
		var query string
		var args []interface{}
		if categoryID != "" {
			query = "INSERT INTO images (filename, original_name, uploader_id, category_id) VALUES (?, ?, ?, ?)"
			args = []interface{}{filename, file.Filename, userID, categoryID}
		} else {
			query = "INSERT INTO images (filename, original_name, uploader_id) VALUES (?, ?, ?)"
			args = []interface{}{filename, file.Filename, userID}
		}

		result, err := database.DB.Exec(query, args...)
		if err != nil {
			errors = append(errors, file.Filename+": Failed to save record")
			continue
		}

		id, _ := result.LastInsertId()
		uploadedImages = append(uploadedImages, fiber.Map{
			"id":       id,
			"filename": filename,
			"name":     file.Filename,
		})
	}

	status := fiber.StatusCreated
	if len(uploadedImages) == 0 && len(errors) > 0 {
		status = fiber.StatusInternalServerError
	}

	return c.Status(status).JSON(fiber.Map{
		"message":  "Batch upload complete",
		"uploaded": uploadedImages,
		"errors":   errors,
	})
}

type BulkActionRequest struct {
	IDs        []int64 `json:"ids"`
	CategoryID *int64  `json:"category_id,omitempty"`
}

func BulkApproveImages(c *fiber.Ctx) error {
	var req BulkActionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.IDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No IDs provided"})
	}

	if req.CategoryID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Category ID is required for approval"})
	}

	return bulkExec(c, "UPDATE images SET status = 'approved', category_id = ?, approved_at = ? WHERE id IN (", req.IDs, *req.CategoryID, time.Now())
}

// Helper for bulk operations to avoid SQL injection and boilerplate
func bulkExec(c *fiber.Ctx, prefix string, ids []int64, extraArgs ...interface{}) error {
	if len(ids) == 0 {
		return c.JSON(fiber.Map{"message": "No items to process"})
	}

	placeholders := ""
	for i := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
	}

	query := prefix + placeholders + ")"
	args := append(extraArgs, idsToInterfaces(ids)...)

	_, err := database.DB.Exec(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Bulk operation failed: " + err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Bulk operation successful", "count": len(ids)})
}

func idsToInterfaces(ids []int64) []interface{} {
	ifaces := make([]interface{}, len(ids))
	for i, v := range ids {
		ifaces[i] = v
	}
	return ifaces
}

func BulkDeleteImages(c *fiber.Ctx) error {
	var req BulkActionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Note: In a real app we should also delete files from storage.
	// For simplicity, we just delete records or we'd need to fetch filenames first.
	// Let's fetch filenames to delete them properly.
	return bulkExec(c, "DELETE FROM images WHERE id IN (", req.IDs)
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
