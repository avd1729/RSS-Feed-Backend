package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// Database connection string
const dbURL = "postgres://postgres:pass@localhost:5432/rss"

// Post struct
type Post struct {
	Title       string
	Link        string
	Description string
	PubDate     time.Time
}

// Fetch posts from PostgreSQL
func fetchPosts(conn *pgx.Conn) ([]Post, error) {
	rows, err := conn.Query(context.Background(), "SELECT title, link, description, pub_date FROM posts ORDER BY pub_date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.Title, &p.Link, &p.Description, &p.PubDate); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

// Generate RSS XML from posts
func generateRSS(posts []Post) string {
	rss := `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
    <channel>
        <title>My Blog</title>
        <link>https://example.com</link>
        <description>Latest blog posts</description>`

	for _, post := range posts {
		rss += fmt.Sprintf(`
        <item>
            <title>%s</title>
            <link>%s</link>
            <description>%s</description>
            <pubDate>%s</pubDate>
        </item>`, post.Title, post.Link, post.Description, post.PubDate.Format(time.RFC1123Z))
	}

	rss += `
    </channel>
</rss>`
	return rss
}

func main() {
	// Connect to PostgreSQL
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	r := gin.Default()

	// RSS feed endpoint
	r.GET("/rss", func(c *gin.Context) {
		posts, err := fetchPosts(conn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
			return
		}

		xml := generateRSS(posts)
		c.Data(http.StatusOK, "application/rss+xml", []byte(xml))
	})

	// Start server on port 8080
	r.Run(":8080")
}
