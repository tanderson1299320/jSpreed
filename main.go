package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var startTime time.Time
var wordCount int

func init() {
	startTime = time.Now()
	wordCount = 0
}

func calculateWPM() {
	elapsedTime := time.Since(startTime)
	minutes := elapsedTime.Minutes()
	if minutes > 0 {
		wpm := float64(wordCount) / minutes
		fmt.Printf("
Words Per Minute: %.2f
", wpm)
	}
}


func main() {
	var (
		wpm                               int
		file                              string
		keepSpecialChars, clearScreenEach bool
	)

	flag.IntVar(&wpm, "wpm", 300, "words per minute")
	flag.StringVar(&file, "file", "", "file to read")
	flag.BoolVar(&keepSpecialChars, "k", false, "keep special characters")
	flag.BoolVar(&clearScreenEach, "c", false, "clear screen each time")

	flag.Parse()

	if file == "" {
		fmt.Println("Please provide a file to read with the -file flag.")
		return
	}

	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file: %v
", err)
		return
	}

	text := string(content)
	if !keepSpecialChars {
		text = strings.ReplaceAll(text, "
", " ")
		text = strings.ReplaceAll(text, "", " ")
		text = strings.ReplaceAll(text, "	", " ")
		// Add more replacements if needed
	}

	words := strings.Fields(text)
	delay := time.Minute / time.Duration(wpm)

	for _, word := range words {
		if clearScreenEach {
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
		fmt.Printf("%s", strings.Repeat(" ", 50)) // Clear the line
		fmt.Printf("%s", word)
		wordCount++
		time.Sleep(delay)
	}

	fmt.Println() // Newline at the end
	calculateWPM()
}
