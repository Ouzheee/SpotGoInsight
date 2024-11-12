package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Book represents a book in the bookshelf
type Book struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Pages int    `json:"pages"`
}

var nextID = 2

// Initialize bookshelf with one book
var bookshelf = []Book{
	{ID: 1, Name: "Blue Bird", Pages: 500},
}

// List all books
func getBooks(c *gin.Context) {
	c.JSON(http.StatusOK, bookshelf)
}

// Get a specific book by ID
func getBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid book ID"})
		return
	}

	for _, book := range bookshelf {
		if book.ID == id {
			c.JSON(http.StatusOK, book)
			return
		}
	}
	c.JSON(404, gin.H{"message": "book not found"})
}

// Add a new book
func addBook(c *gin.Context) {
	var newBook Book
	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(404, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	// Check for duplicate name (case-sensitive)
	for _, book := range bookshelf {
		if book.Name == newBook.Name {
			c.JSON(409, gin.H{"message": "duplicate book name"})
			return
		}
	}

	// Assign a new ID and add the book to the list
	newBook.ID = nextID
	nextID++
	bookshelf = append(bookshelf, newBook)
	c.JSON(201, newBook)
}

// Delete a book by ID
func deleteBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		/*c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid book ID"})
		return*/
	}

	for index, book := range bookshelf {
		if book.ID == id {
			// Remove the book from the slice
			bookshelf = append(bookshelf[:index], bookshelf[index+1:]...)
			c.Status(204) // Respond with 404
			return
		}
	}
	c.Status(204) // post 204 if we delete success
}

// Update a book by ID with new name and pages
func updateBook(c *gin.Context) {
	// Get the book Id from the URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid book ID"})
		return
	}

	// Bind incoming JSON data to a map to allow partial updates
	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(404, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	// Find the book with specified ID
	var book *Book
	for i := range bookshelf {
		if bookshelf[i].ID == id {
			book = &bookshelf[i]
			break
		}
	}
	if book == nil {
		c.JSON(404, gin.H{"message": "book not found"})
		return
	}

	// Update fields based on the provided data
	if name, exists := updateData["name"].(string); exists && name != "" {
		for _, existbook := range bookshelf {
			if existbook.Name == name {
				c.JSON(409, gin.H{"message": "duplicate book name"})
				return
			}
		}
		book.Name = name
	}

	if pages, exists := updateData["pages"].(float64); exists {
		book.Pages = int(pages) // Convert float64 to int
	}

	// Respond with the updated book data
	c.JSON(http.StatusOK, book)
}

func main() {
	route := gin.Default()

	// Define routes
	route.GET("/bookshelf", getBooks)
	route.GET("/bookshelf/:id", getBook)
	route.POST("/bookshelf", addBook)
	route.DELETE("/bookshelf/:id", deleteBook)
	route.PUT("/bookshelf/:id", updateBook)

	// Start server
	if err := route.Run(":8087"); err != nil {
		panic("Failed to run server: " + err.Error())
	}
}
