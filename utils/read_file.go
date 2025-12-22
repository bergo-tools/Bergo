package utils

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

type ReadFile struct {
	Path        string
	LineBudget  int
	WithLineNum bool
}

func (r *ReadFile) ReadFile() ([]string, error) {
	file, err := os.Open(r.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var result []string
	lineNum := 1
	for scanner.Scan() {
		if r.LineBudget > 0 && lineNum > r.LineBudget {
			result = append(result, fmt.Sprintf("...content after %d lines are truncated...\n", r.LineBudget))
			break
		}
		line := scanner.Text()
		if r.WithLineNum {
			result = append(result, fmt.Sprintf("%d|%s\n", lineNum, line))
		} else {
			result = append(result, line+"\n")
		}
		lineNum++
	}
	return result, nil
}

func (r *ReadFile) ReadFileTruncated(start int, end int) ([]string, error) {
	//offset and end are line numbers
	file, err := os.Open(r.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var result []string
	lineNum := 1
	if end == 0 {
		end = math.MaxInt
	}
	for scanner.Scan() {
		if lineNum >= start && lineNum <= end {
			line := scanner.Text()
			if r.WithLineNum {
				result = append(result, fmt.Sprintf("%d|%s\n", lineNum, line))
			} else {
				result = append(result, line+"\n")
			}
		}
		lineNum++
	}
	return result, nil
}

func (r *ReadFile) ReadFileWhole() (string, error) {
	data, err := os.ReadFile(r.Path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
