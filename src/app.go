package main

import (
    "fmt"
    "net/http"
    "time"
    "strings"
    "io/ioutil"
    "bytes"
    "net/url"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/PuerkitoBio/goquery"
)

type SearchResult struct {
    Title string `json:"title"`
    Link  string `json:"link"`
}

func handleSearch(c *gin.Context) {
    // Get the date input from the request URL
    date := c.Query("date")
    page, err := strconv.Atoi(c.Query("page"))
    if err != nil {
        page = 1
    }

    start := (page - 1) * 10

    // Parse the date input into a time.Time object
    t, err := time.Parse("01-02-2006", date)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use MM-DD-YYYY"})
        return
    }

    // Build the search query
    query := "Florida Man "
    formattedDate := t.Format("01-31")
    startQuery := "&start=" + strconv.Itoa(start)

    path := "https://www.google.com/search?q=" + url.QueryEscape(query) + url.QueryEscape(formattedDate) + startQuery;
    resp, err := http.Get(path)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve search results"})
        return
    }
    defer resp.Body.Close() // Tells go to execute after the parent function returns

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read search results"})
        return
    }

    // Scrape the search results
    doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body));
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scrape search results"})
        return
    }


    // Parse the search result titles and links into an array of SearchResult objects
    results := make([] SearchResult, 0)
    seenTitles := make(map[string]bool)

    fmt.Println("Scraping URL: ", path)
    // Extract the search result headlines and links
    doc.Find("div").Each(func(i int, s *goquery.Selection) {
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
                results = append(results, SearchResult{Title: title, Link: link[7:]})
                seenTitles[title] = true
            }
        }
    })
    // Return the search results as JSON
    c.JSON(http.StatusOK, results)
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


