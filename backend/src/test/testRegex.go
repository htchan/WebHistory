package main

import (
	"time"
	"regexp"
	"fmt"
	"io/ioutil"
	"strings"
)

func testRegex(s string, n int) {
	start := time.Now()
	var resultList []string
	for i := 0; i < n; i++ {
		re := regexp.MustCompile("<.*?>")
		result := string(re.ReplaceAll([]byte(s), []byte("\n")))
		resultList = strings.Split(result, "\n")
	}
	end := time.Now()
	// fmt.Println(resultList[0:100])
	fmt.Println("result length:", len(resultList))
	fmt.Println("time taken:", (end.UnixNano() - start.UnixNano()) / int64(time.Millisecond))
}

func testLoop(s string, n int) {
	start := time.Now()
	resultList := make([]string, 1)
	for i := 0; i < n; i++ {
		ignore := false
		tempResult := ""//make([]byte, 0)
		for _, b := range(s) {
			if b == rune('<') {
				ignore = true
			}
			if !ignore {
				// tempResult = append(tempResult, byte(b))
				tempResult += string(b)
			}
			if b == rune('>') {
				ignore = false
				resultList[len(resultList) - 1] = string(tempResult)
				resultList = append(resultList, "")
			}
		}
	}
	end := time.Now()
	// fmt.Println(resultList[0:100])
	fmt.Println("result length:", len(resultList))
	fmt.Println("time taken:", (end.UnixNano() - start.UnixNano()) / int64(time.Millisecond))
}

func main() {
	data, err := ioutil.ReadFile("./data.html")
	if err != nil { panic(err) }
	re := regexp.MustCompile("(<script.*?/script>|<style.*?/style>|<path.*?/path>)")
	dataStr := strings.ReplaceAll(strings.ReplaceAll(string(data), "\n", ""), "\r", "")
	data = re.ReplaceAll([]byte(dataStr), []byte("<script/>"))
	dataStr = string(data)

	trials := 1
	fmt.Println("regex solution :")
	testRegex(dataStr, trials)
	fmt.Println("loop solution :")
	testLoop(dataStr, trials)
	fmt.Println("***regexp is better***")
}