package utils

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
)

type Edit struct {
	Path string
}

func (e *Edit) EditWholeFile(content string) error {
	f, err := os.OpenFile(e.Path, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

var ErrEditMultipleMatch = errors.New("multiple match in file, please expand code block in your search to match more precisely")
var ErrEditNoMatch = errors.New("no match in file, please check your tool call")
var ErrSourceFileEmpty = errors.New("source file is empty, please check your tool call")

func (e *Edit) cutLines(content string) []string {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
func (e *Edit) cutLinesWithoutEmpty(content string) []string {
	lines := e.cutLines(content)
	filteredLines := make([]string, 0)
	for _, line := range lines {
		if line != "" {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines
}

func (e *Edit) EditByDiff(search string, replace string) error {
	if search == "" {
		return ErrEditNoMatch
	}
	// 读取文件原始内容
	rf := &ReadFile{
		Path:        e.Path,
		LineBudget:  math.MaxInt32,
		WithLineNum: false,
	}
	allLines, err := rf.ReadFile()
	if err != nil {
		return err
	}
	if len(allLines) == 0 {
		return ErrSourceFileEmpty
	}
	searchLines := e.cutLinesWithoutEmpty(search)
	replaceLines := e.cutLines(replace)
	searchLineIdx := 0
	success := 0
	endIdx := 0
	startIdx := 0
	for i := 0; i < len(allLines); i++ {
		trimLine := strings.TrimSpace(allLines[i])
		if trimLine == "" {
			continue
		}
		if trimLine == strings.TrimSpace(searchLines[searchLineIdx]) {
			searchLineIdx++
			if searchLineIdx == 1 {
				startIdx = i
			}
		} else {
			searchLineIdx = 0
		}
		if searchLineIdx == len(searchLines) {
			success += 1
			endIdx = i
			searchLineIdx = 0
		}
	}
	if success == 0 {
		return ErrEditNoMatch
	}
	if success > 1 {
		return ErrEditMultipleMatch
	}
	buf := bytes.NewBuffer(nil)
	for i := 0; i < startIdx; i++ {
		buf.WriteString(allLines[i])
	}
	for i := 0; i < len(replaceLines); i++ {
		buf.WriteString(replaceLines[i])
		buf.WriteString("\n")
	}
	for i := endIdx + 1; i < len(allLines); i++ {
		buf.WriteString(allLines[i])
	}
	return e.EditWholeFile(buf.String())
}

func (e *Edit) EditInplace(start int, end int, start_line, end_line, replace string) error {
	if start >= end {
		return fmt.Errorf("invalid params, start line %d must be less than end line %d", start, end)
	}
	if start != 0 && end == 0 {
		end = start
		end_line = start_line
	}
	// 读取文件原始内容
	rf := &ReadFile{
		Path:        e.Path,
		LineBudget:  math.MaxInt32,
		WithLineNum: false,
	}
	allLines, err := rf.ReadFile()
	if err != nil {
		return err
	}
	if end > len(allLines) {
		return fmt.Errorf("invalid params, end line %d must be less than file line count %d", end, len(allLines))
	}
	start_line = strings.TrimSpace(start_line)
	end_line = strings.TrimSpace(end_line)
	if strings.TrimSpace(allLines[start-1]) != start_line {
		return fmt.Errorf("invalid params, start_line [%s] != actual line [%s]", start_line, allLines[start-1])
	}
	if strings.TrimSpace(allLines[end-1]) != end_line {
		return fmt.Errorf("invalid params, end_line [%s] != actual line [%s]", end_line, allLines[end-1])
	}
	buf := bytes.NewBuffer(nil)
	edited := false
	for i := 0; i < len(allLines); i++ {
		lineCount := i + 1
		if lineCount >= start && lineCount <= end {
			if !edited {
				buf.WriteString(replace)
				edited = true
			}
			continue
		}
		buf.WriteString(allLines[i])
	}

	return e.EditWholeFile(buf.String())
}
