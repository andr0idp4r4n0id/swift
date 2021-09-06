package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
)

var ssti_payloads = []string{
	"{{5430*4999}}",
	"[5430*4999]]",
	"{{5430*4999}}",
	"{{5430*4999}}",
	"<%= 5430*4999 %>",
	"${5430*4999}",
	"${{5430*4999}}",
	"@(5430*4999)",
	"#{5430*4999}",
	"#{ 5430*4999 }",
}

func CheckContains(url_t string) bool {
	re := regexp.MustCompile(`\?\w+=.+`)
	return re.MatchString(url_t)
}

func ExtractHostToPrint(url_t string) string {
	uri, _ := url.Parse(url_t)
	return uri.Host + uri.Path
}

func TestOneByOneSQLi(url_t string, name string, wg *sync.WaitGroup, sem chan bool) {
	defer wg.Done()
	<-sem
	payloads := url.Values{}
	var new_url string
	for _, ssti_payload := range ssti_payloads {
		payloads.Set(name, ssti_payload)
		encoded_payloads := payloads.Encode()
		if CheckContains(url_t) {
			new_url = fmt.Sprintf("%s&%s", url_t, encoded_payloads)
		} else {
			new_url = fmt.Sprintf("%s?%s", url_t, encoded_payloads)
		}
		resp, err := http.Get(new_url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		bodyString := string(bodyBytes)
		pattern, err := regexp.Compile("27144570")
		if err != nil {
			continue
		}
		if pattern.MatchString(bodyString) {
			fmt.Printf("Possible SSTI ---> %s?%s=%s", ExtractHostToPrint(url_t), name, ssti_payload)
		}
	}
}

func main() {
	reader := bufio.NewScanner(os.Stdin)
	var wg sync.WaitGroup
	conc := flag.Int("concurrency", 10, "concurrency level")
	sem := make(chan bool, *conc)
	for reader.Scan() {
		url_t := reader.Text()
		parsedUri, _ := url.Parse(url_t)
		query, _ := url.ParseQuery(parsedUri.RawQuery)
		for name := range query {
			wg.Add(1)
			sem <- true
			go TestOneByOneSQLi(url_t, name, &wg, sem)
		}
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	wg.Wait()
}
