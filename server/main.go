package main

import (
	"context"
	"fmt"
	"log"
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
	r.POST("/stats", func(c *gin.Context) { HandleStats(c, x) })

	r.Run()
}

func HandlePick(c *gin.Context, x *Context) {
	type Request struct {
		Username string `json:"username" binding:"required,max=32"`
	}

	type Response struct {
		Content string `json:"content"`
		Author  string `json:"author"`
		Creator string `json:"creator"`
	}

	var r Request
	if c.Bind(&r) != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	tx, err := x.P.Begin(ctx)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var content string
	var author string
	var creator string
	err = tx.QueryRow(context.Background(),
		`SELECT content, author, creator 
		FROM messages 
		ORDER BY random() LIMIT 1`,
	).Scan(&content, &author, &creator)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO visits (username, type) 
		VALUES ($1, 'Pick')`,
		r.Username,
	)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, Response{Content: content, Author: author, Creator: creator})
}

func HandleCreate(c *gin.Context, x *Context) {
	type Request struct {
		Content  string `json:"content" binding:"required,max=512"`
		Author   string `json:"author" binding:"required,max=32"`
		Username string `json:"username" binding:"required,max=32"`
	}

	type Response struct {
		AllCount  uint `json:"all_count" binding:"required"`
		UserCount uint `json:"user_count" binding:"required"`
	}

	var r Request
	if c.Bind(&r) != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	tx, err := x.P.Begin(ctx)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO messages (content, author, creator) 
		VALUES ($1, $2, $3)`,
		r.Content,
		r.Author,
		r.Username,
	)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var allCount uint
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages`,
	).Scan(&allCount)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var userCount uint
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages WHERE creator=$1`,
		r.Username,
	).Scan(&userCount)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO visits (username, type) 
		VALUES ($1, 'Create')`,
		r.Username,
	)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, Response{
		AllCount:  allCount,
		UserCount: userCount,
	})
}

func HandleStats(c *gin.Context, x *Context) {
	type Request struct {
		Username string `json:"username" binding:"required,max=32"`
	}

	type Response struct {
		AllCount    uint `json:"all_count" binding:"required"`
		UserCount   uint `json:"user_count" binding:"required"`
		AllVisits   uint `json:"all_visits" binding:"required"`
		TodayVisits uint `json:"today_visits" binding:"required"`
	}

	var r Request
	if c.Bind(&r) != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	var allCount uint
	err := x.P.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages`,
	).Scan(&allCount)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var userCount uint
	err = x.P.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages WHERE creator=$1`,
		r.Username,
	).Scan(&userCount)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var allVisits uint
	err = x.P.QueryRow(ctx,
		`SELECT COUNT(*) FROM visits`,
	).Scan(&allVisits)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var todayVisits uint
	err = x.P.QueryRow(ctx,
		`SELECT COUNT(*) FROM visits WHERE created >= CURRENT_DATE AND created <= CURRENT_DATE + 1`,
	).Scan(&todayVisits)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, Response{
		AllCount:    allCount,
		UserCount:   userCount,
		AllVisits:   allVisits,
		TodayVisits: todayVisits,
	})
}
