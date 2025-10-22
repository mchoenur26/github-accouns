package handlers

import ( //library
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"my-client-api/models"
	"my-client-api/services"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4/pgxpool"
)

// struktur handler
type ClientHandler struct {
	DB *pgxpool.Pool //koneksi dengan PostgreSQL

	Redis services.RedisCacher //koneksi dengan Redis

	S3 services.FileUploader //inisialisasi upload ke S3
}

// read database agar lebih cepat bisa ditemukan
func (h *ClientHandler) GetClientBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	cachedData, err := h.Redis.GetClient(slug)
	if err != nil {
		log.Printf("ERROR REDIS GET: %v", err)
	} else if cachedData != "" {
		return c.SendString(cachedData)
	}
	//struktur untuk memasukan data ke database
	var client models.Client
	query := `
		SELECT id, name, slug, client_prefix, client_logo_url, address, phone_number, city, created_at, updated_at, deleted_at 
		FROM my_client 
		WHERE slug = $1 AND deleted_at IS NULL
	`
	err = h.DB.QueryRow(context.Background(), query, slug).Scan(
		&client.ID, &client.Name, &client.Slug, &client.ClientPrefix, &client.ClientLogoURL,
		&client.Address, &client.PhoneNumber, &client.City,
		&client.CreatedAt, &client.UpdatedAt, &client.DeletedAt,
	)
	//fungsi error jika terdapat kesalahan pada data
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user tidak ditemukan"})
	} else if err != nil {
		log.Printf("ERROR DB QUERY: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data user"})
	} //memastikan data bisa dikembalikan tanpa rusak dan merusak penyimpanan
	jsonData, marshalErr := json.Marshal(client)
	if marshalErr != nil {
		log.Printf("ERROR JSON MARSHAL: %v", marshalErr)
	} else {
		if cacheErr := h.Redis.SetClient(slug, string(jsonData)); cacheErr != nil {
			log.Printf("ERROR REDIS SET: %v", cacheErr)
		}
	}

	return c.JSON(client)
}

// inisialisasi parsing input
func (h *ClientHandler) CreateClient(c *fiber.Ctx) error {
	var input struct {
		Name         string `json:"name"`
		ClientPrefix string `json:"client_prefix"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	} //memastikan adanya input
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Nama wajib diisi"})
	}
	//pembuatan slug untuk identitas user
	slug := strings.ReplaceAll(strings.ToLower(input.Name), " ", "-")
	//penamaan file yang diinput
	file, err := c.FormFile("client_logo")
	var logoURL string
	//pengecekan file tidak error saat di upload
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		isMissingFileError := ok && fiberErr.Code == fiber.StatusNotFound

		if !isMissingFileError {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Gagal memproses file: %v", err)})
		}
	}
	//upload file ke S3
	if file != nil {
		logoURL, err = h.S3.UploadFile(c.Context(), file, slug)
		if err != nil { //jika terjadi error sat proses upload
			log.Printf("ERROR S3 UPLOAD: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengunggah logo"})
		}
	}
	//penyimpanan ke database
	_, err = h.DB.Exec(c.Context(),
		"INSERT INTO my_client (name, slug, client_logo_url, client_prefix) VALUES ($1, $2, $3, $4)",
		input.Name, slug, logoURL, input.ClientPrefix,
	) //jika penyimpanan gagal
	if err != nil {
		log.Printf("ERROR DB INSERT: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan user"})
	}
	//memastikan nama file dalam database tidak  berulang hingga menyebabkan error
	if err := h.Redis.DeleteClient(slug); err != nil {
		log.Printf("ERROR REDIS DELETE: %v", err)
	}
	//mengembalikan value jika berhasil tersimpan
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "user berhasil dibuat", "slug": slug, "logo_url": logoURL})
}

// optimisasi penyimpanan dengan menghapus cache
func (h *ClientHandler) DeleteClient(c *fiber.Ctx) error {
	slug := c.Params("slug")

	now := time.Now()
	query := "UPDATE my_client SET deleted_at = $1 WHERE slug = $2 AND deleted_at IS NULL"
	//inisialisasi query database
	result, err := h.DB.Exec(context.Background(), query, now, slug)

	if err != nil {
		log.Printf("ERROR DB EXEC: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal melakukan soft delete"})
	}
	//memastikan tidak adanya error dalam proses upload ke database
	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user"})
	}
	//memastikan value chace tidak akan terduplikasi dengan cara menghapus sebagian
	if err := h.Redis.DeleteClient(slug); err != nil {
		log.Printf("ERROR REDIS DELETE: %v", err)
	}
	//jika user berhasil dihapus
	return c.JSON(fiber.Map{"message": "user berhasil dihapus"})
}
