package reader

import (
	"encoding/csv"
	"os"
)

func CSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()

	return records, err
}
