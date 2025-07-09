package database

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user of the application
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"` // Store hashed passwords only!
	// Books        []Book `gorm:"many2many:user_books;"` // For user-specific library features later
}

// Author represents the author of a book
type Author struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex;not null"`
	Books []Book `gorm:"foreignKey:AuthorID"`
}

// Series represents a series of books
type Series struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex;not null"`
	Books []Book `gorm:"foreignKey:SeriesID"`
}

// Book represents an ebook in the library
type Book struct {
	gorm.Model
	Title         string    `gorm:"not null"`
	AuthorID      uint      // Foreign key for Author
	Author        Author    // Belongs to Author
	SeriesID      *uint     // Foreign key for Series (nullable)
	Series        *Series   // Belongs to Series (nullable)
	SeriesIndex   *float64  // Book's index in the series (e.g., 1, 2.5). Nullable.
	ISBN          string    `gorm:"index"`
	FilePath      string    `gorm:"not null;uniqueIndex"` // Path to the original file in the library
	CoverPath     string    // Path to the extracted cover image
	Format        string    // Original format of the book (e.g., EPUB, MOBI, AZW3)
	AddedAt       time.Time `gorm:"autoCreateTime"`
	LastReadAt    *time.Time
	ReadingETag   string // For concurrency control when syncing reading position
	Publisher     string
	PublishedDate string
	Description   string
	Language      string
	Rating        *float32 // User rating, 0-5 stars
	PageCount     *int
	// Tags          []Tag `gorm:"many2many:book_tags;"` // For user-defined tags
}

// TODO: Consider a Tag model if we want user-defined tagging later
// type Tag struct {
// 	gorm.Model
// 	Name  string `gorm:"uniqueIndex;not null"`
// 	Books []Book `gorm:"many2many:book_tags;"`
// }
