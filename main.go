//
// @bp0lr - 19/09/2022
//

package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	pb "github.com/cheggaaa/pb/v3"
	flag "github.com/spf13/pflag"
)

var (
	workersArg    int
	yoururlArg    string
	serverurlArg  string
	needleArg     string
	proxyArg      string
	outputFileArg string
	verboseArg    bool
	pbArg         bool
)

var bar *pb.ProgressBar

func main() {

	flag.IntVarP(&workersArg, "workers", "w", 50, "Workers amount")
	flag.StringVarP(&yoururlArg, "fronturl", "f", "", "your host")
	flag.StringVarP(&serverurlArg, "testurl", "u", "", "host to test")
	flag.StringVarP(&needleArg, "needle", "n", "", "the string to confirm that fronting works")
	flag.BoolVarP(&verboseArg, "verbose", "v", false, "Display extra info about what is going on")
	flag.StringVarP(&proxyArg, "proxy", "p", "", "Add a HTTP proxy")
	flag.StringVarP(&outputFileArg, "output", "o", "", "Output file to save the results to")
	flag.BoolVar(&pbArg, "use-pb", false, "use a progress bar")

	flag.Parse()

	//concurrency
	workers := 50
	if workersArg > 0 && workersArg < 100 {
		workers = workersArg
	}

	client := newClient(proxyArg)

	var outputFile *os.File
	var err0 error
	if outputFileArg != "" {
		outputFile, err0 = os.OpenFile(outputFileArg, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err0 != nil {
			fmt.Printf("cannot write %s: %s", outputFileArg, err0.Error())
			return
		}

		defer outputFile.Close()
	}

	yourServer, err := url.ParseRequestURI("https://" + yoururlArg)
	if err != nil {
		if verboseArg {
			fmt.Printf("[-] Invalid fronting url: %s\n", yoururlArg)
		}
		return
	}

	var WorksToDo []string

	///////////////////////////////
	// generate taks list
	///////////////////////////////
	var jobs []string

	if len(serverurlArg) < 1 {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			jobs = append(jobs, sc.Text())
		}
	} else {
		jobs = append(jobs, serverurlArg)
	}

	targetDomains := make(chan string)
	var wg sync.WaitGroup
	var mu = &sync.Mutex{}

	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			for task := range targetDomains {
				serverURLTest, err := url.ParseRequestURI("https://" + task)
				if err != nil {
					if verboseArg {
						fmt.Printf("[-] Invalid url: %s\n", task)
					}
					continue
				}
				mu.Lock()
				WorksToDo = append(WorksToDo, serverURLTest.Host)
				mu.Unlock()

			}
			wg.Done()
		}()
	}

	for _, line := range jobs {
		targetDomains <- line
	}

	close(targetDomains)
	wg.Wait()

	///////////////////////////////
	// Lets process the task list
	///////////////////////////////
	if verboseArg {
		fmt.Printf("Works to do: %v\n", len(WorksToDo))
	}

	if pbArg {
		tmpl := `{{ white "Mutations:" }} {{counters . | red}}  {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} [{{speed . | green }}] [{{rtime . "ETA %s"}}]`
		bar = pb.ProgressBarTemplate(tmpl).Start(len(WorksToDo))
		//bar = pb.Full.Start(len(GlobalStats.WorksToDo))
		bar.SetWidth(100)
	}

	tasks := make(chan string)
	var wg2 sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg2.Add(1)
		go func() {
			for task := range tasks {
				processRequest(task, yourServer, client, outputFile, needleArg)
			}
			wg2.Done()
		}()
	}

	for _, line := range WorksToDo {
		tasks <- line
	}

	close(tasks)
	wg2.Wait()

	if pbArg {
		bar.Finish()
	}
}

func processRequest(remoteServer string, frontingServer *url.URL, client *http.Client, outputFile *os.File, needle string) {

	if verboseArg && !pbArg {
		fmt.Printf("[+] Testing: %v => %v\n", remoteServer, frontingServer.String())
	}

	//check read
	/////////////////////////////////////////////////////////////////////////
	res, err := check(remoteServer, frontingServer, client, needle)

	if err != nil && verboseArg && !pbArg {
		fmt.Printf("[-] Error: %v [%v]\n", err, remoteServer)
	} else if res {
		if outputFileArg != "" {
			outputFile.WriteString(remoteServer + "\n")
		}

		if !pbArg {
			fmt.Printf("[+] %v\n", remoteServer)
		}

	} else {
		if verboseArg && !pbArg {
			fmt.Printf("[-] %v\n", remoteServer)
		}
	}

	if pbArg {
		bar.Increment()
	}
}

func newClient(proxy string) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:    30,
		IdleConnTimeout: time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout: time.Second * 10,
		}).DialContext,
	}

	if proxy != "" {
		if p, err := url.Parse(proxy); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return client
}

func getUserAgent() string {
	payload := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 8_4_1 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12H321 Safari/600.1.4",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (iPad; CPU OS 7_1_2 like Mac OS X) AppleWebKit/537.51.2 (KHTML, like Gecko) Version/7.0 Mobile/11D257 Safari/9537.53",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(payload))

	pick := payload[randomIndex]

	return pick
}

//CheckRead desc
func check(remoteServer string, frontingServer *url.URL, client *http.Client, needle string) (bool, error) {

	var urlLocal string
	var req *http.Request
	var err error

	form := url.Values{}
	form.Add("op", "d3bug")

	urlLocal = "https://" + remoteServer

	req, err = http.NewRequest("POST", urlLocal, strings.NewReader(form.Encode()))
	req.Host = frontingServer.Host
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// send the request
	resp, err := client.Do(req)

	if err != nil {
		if verboseArg {
			fmt.Printf("[-] Error: %v\n", err)
		}
		return false, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, nil
	}

	if bytes.Contains(body, []byte(needle)) {
		return true, nil
	}

	return false, nil
}
