package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type CustomerData struct {
	Name string
	Type string
}

type LocationData struct {
	District  string
	Latitude  string
	Longitude string
}

func main() {
	ctx := context.Background()

	// 1. Read the JSON credentials file
	b, err := os.ReadFile("../credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	// 2. Initialize Service
	srv, err := sheets.NewService(ctx,
		option.WithCredentialsJSON(b),
		option.WithScopes(sheets.SpreadsheetsScope, drive.DriveReadonlyScope),
	)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// --- 3. EXPLICIT TRUE SOURCE IDs ---
	masterSystemID := "1V6UzkyQ6CHRN1RbXi039GJzp00QxmW49zIR6T-xlu1g"
	destID := "1ASYXQ3Bdt0FWHPqG7nQ1ZVKDCzk0Lk2CsHiTHHP4jaA"

	customerMasterID := "1GJy4RzaC8ws8QQ5HMG1kONMO37kW6uFvJGcuBZit9iM"
	locationMasterID := "15R0AfrWDvkMN1VTfap-NNrH-gca7adRvphF-8O0EOVs"
	populationMasterID := "1qhinR30z8MdoyBlRgi-MnFegivSDMdwQ8oH2GbqcYFg"

	fmt.Println("🛰️ Fetching raw data straight from true source sheets...")

	// --- 4. FETCH RAW SYSTEM DATA ---
	systemResp, err := srv.Spreadsheets.Values.Get(masterSystemID, "'System'!A:K").Do()
	if err != nil {
		log.Fatalf("Failed to read raw Master System sheet: %v", err)
	}

	// --- 5. FETCH AND BUILD IN-MEMORY LOOKUP TABLES ---

	// Fetch Customer Tab (Columns A to D)
	customerResp, err := srv.Spreadsheets.Values.Get(customerMasterID, "'Customer'!A:D").Do()
	if err != nil {
		log.Fatalf("Failed to read true Customer source master: %v.", err)
	}
	customerMap := make(map[string]CustomerData)
	for i, row := range customerResp.Values {
		if i == 0 || len(row) < 4 {
			continue
		}
		custID := strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		customerMap[custID] = CustomerData{
			Name: strings.TrimSpace(fmt.Sprintf("%v", row[1])),
			Type: strings.TrimSpace(fmt.Sprintf("%v", row[3])),
		}
	}

	// Fetch Location Tab with Formatted Value Rendering Option
	locationResp, err := srv.Spreadsheets.Values.Get(locationMasterID, "'Location'!A:Z").
		ValueRenderOption("FORMATTED_VALUE").
		Do()
	if err != nil {
		log.Fatalf("Failed to read true Location source master: %v.", err)
	}
	locationMap := make(map[string]LocationData)

	fmt.Println("🔍 Dynamically mapping Location columns by header name...")

	// Default fallbacks if headers aren't detected explicitly
	idIdx := 0
	latIdx := -1
	longIdx := -1
	districtIdx := -1
	coordStrIdx := -1

	// Scan the header row to lock onto coordinates dynamically
	if len(locationResp.Values) > 0 {
		headerRow := locationResp.Values[0]
		for idx, cellVal := range headerRow {
			cellStr := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", cellVal)))
			switch {
			case cellStr == "customer id" || cellStr == "id":
				idIdx = idx
			case strings.Contains(cellStr, "latitude"):
				latIdx = idx
			case strings.Contains(cellStr, "longitude"):
				longIdx = idx
			case cellStr == "district":
				districtIdx = idx
			case strings.Contains(cellStr, "coordinate") || strings.Contains(cellStr, "gps"):
				coordStrIdx = idx
			}
		}
	}

	parsedCoordinatesCount := 0

	for i, row := range locationResp.Values {
		if i == 0 || len(row) <= idIdx {
			continue
		}

		custID := strings.TrimSpace(fmt.Sprintf("%v", row[idIdx]))
		if custID == "" || custID == "Customer ID" {
			continue
		}

		var lat, long, district string

		// 1. Pull District if found dynamically
		if districtIdx != -1 && len(row) > districtIdx {
			district = strings.TrimSpace(fmt.Sprintf("%v", row[districtIdx]))
		}

		// 2. Extract Lat/Long values using split individual or combined columns dynamically
		if latIdx != -1 && longIdx != -1 && len(row) > latIdx && len(row) > longIdx {
			lat = strings.TrimSpace(fmt.Sprintf("%v", row[latIdx]))
			long = strings.TrimSpace(fmt.Sprintf("%v", row[longIdx]))
		}

		// If explicit Lat/Long columns were empty or missing, fallback to parsing a combined coordinates string column
		if (lat == "" || long == "") && coordStrIdx != -1 && len(row) > coordStrIdx {
			coordStr := strings.TrimSpace(fmt.Sprintf("%v", row[coordStrIdx]))
			if coordStr != "" && strings.Contains(coordStr, ",") {
				if parts := strings.Split(coordStr, ","); len(parts) >= 2 {
					lat = strings.TrimSpace(parts[0])
					long = strings.TrimSpace(parts[1])
				}
			}
		}

		if lat != "" && long != "" {
			// Double-check we didn't pull descriptive names into numerical spaces
			_, errLat := strconv.ParseFloat(lat, 64)
			_, errLong := strconv.ParseFloat(long, 64)
			if errLat == nil && errLong == nil {
				parsedCoordinatesCount++
				locationMap[custID] = LocationData{
					District:  district,
					Latitude:  lat,
					Longitude: long,
				}
			}
		}
	}
	fmt.Printf("ℹ️  Successfully processed map coordinates for %d rows out of raw source.\n", parsedCoordinatesCount)

	// Fetch Population Tab (Columns A to M)
	popResp, err := srv.Spreadsheets.Values.Get(populationMasterID, "'Population'!A:M").Do()
	if err != nil {
		log.Fatalf("Failed to read true Population source master: %v.", err)
	}

	latestPopYear := make(map[string]int)
	latestPopVal := make(map[string]string)

	for i, row := range popResp.Values {
		if i == 0 || len(row) < 13 {
			continue
		}
		custID := strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		yearStr := strings.TrimSpace(fmt.Sprintf("%v", row[4]))
		popVal := strings.TrimSpace(fmt.Sprintf("%v", row[6]))

		yearInt, _ := strconv.Atoi(yearStr)

		if currentMaxYear, exists := latestPopYear[custID]; !exists || yearInt > currentMaxYear {
			latestPopYear[custID] = yearInt
			latestPopVal[custID] = popVal
		}
	}

	// --- 6. PROCESS, FILTER, SORT, AND JOIN DATA ---
	var tempRows [][]string
	customerSystemCounts := make(map[string]int)

	for i, row := range systemResp.Values {
		if i == 0 || len(row) < 9 {
			continue
		}

		status := strings.TrimSpace(fmt.Sprintf("%v", row[7]))
		if strings.EqualFold(status, "Installed") {
			custID := strings.TrimSpace(fmt.Sprintf("%v", row[3]))

			customerSystemCounts[custID]++
			sysNum := strconv.Itoa(customerSystemCounts[custID])

			tempRows = append(tempRows, []string{
				custID,
				strings.TrimSpace(fmt.Sprintf("%v", row[0])),
				strings.TrimSpace(fmt.Sprintf("%v", row[5])),
				strings.TrimSpace(fmt.Sprintf("%v", row[2])),
				status,
				strings.TrimSpace(fmt.Sprintf("%v", row[8])),
				sysNum,
			})
		}
	}

	// Sort globally by Customer ID (Numeric)
	sort.Slice(tempRows, func(i, j int) bool {
		valI, _ := strconv.Atoi(tempRows[i][0])
		valJ, _ := strconv.Atoi(tempRows[j][0])
		return valI < valJ
	})

	// Assemble final combined matrix
	var finalData [][]interface{}
	finalData = append(finalData, []interface{}{
		"Customer ID", "System ID", "System Name", "Design Type", "Status", "Install Date",
		"Customer System Number", "Customer System Total",
		"Matched Customer Name", "Matched Customer Type",
		"Matched District", "Matched Latitude", "Matched Longitude", "Matched Population",
	})

	for _, r := range tempRows {
		custID := r[0]

		custInfo := customerMap[custID]
		locInfo := locationMap[custID]
		popVal := latestPopVal[custID]
		totalSys := strconv.Itoa(customerSystemCounts[custID])

		finalData = append(finalData, []interface{}{
			custID,
			r[1],
			r[2],
			r[3],
			r[4],
			r[5],
			r[6],
			totalSys,
			custInfo.Name,
			custInfo.Type,
			locInfo.District,
			locInfo.Latitude,
			locInfo.Longitude,
			popVal,
		})
	}

	// --- 🎯 DATA INTEGRITY LOGGING ---
	fmt.Println("🕵️ Pre-validating dataset integrity inside internal matrix...")
	preWriteEmptyLats := 0
	preWriteEmptyLongs := 0
	for idx, row := range finalData {
		if idx == 0 {
			continue
		}
		if fmt.Sprintf("%v", row[11]) == "" {
			preWriteEmptyLats++
		}
		if fmt.Sprintf("%v", row[12]) == "" {
			preWriteEmptyLongs++
		}
	}

	if preWriteEmptyLats > 0 || preWriteEmptyLongs > 0 {
		fmt.Printf("⚠️  Data Notice: Out of %d total joined records, %d rows contain unmapped/empty coordinates from source entries.\n", len(finalData)-1, preWriteEmptyLats)
	}

	// --- 7. OVERWRITE DESTINATION 'FinalIME' ---
	_, err = srv.Spreadsheets.Values.Clear(destID, "FinalIME!A:Z", &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		log.Fatalf("Clear error: %v", err)
	}

	rb := &sheets.ValueRange{Values: finalData}
	_, err = srv.Spreadsheets.Values.Update(destID, "FinalIME!A1", rb).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Write error: %v", err)
	}

	// --- 8. POST-WRITE DATA VALIDATION CHECK ---
	fmt.Println("🕵️ Running output validation check on live destination sheet...")
	verifyResp, err := srv.Spreadsheets.Values.Get(destID, "FinalIME!A:N").Do()
	if err != nil {
		log.Fatalf("Post-validation fetch error: %v", err)
	}

	filledCoordinatesCount := 0
	for rowIdx, rowValues := range verifyResp.Values {
		if rowIdx == 0 {
			continue
		}

		var latCell, longCell string
		if len(rowValues) > 11 {
			latCell = strings.TrimSpace(fmt.Sprintf("%v", rowValues[11]))
		}
		if len(rowValues) > 12 {
			longCell = strings.TrimSpace(fmt.Sprintf("%v", rowValues[12]))
		}

		if latCell != "" && longCell != "" {
			filledCoordinatesCount++
		}
	}

	fmt.Printf("✅ Success! Safely verified %d records with fully populated geographic coordinate pairs in FinalIME.\n", filledCoordinatesCount)
	fmt.Printf("🎉 Done! Compiled %d total records straight into FinalIME.\n", len(finalData)-1)
}
