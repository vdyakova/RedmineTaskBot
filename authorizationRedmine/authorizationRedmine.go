package authorization

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

var loginURL = "https://redmin.org/login"
var targetURL = "https://redmine..org/projects/..."

func AuthorizationRedmine(login string, pass string, userFirstName string, ch chan<- string) {
	nameSurname := FetchUsernamesFromDB(userFirstName)
	fmt.Println(nameSurname)
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Jar: jar,
	}
	resp, err := client.Get(loginURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	loginForm := url.Values{}

	var parseForm func(*html.Node)
	parseForm = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var name, value string
			for _, attr := range n.Attr {
				if attr.Key == "name" {
					name = attr.Val
				}
				if attr.Key == "value" {
					value = attr.Val
				}
			}
			if name != "" {
				loginForm.Set(name, value)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseForm(c)
		}
	}
	parseForm(doc)
	loginForm.Set("username", login)
	loginForm.Set("password", pass)

	resp, err = client.PostForm(loginURL, loginForm)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Login error: ", resp.Status)
	}
	//parsing html-странцы -> чтобы конкретный пользователь получил свои задачи и статус
	resp, err = client.Get(targetURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	page, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// функция для парсинга страницы
	var parsingPage func(*html.Node)
	parsingPage = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var taskID, taskStatus, taskDesc, taskAssignedTo string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "td" {
					for _, attr := range c.Attr {
						if attr.Key == "class" && attr.Val == "assigned_to" {
							if c.FirstChild != nil && c.FirstChild.FirstChild != nil {
								var sb strings.Builder
								html.Render(&sb, c.FirstChild.FirstChild)
								taskAssignedTo = sb.String()
							}
						}
						if attr.Key == "class" && attr.Val == "id" {
							if c.FirstChild != nil && c.FirstChild.FirstChild != nil {
								var sb strings.Builder
								html.Render(&sb, c.FirstChild.FirstChild)
								taskID = sb.String()
							}
						}
						if attr.Key == "class" && attr.Val == "status" {
							if c.FirstChild != nil {
								var sb strings.Builder
								html.Render(&sb, c.FirstChild)
								taskStatus = sb.String()
							}
						}
						if attr.Key == "class" && attr.Val == "subject" {
							if c.FirstChild != nil && c.FirstChild.FirstChild != nil {
								var sb strings.Builder
								html.Render(&sb, c.FirstChild.FirstChild)
								taskDesc = sb.String()
							}
						}
					}
				}
			}
			if taskAssignedTo == nameSurname {
				fmt.Println(taskID, "статус", taskStatus, "описание", taskDesc)
				message := fmt.Sprintf("TaskID: %s\nStatus: %s\nDescription: %s\nAssignedTo: %s", taskID, taskStatus, taskDesc, taskAssignedTo)
				ch <- message
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parsingPage(c)
		}
	}
	parsingPage(page)
	close(ch)
}

func FetchUsernamesFromDB(tgacc string) string {
	var nameSurname string
	connStr := "user=postgres password = ..  dbname=tgacc sslmode=disable client_encoding=UTF8"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.QueryRow("SELECT namesurname FROM tgacc WHERE tgacc = $1", tgacc).Scan(&nameSurname)
	if err != nil {
		log.Fatal(err)
	}
	return nameSurname
}
