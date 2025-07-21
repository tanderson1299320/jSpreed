package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
	"golang.org/x/term"
)

const (
	linesPerPage = 20
)

type BionicReader struct {
	lines        []string
	currentPage  int
	totalPages   int
	filename     string
	termWidth    int
}

func NewBionicReader(filename string) (*BionicReader, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80 // fallback
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		text := scanner.Text()
		wrappedLines := wrapText(text, width)
		lines = append(lines, wrappedLines...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	totalPages := (len(lines) + linesPerPage - 1) / linesPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	return &BionicReader{
		lines:       lines,
		currentPage: 0,
		totalPages:  totalPages,
		filename:    filename,
		termWidth:   width,
	}, nil
}

func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func bionicFormat(text string) string {
	words := strings.Fields(text)
	var formatted []string

	for _, word := range words {
		formatted = append(formatted, formatWord(word))
	}

	return strings.Join(formatted, " ")
}

func formatWord(word string) string {
	if len(word) <= 1 {
		return "\033[1m" + word + "\033[0m"
	}

	runes := []rune(word)
	var letterCount int
	var firstLetterIdx, lastLetterIdx int

	for i, r := range runes {
		if unicode.IsLetter(r) {
			if letterCount == 0 {
				firstLetterIdx = i
			}
			lastLetterIdx = i
			letterCount++
		}
	}

	if letterCount == 0 {
		return word
	}

	boldCount := (letterCount + 1) / 2
	if boldCount < 1 {
		boldCount = 1
	}

	result := string(runes[:firstLetterIdx])
	lettersSeen := 0
	
	for i := firstLetterIdx; i <= lastLetterIdx; i++ {
		if unicode.IsLetter(runes[i]) {
			if lettersSeen < boldCount {
				result += "\033[1m" + string(runes[i]) + "\033[0m"
			} else {
				result += string(runes[i])
			}
			lettersSeen++
		} else {
			result += string(runes[i])
		}
	}

	if lastLetterIdx < len(runes)-1 {
		result += string(runes[lastLetterIdx+1:])
	}

	return result
}

func (br *BionicReader) displayPage() {
	fmt.Print("\033[2J\033[H")

	start := br.currentPage * linesPerPage
	end := start + linesPerPage
	if end > len(br.lines) {
		end = len(br.lines)
	}

	for i := start; i < end; i++ {
		fmt.Print(bionicFormat(br.lines[i]) + "\r\n")
	}

	for i := end - start; i < linesPerPage; i++ {
		fmt.Print("\r\n")
	}

	progress := fmt.Sprintf("Page %d/%d (%.1f%%) - %s - Press Enter/Space for next page, 'q' to quit",
		br.currentPage+1, br.totalPages,
		float64(br.currentPage+1)/float64(br.totalPages)*100,
		br.filename)
	
	fmt.Printf("\033[%d;1H\033[7m%s\033[0m", linesPerPage+2, 
		fmt.Sprintf("%-*s", br.termWidth, progress))
}

func (br *BionicReader) run() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	
	for {
		br.displayPage()
		
		var b [1]byte
		os.Stdin.Read(b[:])
		
		if b[0] == 'q' || b[0] == 'Q' || b[0] == 3 { // 3 is Ctrl+C
			break
		}
		
		if b[0] == '\n' || b[0] == '\r' || b[0] == ' ' {
			if br.currentPage < br.totalPages-1 {
				br.currentPage++
			} else {
				fmt.Print("\r\nEnd of file reached. Press any key to continue or 'q' to quit.")
			}
		}
	}
	
	fmt.Print("\033[2J\033[H")
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <filename>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	reader, err := NewBionicReader(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	reader.run()
}