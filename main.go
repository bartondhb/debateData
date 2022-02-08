package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Debate struct {
	Date       string
	Candidates []Candidate
}

type Candidate struct {
	Name       string
	IssueCount map[string]int
}

// To execute this code, type `go run main.go` in a terminal
func main() {

	csvFile, err := readCsv("./debate_data.csv")

	if err != nil {
		panic(err)
	}

	debates, err := parseCsvData(csvFile)

	if err != nil {
		panic(err)
	}

	summary, err := summarize(&debates)

	if err != nil {
		panic(err)
	}

	if err = writeCsv("./output.csv", summary); err != nil {
		panic(err)
	}

}

// summarize function summarizes parsed input CSV data
func summarize(debates *[]Debate) ([][]string, error) {

	// Create a slice of string slices to be used by the CSV Writer
	var rows [][]string

	sortedIssues := getIssues(debates)

	sort.Sort(sort.Reverse(sort.StringSlice(sortedIssues)))
	// Build the header based on collection of issues discussed in each debate
	header := append([]string{"Date", "Candidate"}, sortedIssues...)

	// add the header to the CSV
	rows = append(rows, header)

	// iterate the debates and each candidate
	for _, debate := range *debates {

		for _, candidate := range debate.Candidates {

			// create a new row for each candidate
			var row = make([]string, len(header))

			for hk, h := range header {

				switch h {
				case "Date":
					row[hk] = debate.Date
				case "Candidate":
					row[hk] = candidate.Name
				default:
					row[hk] = strconv.Itoa(candidate.IssueCount[h])
				}
			}

			// add each row to the CSV
			rows = append(rows, row)

		}

	}

	// Create a final row -- this is used to summarize each issue category
	var finalRow = make([]string, len(header))

	finalRow[1] = "Total"

	// Start 2 columns in, because the first two columns are date and candidate
	// Get the length of the header row so you know how many columns to expect
	for colNum := 2; colNum < len(rows[0]); colNum++ {
		var total = 0

		// Start 1 row down, because the first row is the header
		for rowNum := 1; rowNum < len(rows); rowNum++ {
			val, err := strconv.Atoi(rows[rowNum][colNum])

			if err != nil {
				return nil, fmt.Errorf("data integrity error: %+v", err)
			}

			// Calculate the running total for each row in the given column
			total += val
		}

		// Add the total to the final row in the appropriate column
		finalRow[colNum] = strconv.Itoa(total)
	}

	rows = append(rows, finalRow)

	return rows, nil

}

// getIssues returns a deduplicated list of the issues discussed during the debates
func getIssues(debates *[]Debate) []string {

	var issues = make(map[string]interface{})

	// Iterate all debates and candidates to get a unique list of Issues
	for _, debate := range *debates {
		for _, candidate := range debate.Candidates {
			for issue := range candidate.IssueCount {
				issues[issue] = nil
			}
		}
	}

	var issueSlice []string

	// Convert the map to a slice and return it
	for k, _ := range issues {
		issueSlice = append(issueSlice, k)
	}

	return issueSlice
}

// Take CSV data and convert it to a native data structure
func parseCsvData(data [][]string) ([]Debate, error) {

	var debates = make([]Debate, 0)

	// Create a map of the column indices for each Candidate. This is necessary because each Candidate has data
	// across multiple columns with different naming patterns for each debate round ([1], [2], [3], etc)
	indexMap := make(map[string][]int)

	for k, v := range data[0] {
		sanitizedValue := sanitizeColumnName(v)
		indexMap[sanitizedValue] = append(indexMap[sanitizedValue], k)
	}

	// Iterate the raw CSV data starting with index 1 to skip the header row
	for _, debateData := range data[1:] {

		// Create an instance of Debate to store data about the debate
		var debate Debate

		// Iterate the indexMap so we can determine which columns contain which data.
		for rowKey, index := range indexMap {
			switch true {
			// In the case of the date, we are only expecting one column
			case strings.Contains(rowKey, "Date"):
				if len(index) > 1 {
					return nil, fmt.Errorf("data integrity error: the source data contains more than one date column")
				}
				debate.Date = debateData[index[0]]
			// The rest of the columns are Candidate data
			default:
				var candidate Candidate
				candidate.Name = rowKey
				candidate.IssueCount = make(map[string]int)

				for _, indexVal := range index {

					// Here we take data from each Candidate cell, split it by the comma, and remove up any whitespace
					// to get a clean issue name
					issues := strings.Split(debateData[indexVal], ",")

					for _, issue := range issues {
						issue = strings.TrimSpace(issue)

						// handle empty cells
						if issue != "" {

							if _, exists := candidate.IssueCount[issue]; !exists {
								candidate.IssueCount[issue] = 1
							} else {
								candidate.IssueCount[issue]++
							}
						}
					}

				}

				// Add the candidate to the debate
				debate.Candidates = append(debate.Candidates, candidate)
			}

		}

		// Add the debate to the debates slice
		debates = append(debates, debate)

	}

	// return the conditioned data
	return debates, nil
}

// sanitizeColumnName remove [#] from column title.
func sanitizeColumnName(val string) string {

	rex := regexp.MustCompile(`\[\d\]`)
	candName := strings.Trim(rex.ReplaceAllString(val, ""), " ")

	return candName

}

// writeCsv is a helper function that writes data to a CSV file
func writeCsv(fileName string, data [][]string) error {
	f, err := os.Create(fileName)

	if err != nil {
		return fmt.Errorf("could not open csv: %v", err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	csvWriter := csv.NewWriter(f)

	err = csvWriter.WriteAll(data)

	if err != nil {
		return fmt.Errorf("could not write to csv file '%v': %v", err)
	}

	return nil

}

// readCsv is a helper function which reads data from a CSV file
func readCsv(fileName string) ([][]string, error) {
	f, err := os.Open(fileName)

	if err != nil {
		return nil, fmt.Errorf("could not open csv: %v", err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()

	if err != nil {
		return nil, fmt.Errorf("could not read csv: %v", err)
	}

	return records, nil

}
