package common

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Get last n rows from a file.
// the order is reversed as it reads bottom up. So The last line in the file will the first entry in the list
// Last in first out in relation to the file being read.
func Tail(path string, n int64) []string {
	file, err := os.Open(path)
	if err != nil {
		log.Println("util.tail: ", err)
		return nil
	}
	defer file.Close()

	var linesFound []string
	var lineBuffer []byte

	_, _ = file.Seek(0, 2) // Seek to the end of the file.
	stat, _ := os.Stat(path)
	size := stat.Size()
	index := 1

	// Read each byte and add it to a list until '\n' is found, that is a 'line'
	for len(linesFound) < int(n) {
		buffer := make([]byte, 1) // Buffer of 1 as we only want to read 1 byte.
		offset := size - int64(index)
		if offset < 0 {
			break
		}
		_, err := file.ReadAt(buffer, offset)
		if err != nil {
			log.Println("Utils.Tail: ", err)
			return nil
		}
		index++
		b := buffer[0]

		if b == byte(0x0a) {
			if len(lineBuffer) > 1 { // Only add line to list if it's not empty.
				linesFound = append(linesFound, reverse(string(lineBuffer)))
			}
			lineBuffer = nil
		} else {
			lineBuffer = append(lineBuffer, b)
		}

	}

	return linesFound
}

// Reverse String.
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Convert string to float but return 0 when error instead of panic.
func String2float(s string) float64 {
	if f, err := strconv.ParseFloat(s, 32); err == nil {
		return f
	}
	return 0
}

// Convert CSV to a list of maps with first row as headers.
func CSVToMap(reader io.Reader) []map[string]string {
	r := csv.NewReader(reader)
	var rows []map[string]string
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("ps_client.CSVToMap: ", err)
			return nil
		}
		if header == nil {
			header = record
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = record[i]
			}
			rows = append(rows, dict)
		}
	}
	return rows
}

// Replace time in the header row to string "time" so it's easier to get.
func changeTimeHeader(metrics string) string {
	re := regexp.MustCompile(`^(.*?)(?:[01]\d|2[0123]):(?:[012345]\d):(?:[012345]\d)(.*)`)
	return re.ReplaceAllString(metrics, "${1}time$2")
}

// Function to check if a string is present in a slice.
func Contains(itemList []string, str string) bool {
	for _, a := range itemList {
		if a == str {
			return true
		}
	}
	return false
}

// Change header time label, and get mapped csv rows.
func GetRows(data string) []map[string]string {
	data = changeTimeHeader(data)
	stringReader := strings.NewReader(data)
	return CSVToMap(stringReader)
}
