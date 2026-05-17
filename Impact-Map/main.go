package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	ctx := context.Background()

	b, err := os.ReadFile("../credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(b))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	spreadsheetID := "1ASYXQ3Bdt0FWHPqG7nQ1ZVKDCzk0Lk2CsHiTHHP4jaA"

	fmt.Println("🛰️  Fetching tabs...")

	bridgeResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, "'System Bridge (Anchor point)'!A:Z").
		ValueRenderOption("FORMATTED_VALUE").Do()
	if err != nil {
		log.Fatalf("Failed to read System Bridge: %v", err)
	}

	imeResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, "'FinalIME'!A:Z").
		ValueRenderOption("FORMATTED_VALUE").Do()
	if err != nil {
		log.Fatalf("Failed to read FinalIME: %v", err)
	}

	bridgeRows := bridgeResp.Values
	imeRows := imeResp.Values

	if len(bridgeRows) == 0 || len(imeRows) == 0 {
		log.Fatalf("❌ One or both sheets are completely empty.")
	}

	// 🎯 1. Map Header Positions for Mutual Column Names
	bridgeHeaders := bridgeRows[0]
	imeHeaders := imeRows[0]

	type ColumnMatch struct {
		Name      string
		BridgeIdx int
		ImeIdx    int
	}

	var matchedColumns []ColumnMatch

	fmt.Println("🔍 Aligning column names...")
	for bIdx, bCell := range bridgeHeaders {
		bName := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", bCell)))
		if bName == "" {
			continue
		}

		for iIdx, iCell := range imeHeaders {
			iName := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", iCell)))
			if bName == iName {
				matchedColumns = append(matchedColumns, ColumnMatch{
					Name:      strings.TrimSpace(fmt.Sprintf("%v", bCell)),
					BridgeIdx: bIdx,
					ImeIdx:    iIdx,
				})
				break
			}
		}
	}

	fmt.Println("📋 Columns evaluated for comparison:")
	for _, col := range matchedColumns {
		fmt.Printf("   ✅ %s (Bridge Col %d ↔️ FinalIME Col %d)\n", col.Name, col.BridgeIdx+1, col.ImeIdx+1)
	}
	fmt.Println("--------------------------------------------------")

	// 🎯 2. Compare Data Rows based on Matched Column Map
	maxRows := len(bridgeRows)
	if len(imeRows) > maxRows {
		maxRows = len(imeRows)
	}

	mismatches := 0

	for i := 1; i < maxRows; i++ { // Start at 1 to skip header row
		var bRow []interface{}
		if i < len(bridgeRows) {
			bRow = bridgeRows[i]
		}

		var iRow []interface{}
		if i < len(imeRows) {
			iRow = imeRows[i]
		}

		// Helper to check if a row is completely blank or missing
		isBridgeEmpty := len(bRow) == 0
		isImeEmpty := len(iRow) == 0

		if isBridgeEmpty && isImeEmpty {
			continue // Skip matching empty rows
		}

		// Grab an ID label for clear error reporting
		rowID := fmt.Sprintf("Row %d", i+1)
		if !isBridgeEmpty && len(bRow) > 0 {
			rowID = fmt.Sprintf("Row %d (ID: %v)", i+1, bRow[0])
		} else if !isImeEmpty && len(iRow) > 0 {
			rowID = fmt.Sprintf("Row %d (ID: %v)", i+1, iRow[0])
		}

		if isBridgeEmpty {
			fmt.Printf("❌ %s exists in FinalIME but is completely blank in System Bridge.\n", rowID)
			mismatches++
			continue
		}
		if isImeEmpty {
			fmt.Printf("❌ %s exists in System Bridge but is completely blank in FinalIME.\n", rowID)
			mismatches++
			continue
		}

		// Evaluate column intersections row-by-row
		rowHasMismatch := false
		var diffLogs []string

		for _, col := range matchedColumns {
			var bCell, iCell string

			if col.BridgeIdx < len(bRow) {
				bCell = strings.TrimSpace(fmt.Sprintf("%v", bRow[col.BridgeIdx]))
			}
			if col.ImeIdx < len(iRow) {
				iCell = strings.TrimSpace(fmt.Sprintf("%v", iRow[col.ImeIdx]))
			}

			if bCell != iCell {
				rowHasMismatch = true
				diffLogs = append(diffLogs, fmt.Sprintf("   👉 [%s] -> System Bridge: %q | FinalIME: %q", col.Name, bCell, iCell))
			}
		}

		if rowHasMismatch {
			mismatches++
			fmt.Printf("❌ Mismatch at %s:\n", rowID)
			for _, logStr := range diffLogs {
				fmt.Println(logStr)
			}

			if mismatches >= 10 {
				fmt.Println("\n⚠️  Showing first 10 data mismatches. Halting to prevent terminal log spam.")
				break
			}
		}
	}

	if mismatches == 0 {
		fmt.Println("✅ Success! Shared data values across matching column names are identical.")
	} else {
		fmt.Printf("\n❌ Total row data mismatches identified: %d\n", mismatches)
	}
}
