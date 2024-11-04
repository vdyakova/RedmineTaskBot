package authorization

import (
	"fmt"
	_ "github.com/lib/pq"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"strings"
)

func AuthorizationRedmine(login string, pass string, userFirstName string, ch chan<- string) {
	loginURL := os.Getenv("LOGIN_URL")
	targetURL := os.Getenv("TARGET_URL")
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
	var loginForm = &url.Values{}
	parseForm(doc, loginForm)
	loginForm.Set("username", login)
	loginForm.Set("password", pass)

	resp, err = client.PostForm(loginURL, *loginForm)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Login error: ", resp.Status)
	}
	resp, err = client.Get(targetURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	page, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	tasks := parseTasks(page, nameSurname)

	for _, task := range tasks {
		ch <- task
	}
	close(ch)
}

func parseForm(n *html.Node, loginForm *url.Values) {
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
		parseForm(c, loginForm)
	}
}
func parseTasks(n *html.Node, nameSurname string) []string {
	var tasks []string
	if n.Type == html.ElementNode && n.Data == "tr" {
		var taskID, taskStatus, taskDesc, taskAssignedTo string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "td" {
				for _, attr := range c.Attr {
					if attr.Key == "class" {
						switch attr.Val {
						case "assigned_to":
							if c.FirstChild != nil && c.FirstChild.FirstChild != nil {
								taskAssignedTo = getTextContent(c.FirstChild.FirstChild)
							}
						case "id":
							if c.FirstChild != nil && c.FirstChild.FirstChild != nil {
								taskID = getTextContent(c.FirstChild.FirstChild)
							}
						case "status":
							if c.FirstChild != nil {
								taskStatus = getTextContent(c.FirstChild)
							}
						case "subject":
							if c.FirstChild != nil && c.FirstChild.FirstChild != nil {
								taskDesc = getTextContent(c.FirstChild.FirstChild)
							}
						}
					}
				}
			}
		}
		if taskAssignedTo == nameSurname {
			message := fmt.Sprintf("TaskID: %s\nStatus: %s\nDescription: %s\nAssignedTo: %s", taskID, taskStatus, taskDesc, taskAssignedTo)
			tasks = append(tasks, message)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		tasks = append(tasks, parseTasks(c, nameSurname)...)
	}
	return tasks
}
func getTextContent(n *html.Node) string {
	var sb strings.Builder
	html.Render(&sb, n)
	return sb.String()
}
