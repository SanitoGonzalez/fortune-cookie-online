package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Settings struct {
	DbHost     string
	DbPort     int
	DbName     string
	DbUser     string
	DbPassword string
}

func newSettings() *Settings {
	s := &Settings{}

	s.DbHost = os.Getenv("FORTUNE_DB_HOST")

	port, err := strconv.Atoi(os.Getenv("FORTUNE_DB_PORT"))
	if err != nil {
		panic(err)
	}
	s.DbPort = port

	s.DbName = os.Getenv("FORTUNE_DB_NAME")

	s.DbUser = os.Getenv("FORTUNE_DB_USER")

	data, err := os.ReadFile(os.Getenv("FORTUNE_DB_PASSWORD_FILE"))
	if err != nil {
		panic(err)
	}
	s.DbPassword = strings.TrimSpace(string(data))

	return s
}

type Context struct {
	S *Settings
	P *pgxpool.Pool
}

func newContext() *Context {
	s := newSettings()

	pool, err := pgxpool.New(context.Background(), fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		s.DbUser, s.DbPassword, s.DbHost, s.DbPort, s.DbName))
	if err != nil {
		panic(err)
	}

	return &Context{
		S: s,
		P: pool,
	}
}

func main() {
	x := newContext()

	r := gin.Default()
	r.POST("/pick", func(c *gin.Context) { HandlePick(c, x) })
	r.POST("/create", func(c *gin.Context) { HandleCreate(c, x) })

	r.Run()
}

func HandlePick(c *gin.Context, x *Context) {
	type Response struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}

	var content string
	var author string
	err := x.P.QueryRow(context.Background(),
		`SELECT content, author 
		FROM messages 
		ORDER BY random() LIMIT 1`,
	).Scan(&content, &author)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, Response{Content: content, Author: author})
}

func HandleCreate(c *gin.Context, x *Context) {

}
