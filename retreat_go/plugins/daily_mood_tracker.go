package plugins

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/MatthewTe/retreat/database"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func SetDailyMoodGoogleSheets(moodValue float64, spreadsheetId string) {

	ctx := context.Background()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to find home directory: %v", err)
	}
	credPath := filepath.Join(homeDir, ".google_api_credentials.json")

	b, err := os.ReadFile(credPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := GetGoogleClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Currently don't care if it throws in exception because tbl will already exist after first run - handle error later:
	CreateGoogleSheetTable(spreadsheetId)

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	newRow := sheets.RowData{
		Values: []*sheets.CellData{
			{UserEnteredValue: &sheets.ExtendedValue{StringValue: &currentTime}},
			{UserEnteredValue: &sheets.ExtendedValue{NumberValue: &moodValue}},
		}}

	appendCellsRequest := sheets.AppendCellsRequest{
		Fields:  "*",
		TableId: "daily_tracker",
		Rows:    []*sheets.RowData{&newRow},
	}

	batchUpdateRequest := sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{AppendCells: &appendCellsRequest},
		},
	}

	tblUpdateResp, err := srv.Spreadsheets.BatchUpdate(spreadsheetId, &batchUpdateRequest).Do()
	if err != nil {
		log.Fatalf("Unable to update data from sheet: %v", err)
	}

	fmt.Println(tblUpdateResp.ServerResponse)

}

// func that gets all moods and aggregate by day
func AggregateTimeseriesValues(dbValues []database.TimeseriesValues) (map[time.Time]float32, error) {

	var valuesAggPerDay map[string][]float32 = map[string][]float32{}

	for _, value := range dbValues {
		dateObject, err := time.Parse("2025-08-02 17:04:06", value.Date)
		if err != nil {
			log.Fatalln(err)
		}

		if seriesArray, ok := valuesAggPerDay[dateObject.Format("2025-08-02")]; ok {
			seriesArray = append(seriesArray, value.Value)
			valuesAggPerDay[dateObject.Format("2025-08-02")] = seriesArray
		} else {
			valuesAggPerDay[dateObject.Format("2025-08-02")] = []float32{value.Value}
		}

	}

	// Averaging values:
	var calAvgValues map[time.Time]float32 = map[time.Time]float32{}

	for dateStr, associatedVals := range valuesAggPerDay {

		var sum float32 = 0.0
		for _, num := range associatedVals {
			sum += num
		}

		reParsedTime, err := time.Parse("2025-08-02", dateStr)
		calAvgValues[reParsedTime] = sum / float32(len(associatedVals))
		if err != nil {
			log.Fatalln(err)
		}
	}

	return calAvgValues, nil

}

// func that takes in pointer to model object and sets the state / does other stuff.

// Might need a view() function
