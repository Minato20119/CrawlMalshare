package CrawlerInsertDB

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

const URL_HOME = "https://malshare.com/daily/"

var FILTER_LINK_REGEX = regexp.MustCompile("href=\"([\\w|.|-]+)/")
var MALSHARE_TXT_REGEX = regexp.MustCompile("href=\"([\\w|.|-]+)\"")

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
	linkTxt := MALSHARE_TXT_REGEX.FindAllString(result, -1)
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
	linkDateDirectory := FILTER_LINK_REGEX.FindAllString(result, -1)
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

func insertToMongo(url string) {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		fmt.Println("Error connect database...")
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("Hash").C("malshare")
	fmt.Println("Connected Database!")

	linkFirst := filterLink(url)
	lenFirst := len(linkFirst)

	for i := 0; i < lenFirst; i++ {
		pathUrl := url + linkFirst[i]
		if linkFirst[i] == "_disabled" {
			linkSecond := filterLink(pathUrl)
			lenSecond := len(linkSecond)
			for j := 0; j < lenSecond; j++ {
				pathUrlDisabled := pathUrl + "/" + linkSecond[j] + "/"
				fmt.Println(getNameFileTxt(pathUrlDisabled))
				err = c.Insert(&Malshare{linkSecond[j], getNameFileTxt(pathUrlDisabled)})
				checkError(err)
			}
			break
		}
	}

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
	start := time.Now()
	fmt.Println("Begin program: ", start)
	insertToMongo(URL_HOME)
	fmt.Printf("Duration: %s\n", time.Since(start))
}
