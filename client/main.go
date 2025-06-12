package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server struct {
		Host string `toml:"host"`
		Port uint16 `toml:"port"`
	} `toml:"server"`
}

func (c *Config) server_url() string {
	return fmt.Sprintf("http://%s:%d", c.Server.Host, c.Server.Port)
}

func loadConfig() *Config {
	config := &Config{}
	if _, err := toml.DecodeFile("config.toml", config); err != nil {
		fmt.Printf("config.toml 파일이 존재하지 않거나 잘못되었습니다!")
		os.Exit(1)
	}

	return config
}

func main() {
	config := loadConfig()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(
			`----- 선택 -----
(1) 포츈 쿠키 열기
(2) 포츈 쿠키 만들기
`)

		action, _ := reader.ReadString('\n')
		action = strings.TrimSpace(action)
		i, err := strconv.Atoi(action)
		if err != nil {
			fmt.Print("1 또는 2를 입력하세요.\n\n")
			continue
		}

		switch i {
		case 1:
			handlePick(config)
		case 2:
			handleCreate(config, reader)
		default:
			fmt.Print("1 또는 2를 입력하세요.\n\n")
			continue
		}

		break
	}
}

func handlePick(c *Config) {
	resp, err := http.Post(c.server_url()+"/pick", "application/json", nil)
	if err != nil {
		fmt.Println("요청 실패:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("잘못된 응답:", resp.Status)
		return
	}

	var result struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("잘못된 응답 데이터:", err)
		return
	}

	fmt.Printf("\n\"%s\"\n  - %s\n", result.Content, result.Author)
}

func handleCreate(c *Config, reader *bufio.Reader) {
	fmt.Print("문구를 입력하세요: ")
	content, _ := reader.ReadString('\n')
	content = strings.TrimSpace(content)

	if len(content) == 0 || len(content) > 256 {
		fmt.Println("문구의 길이가 잘못됐습니다.")
		return
	}

	fmt.Print("이름을 입력하세요: ")
	author, _ := reader.ReadString('\n')
	author = strings.TrimSpace(author)

	if len(author) == 0 || len(author) > 32 {
		fmt.Println("이름의 길이가 잘못됐습니다.")
		return
	}

	payload := map[string]string{
		"content": content,
		"author":  author,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(c.server_url()+"/create", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("요청 실패:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("잘못된 응답:", resp.Status)
		return
	}

	var result struct {
		AllCount    uint `json:"all_count"`
		AuthorCount uint `json:"author_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("잘못된 응답 데이터:", err)
		return
	}

	fmt.Printf("포츈 쿠키를 만들었습니다!\n")
	fmt.Printf("포츈 쿠키 전체 갯수: %d\n", result.AllCount)
	fmt.Printf("%s이(가) 만든 포츈 쿠키 갯수: %d\n", author, result.AuthorCount)
}
