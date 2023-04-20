package main

import (
    "log"
    "fmt"
    "net/http"
    "time"
    "strings"
    "io/ioutil"
    "bytes"
    "net/url"
    "strconv"

    "github.com/PuerkitoBio/goquery"
    "github.com/gin-gonic/gin"
)

type SearchResult struct {
    Title string `json:"title"`
    Link  string `json:"link"`
}


func search(
    query string, 
    formattedDate string, 
    start int, 
    results * [] SearchResult, 
    done chan bool, 
    stop chan bool,
) {
    path := "https://www.google.com/search?q=" + 
        url.QueryEscape(query) + 
        url.QueryEscape(formattedDate) + 
        ("&start=" + strconv.Itoa(start))

    // Create a HTTP client with custom User-Agent header
    client := &http.Client{}
    req, err := http.NewRequest("GET", path, nil)
    if err != nil {
        log.Fatal("Failed to scrape search results: ", err)
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
    resp, err := client.Do(req)
    if err != nil {
        log.Fatal("Failed to scrape search results: ", err)
    }

    defer resp.Body.Close() // Tells go to execute after the parent function returns

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal("Failed to read search results: ", err)
    }

    // Scrape the search results
    doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
    if err != nil {
        log.Fatal("Failed to scrape search results: ", err)
    }

    // Parse the search result titles and links into an array of SearchResult objects
    seenTitles := make(map[string]bool)

    fmt.Println("Scraping URL: ", path)
    // Extract the search result headlines and links
    count := 0
    doc.Find("div").Each(func(i int, s * goquery.Selection) {
        title := s.Find("h3").Text()
        link, exists := s.Find("a").Attr("href")
        index := strings.Index(link, "/&sa")
        if index == -1 {
            index = strings.Index(link, "&sa")
        
        }
        if index != -1 {
            link = link[:index]
        }

        if exists && strings.Contains(strings.ToLower(title), "florida man") && strings.Contains(link, "https") {
            if (!seenTitles[title]) {
                * results = append(* results, SearchResult { Title: title, Link: link[8:] })
                seenTitles[title] = true
                count += 1
            }
        }
    })

    if (count == 0) {
        stop <- true
    }  
    
    done <- true
}

func handleSearch(c *gin.Context) {
    // Get the date input from the request URL
    date := c.Query("date")

    // Parse the date input into a time.Time object
    t, err := time.Parse("01-02-2006", date)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use MM-DD-YYYY"})
        return
    }

    // Build the search query
    query := "Florida Man "
    formattedDate := t.Format("01-31")

    var allResults [] SearchResult
    keepGoing := true
    for (keepGoing) {
        var pageResults []SearchResult
        done := make(chan bool)
        stop := make(chan bool)
        for i := 0; i < 5; i++ {
            go search(
                query, 
                formattedDate, 
                i * 10, 
                &pageResults, 
                done, 
                stop,
            )
        }

        
        for i := 0; i < 5; i++ {
            select {
                case <-stop:
                    keepGoing = false 
                    fmt.Println("Last batch")
                case <-done:
                    continue;
            }
        }

        if len(pageResults) == 0 {
            break
        }
        allResults = append(allResults, pageResults...)
    }

    // Return the search results as JSON
    c.JSON(http.StatusOK, allResults)
}

func main() {
    // Create the Gin router
    router := gin.Default()

    // Register the search endpoint
    router.GET("/v1/search", handleSearch)

    // Start the HTTP server
    if err := router.Run(":8080"); err != nil {
        panic(err)
    }
}


