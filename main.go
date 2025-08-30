package main

import (
	"log"
	"net/http"

	"strconv"

	"github.com/clintongilders/go-api-client/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("/tmp/test.db?_foreign_keys=on"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	} else {
		println("Database connection successful.")
	}
	// Migrate the schema
	db.AutoMigrate(&models.Region{}, &models.PokemonSpecies{})
	println("Database Migrated")
	return db
}

func main() {
	// Initialize GORM with SQLite
	db := InitDB()
	r := gin.Default()

	// Read all
	r.GET("/v1/regions", func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("pageSize", "10")

		page, _ := strconv.Atoi(pageStr)
		pageSize, _ := strconv.Atoi(pageSizeStr)

		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		var totalRecords int64
		var items []models.Region
		// Count total records
		db.Model(&items).Count(&totalRecords)

		// Fetch paginated records
		db.Limit(pageSize).Offset(offset).Find(&items)

		// Calculate total pages
		totalPages := (totalRecords + int64(pageSize) - 1) / int64(pageSize)
		var nextPageUri, previousPageUri string
		nextPage := page + 1
		if int64(nextPage) > totalPages {
			nextPage = 0
		} else {
			nextPageUri = c.Request.URL.Path + "?page=" + strconv.Itoa(nextPage) + "&pageSize=" + strconv.Itoa(pageSize)
		}
		previousPage := page - 1
		if previousPage <= 0 {
			previousPage = 0
		} else {
			previousPageUri = c.Request.URL.Path + "?page=" + strconv.Itoa(previousPage) + "&pageSize=" + strconv.Itoa(pageSize)
		}

		c.JSON(200, gin.H{
			"currentPage":     page,
			"pageSize":        pageSize,
			"totalPages":      totalPages,
			"totalRecords":    totalRecords,
			"nextPageUri":     nextPageUri,
			"previousPageUri": previousPageUri,
			"data":            items,
		})
	})

	// Read one
	r.GET("/v1/regions/:id", func(c *gin.Context) {
		var item models.Region
		if err := db.First(&item, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		//c.JSON(http.StatusOK, item)
		c.JSON(200, gin.H{
			"data":         item,
			"currentPage":  1,
			"pageSize":     1,
			"totalPages":   1,
			"totalRecords": 1,
		})
	})

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("failed to run server")
	}
}
