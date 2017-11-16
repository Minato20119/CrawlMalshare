package CrawlerInsertDB

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"sync"
	"time"
)

const UrlHome = "https://malshare.com/daily/"

var FilterLinkRegex = regexp.MustCompile("href=\"([\\w|.|-]+)/")
var MalshareTxtRegex = regexp.MustCompile("href=\"([\\w|.|-]+)\"")

type Malshare struct {
	Date  string
	Files []string
}

// Get Content File
func getContentUrlText(url string) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("Error getContentUrlText: ")
		return "", err
	}
	defer resp.Body.Close()

	sourcePage, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error getContentUrlText: ")
		return "", err
	}
	return string(sourcePage), err
}

// Regex
func getNameFileTxt(url string) []string {
	result, err := getContentUrlText(url)
	if err != nil {
		fmt.Println(err)
	}
	linkTxt := MalshareTxtRegex.FindAllString(result, -1)
	var nameFileTxt = linkTxt

	if len(linkTxt) > 0 {
		for i := 0; i < len(linkTxt); i++ {
			tempLink := linkTxt[i]

			if len(tempLink) > 6 {
				nameFileTxt[i] = tempLink[6 : len(linkTxt[i])-1]
				nameFileTxt[i] = url + nameFileTxt[i]
			}
		}
	}
	return nameFileTxt
}

func filterLink(url string) []string {
	result, err := getContentUrlText(url)
	if err != nil {
		fmt.Println(err)
	}
	linkDateDirectory := FilterLinkRegex.FindAllString(result, -1)
	var nameDate = linkDateDirectory
	if len(linkDateDirectory) > 0 {
		for i := 0; i < len(linkDateDirectory); i++ {
			tempLink := linkDateDirectory[i]
			if len(tempLink) > 6 {
				nameDate[i] = tempLink[6 : len(linkDateDirectory[i])-1]
			}
		}
	}
	return nameDate
}

func disabled(url string, begin int, end int) {

	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		fmt.Println("Error connect database...")
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("tempHash").C("tempMalshare7")
	fmt.Println("Connected Database with Disable!")

	linkFirst := filterLink(url)
	lenFirst := len(linkFirst)

	for i := 0; i < lenFirst; i++ {
		pathUrl := url + linkFirst[i]
		if linkFirst[i] == "_disabled" {
			linkSecond := filterLink(pathUrl)
			fmt.Println("begin: ", begin, "end: ", end)
			for j := begin; j < end; j++ {
				pathUrlDisabled := pathUrl + "/" + linkSecond[j] + "/"
				fmt.Println(getNameFileTxt(pathUrlDisabled))
				err = c.Insert(&Malshare{linkSecond[j], getNameFileTxt(pathUrlDisabled)})
				checkError(err)
			}
			break
		}
	}
}

func notDisable(url string) {

	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		fmt.Println("Error connect database...")
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("tempHash").C("tempMalshare7")
	fmt.Println("Connected Database with notDisable!")

	linkFirst := filterLink(url)
	lenFirst := len(linkFirst)

	for i := 0; i < lenFirst; i++ {
		if linkFirst[i] == "_disabled" {
			break
		}
		pathUrl := url + linkFirst[i] + "/"
		fmt.Println(getNameFileTxt(pathUrl))
		err = c.Insert(&Malshare{linkFirst[i], getNameFileTxt(pathUrl)})
		checkError(err)
	}

	fmt.Println(getNameFileTxt(url))
	err = c.Insert(&Malshare{"Today", getNameFileTxt(url)})
	checkError(err)

	for i := 0; i < lenFirst; i++ {
		if linkFirst[i] == "archive" {
			fmt.Println(getNameFileTxt(url + "archive/"))
			err = c.Insert(&Malshare{"archive", getNameFileTxt(url + "archive/")})
			break
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var wg sync.WaitGroup
	start := time.Now()
	fmt.Println("Begin program: ", start)
	runtime.GOMAXPROCS(runtime.NumCPU())

	pathUrl := UrlHome + "_disabled"
	linkSecond := filterLink(pathUrl)
	lenSecond := len(linkSecond)
	workers := 20
	temp := lenSecond / workers
	wg.Add(20)
	go notDisable(UrlHome)

	for worker := 0; worker < workers-1; worker++ {
		go disabled(UrlHome, worker*temp, (worker+1)*temp)
		fmt.Println(worker)
	}

	disabled(UrlHome, (workers-1)*temp, lenSecond)
	fmt.Printf("Duration: %s\n", time.Since(start))
}
