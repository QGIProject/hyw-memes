package handlers

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"hyw-webpics/config"
	"hyw-webpics/database"
	"hyw-webpics/models"

	"github.com/gofiber/fiber/v2"
)

type AdminLoginRequest struct {
	Password string `json:"password"`
}

func AdminLogin(c *fiber.Ctx) error {
	var req AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Password != config.AppConfig.AdminPassword {
		// DEBUG LOGGING
		// fmt.Printf("DEBUG LOGIN: Received '%s', Expected '%s'\n", req.Password, config.AppConfig.AdminPassword)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid admin password",
		})
	}

	// Set admin session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "admin_session",
		Value:    "authenticated",
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.JSON(fiber.Map{
		"message": "Admin login successful",
	})
}

func AdminLogout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "admin_session",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.JSON(fiber.Map{
		"message": "Admin logout successful",
	})
}

func GetPendingImages(c *fiber.Ctx) error {
	rows, err := database.DB.Query(
		`SELECT id, filename, original_name, uploader_id, category_id, status, created_at 
		 FROM images WHERE status = 'pending' ORDER BY created_at ASC`,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch pending images",
		})
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		err := rows.Scan(&img.ID, &img.Filename, &img.OriginalName, &img.UploaderID, &img.CategoryID, &img.Status, &img.CreatedAt)
		if err != nil {
			continue
		}
		images = append(images, img)
	}

	return c.JSON(fiber.Map{
		"images": images,
		"count":  len(images),
	})
}

type ApproveRequest struct {
	CategoryID int64 `json:"category_id"`
}

func ApproveImage(c *fiber.Ctx) error {
	id := c.Params("id")

	var req ApproveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.CategoryID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Category ID is required",
		})
	}

	result, err := database.DB.Exec(
		"UPDATE images SET status = 'approved', approved_at = ?, category_id = ? WHERE id = ? AND status = 'pending'",
		time.Now(), req.CategoryID, id,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to approve image",
		})
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found or already processed",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Image approved successfully",
	})
}

func RejectImage(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get filename before deleting
	var filename string
	err := database.DB.QueryRow("SELECT filename FROM images WHERE id = ?", id).Scan(&filename)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found",
		})
	}

	// Delete from database
	result, err := database.DB.Exec("DELETE FROM images WHERE id = ?", id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete image",
		})
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found",
		})
	}

	// Delete file from disk
	filePath := filepath.Join(config.AppConfig.UploadDir, filename)
	os.Remove(filePath)

	return c.JSON(fiber.Map{
		"message": "Image and file deleted successfully",
	})
}

// Updated GetAdminImages to support listing all images with pagination
func GetAdminImages(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	offset := (page - 1) * limit
	status := c.Query("status") // "pending", "approved", or empty for all

	// Build query
	var query string
	var countQuery string
	var args []interface{}

	if status != "" {
		query = "SELECT id, filename, original_name, uploader_id, category_id, status, created_at FROM images WHERE status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
		countQuery = "SELECT COUNT(*) FROM images WHERE status = ?"
		args = append(args, status)
	} else {
		query = "SELECT id, filename, original_name, uploader_id, category_id, status, created_at FROM images ORDER BY created_at DESC LIMIT ? OFFSET ?"
		countQuery = "SELECT COUNT(*) FROM images"
	}

	// Execute count query
	var total int
	database.DB.QueryRow(countQuery, args...).Scan(&total)

	// Execute data query
	args = append(args, limit, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch images"})
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		if err := rows.Scan(&img.ID, &img.Filename, &img.OriginalName, &img.UploaderID, &img.CategoryID, &img.Status, &img.CreatedAt); err == nil {
			images = append(images, img)
		}
	}

	return c.JSON(fiber.Map{
		"images": images,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func GetAdminStats(c *fiber.Ctx) error {
	var totalImages, pendingImages, approvedImages, totalCategories int

	database.DB.QueryRow("SELECT COUNT(*) FROM images").Scan(&totalImages)
	database.DB.QueryRow("SELECT COUNT(*) FROM images WHERE status = 'pending'").Scan(&pendingImages)
	database.DB.QueryRow("SELECT COUNT(*) FROM images WHERE status = 'approved'").Scan(&approvedImages)
	database.DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&totalCategories)

	return c.JSON(fiber.Map{
		"total_images":     totalImages,
		"pending_images":   pendingImages,
		"approved_images":  approvedImages,
		"total_categories": totalCategories,
	})
}
