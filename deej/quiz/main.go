package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Create a struct for holding the problem records
type probRecord struct {
	q string
	a string
}

// instantiate the variables we'll be using
var (
	filename    = flag.String("file", "./problems.csv", "CSV formatted file containing question and answer pairs.")
	shuffleBool = flag.Bool("shuffle", false, "Shuffle the lines from the quiz file.")
	timeout     = flag.Int("timeout", 30, "Set a timeout for the quiz.")
	scoreChan   = make(chan int, 1)
)

// shuffle a slice of probRecords
func shuffleSlice(slice []probRecord) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// clean up answers
func cleanString(s string) (cleanString string) {
	return strings.ToLower(strings.TrimSpace(s))
}

// convert a csv file into a slice of probRecords
func sliceCSVFile(filename string) []probRecord {
	f, err := os.Open(filename)
	defer f.Close()

	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(f)
	o := []probRecord{}

	// parse the csv reader into the slice
	for {
		line, err := r.Read()
		// catch the EOF error to stop looping
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// create probRecord to contain the line
		var record probRecord
		record.q, record.a = line[0], line[1]

		o = append(o, record)
	}

	return o
}

func timeitout(timeout int, len int) {
	duration := time.Duration(timeout)
	time.Sleep(time.Second * duration)
	fmt.Printf("\nYou've run out of time.\n")
	// read current score from reader input loop
	timeoutScore := <-scoreChan
	printScore(timeoutScore, len)
	os.Exit(0)
}

// loop through slice of probRecords, prompt for answers, track score
func runQuiz(prs []probRecord) (finalScore int) {
	var (
		score     int
		userInput string
	)

	for _, problem := range prs {

		fmt.Printf(problem.q + " = ")
		fmt.Scanln(&userInput)

		if cleanString(problem.a) == cleanString(userInput) {
			score++
			// clear current score from channel
			for len(scoreChan) > 0 {
				<-scoreChan
			}
			// update channel with new score
			scoreChan <- score
		}
	}

	return score
}

// print the final score
func printScore(score int, len int) {
	fmt.Printf("Your final score was %d out of %d.\n", score, len)
}

func main() {
	// read command-line flags
	flag.Parse()

	// create probRecords slice from the csv file
	prs := sliceCSVFile(*filename)

	// shuffle the probRecords if requested
	if *shuffleBool {
		shuffleSlice(prs)
	}

	// inform user of timeout and prompt to begin quiz
	fmt.Printf("You have %d seconds to complete the quiz.\n", *timeout)
	fmt.Println("Press enter to begin.")
	fmt.Scanln()

	// start routine to enforce quiz time limit
	go timeitout(*timeout, len(prs))

	finalScore := runQuiz(prs)

	printScore(finalScore, len(prs))
}
