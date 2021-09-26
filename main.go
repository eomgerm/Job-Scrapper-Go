package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
	"os"
	"encoding/csv"
)

var baseUrl string = "https://kr.indeed.com/jobs?q=python&limit=50"

type jobDetail struct {
	id string
	title string
	company string
	location string
	salary string
	summary string
}

func main(){
	totalPages := getPages()
	fmt.Println("Total Page: " + strconv.Itoa(totalPages))
	
	var jobs []jobDetail
	for i := 0; i < totalPages; i++ {
		extractedJobs := getPage(i)
		jobs = append(jobs, extractedJobs...)
	}
	
	writeJobs(jobs)
}

func getPages() int {
	pages := 0
	res, err := http.Get(baseUrl)
	checkErr(err)
	checkCode(res)
	
	defer res.Body.Close()
	
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection){
		pages = s.Find("a").Length()
	})
	
	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request Failed with Status Code: ", res.StatusCode)
	}
}

func getPage(page int) []jobDetail {
	pageUrl := baseUrl + "&start=" + strconv.Itoa(page * 50)
	fmt.Println("Requesting " + pageUrl)
	
	res, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(res)
	
	defer res.Body.Close()
	
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	
	c := make(chan jobDetail)
	
	var jobs []jobDetail
	cards := doc.Find("#mosaic-provider-jobcards>a")
	cards.Each(func(i int, card *goquery.Selection){
		go extractJob(card, c)
	})
	
	for i := 0; i < cards.Length(); i++ {
		job := <-c
		jobs := append(jobs, job)
	}
	
	return jobs
}

func extractJob(card *goquery.Selection, c chan<- jobDetail) {
	id, _ := card.Attr("data-jk")
	id = cleanString(id)
	title := cleanString(card.Find(".jobTitle>span").Text())
	company := cleanString(card.Find(".companyName").Text())
	location := cleanString(card.Find(".companyLocation").Text())
	salary := cleanString(card.Find(".salary-snippet").Text())
	summary := cleanString(card.Find(".job-snippet").Text())
	c <- jobDetail {
		id: id,
		title: title,
		company: company,
		location: location,
		salary: salary,
		summary: summary }
}

func cleanString(str string) string{
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func writeJobs(jobs []jobDetail) {
	file, err := os.Create("jobs.csv")
	checkErr(err)
	
	w := csv.NewWriter(file)
	defer w.Flush()
	
	headers := []string{"ID", "Title", "Company", "Location", "Salary", "Summary"}
	
	wErr := w.Write(headers)
	checkErr(wErr)
	
	for _, job := range jobs {
		jobSlice := []string{"https://kr.indeed.com/%EC%B1%84%EC%9A%A9%EB%B3%B4%EA%B8%B0?jk=" + job.id, job.title, job.company, job.location, job.salary, job.summary}
		
		jWErr := w.Write(jobSlice)
		checkErr(jWErr)
	}
	fmt.Println("Writing done, ", len(jobs), " jobs extracted.")
}