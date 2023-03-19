package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"

	vision "rita-go/parsers"
	"rita-go/structs"
	"time"

	"github.com/gin-gonic/gin"
)

func parseAnnotations(c *gin.Context) {
	jsonData, _ := ioutil.ReadAll(c.Request.Body)

	var response structs.VisionApiResponse

	json.Unmarshal(jsonData, &response)
	var start = time.Now()
	var wordsWithCoords = vision.GetWordsWithCoords(response)
	var wordMatches = vision.GetMatches(wordsWithCoords)
	var lines = vision.BuildLines(wordMatches, wordsWithCoords)
	var end = time.Now()

	keys := make([]int, 0, len(lines))
	for k := range lines {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	var message string = ""

	for _, k := range keys {
		message = message + lines[k] + "\n"
	}

	fmt.Println(end.Sub(start))

	c.JSON(200, gin.H{
		"text": message,
	})
}

func main() {
	r := gin.Default()
	r.POST("/parseAnnotations", parseAnnotations)
	r.Run() // listen and serve on 0.0.0.0:8080
}
