package handlers

import (
	"hyw-webpics/database"
	"hyw-webpics/models"

	"github.com/gofiber/fiber/v2"
)

func GetCategories(c *fiber.Ctx) error {
	rows, err := database.DB.Query("SELECT id, name, slug FROM categories")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch categories",
		})
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug); err != nil {
			continue
		}
		categories = append(categories, cat)
	}

	return c.JSON(categories)
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func CreateCategory(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Name == "" || req.Slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name and Slug are required"})
	}

	result, err := database.DB.Exec("INSERT INTO categories (name, slug) VALUES (?, ?)", req.Name, req.Slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create category: " + err.Error()})
	}

	id, _ := result.LastInsertId()
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "name": req.Name, "slug": req.Slug})
}

func UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	query := "UPDATE categories SET name = ?, slug = ? WHERE id = ?"
	_, err := database.DB.Exec(query, req.Name, req.Slug, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update category"})
	}
	return c.JSON(fiber.Map{"message": "Category updated"})
}

func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	// Check if used
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM images WHERE category_id = ?", id).Scan(&count)
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot delete category used by images"})
	}

	_, err := database.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete category"})
	}
	return c.JSON(fiber.Map{"message": "Category deleted"})
}
