package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [folder]")
		return
	}

	folder := os.Args[1]

	// フォルダ内のCSVのみを対象とする
	files, err := filepath.Glob(filepath.Join(folder, "*.txt"))
	if err != nil {
		fmt.Println("Error reading folder:", err)
		return
	}

	lo.ForEach(files, func(f string, index int) {
		fmt.Printf("処理対象 : %s\n", f)
		processFile(f)
	})
}

func processFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", filePath, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV file:", filePath, err)
		return
	}

	outputFilePath := strings.TrimSuffix(filePath, ".csv") + "_out.csv"
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("Error creating output file:", outputFilePath, err)
		return
	}
	defer outputFile.Close()

	// Excelで文字化けしないようにする設定。BOM付きUTF8をfileの先頭に付与。
	bw := bufio.NewWriter(outputFile)
	bw.Write([]byte{0xEF, 0xBB, 0xBF})

	// writer := csv.NewWriter(outputFile)
	writer := csv.NewWriter(bw)
	writer.UseCRLF = true
	defer writer.Flush()

	// ヘッダ
	outputRecord := []string{"住民状態", "自治体コード", "町名コード", "都道府県名", "市区郡町村名", "番地", " 都道府県名⇔番地", "市区郡町村名⇔番地"}
	if err := writer.Write(outputRecord); err != nil {
		fmt.Println("Error writing to output file:", outputFilePath, err)
		return
	}

	lo.ForEach(records, func(record []string, index int) {
		var juminJotai, jititaiCode, tyomeiCode, todofuken, sikugunchoson, banti string

		juminJotai = record[1]
		if len(record) == 10 { // 本籍以外
			jititaiCode = record[2]
			tyomeiCode = record[3]
			todofuken = record[4]
			sikugunchoson = record[5]
			banti = record[7]
		} else { //本籍
			jititaiCode = record[3]
			tyomeiCode = record[4]
			todofuken = record[5]
			sikugunchoson = record[6]
			banti = record[8]
		}

		// 都道府県名がセットされているレコードのみ対象とする
		if todofuken != "" {
			// 都道府県名⇔番地
			// → 愛媛県、茨城県など、県名が異なるものを拾う
			mismatchCount1 := countMismatches(todofuken, banti)
			// 市区郡町村名⇔番地
			// → 横浜市戸塚区
			mismatchCount2 := countMismatches(sikugunchoson, banti)
			outputRecord := []string{juminJotai, jititaiCode, tyomeiCode, todofuken, sikugunchoson, banti, fmt.Sprintf("%d", mismatchCount1), fmt.Sprintf("%d", mismatchCount2)}
			err := writer.Write(outputRecord)
			if err != nil {
				fmt.Println("Error writing to output file:", outputFilePath, err)
				return
			}
		}
	})
}

func countMismatches(strA, strB string) int {
	mismatches := 0
	sliceA := strings.Split(strA, "")
	sliceB := strings.Split(strB, "")
	for i := 0; i < len(sliceA) && i < len(sliceB); i++ {
		if sliceA[i] != sliceB[i] {
			mismatches++
		}
	}
	return mismatches
}
