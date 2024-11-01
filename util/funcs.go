package util

import (
	"database/sql"
	"encoding/json"
	"example/axa_util/types"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func GetPolicyDuration(Policies []types.Policy, tn, tn2, product string) (string, string) {
	for _, val := range Policies {
		fmt.Println(tn, tn2)
		if cleanErrorInPolicyNumber(val.TrackingNumber) == cleanErrorInPolicyNumber(tn) || cleanErrorInPolicyNumber(val.TrackingNumber) == cleanErrorInPolicyNumber(tn2) {
			entry, err := time.Parse("2006-01-02", strings.Split(val.PolicyEndDate, "T")[0])
			if err != nil {
				return "", ""
			}

			entry = entry.Add(-time.Hour * 24 * time.Duration(types.Duration[product]))
			return entry.Format("2006-01-02"), strings.Split(val.PolicyEndDate, "T")[0]
		}
	}
	return "", ""
}

func UpdatePayload(db *sql.DB, id, payload string) error {
	_, err := db.Exec("UPDATE revenues SET payload = ? WHERE rev_id = ?", payload, id)
	return err
}

func GetAxaPolicyByProduct(product string, policies []types.Policy) *types.Policy {
	pol := types.PolicyType[product]
	for _, val := range policies {
		if CompareStrings(pol, val.BundleName) > 90 {
			return &val
		}
	}
	return nil
}

func GetPolicyByPhoneAndProduct(db *sql.DB, phone, product string) string {
	tn := ""
	row := db.QueryRow(`SELECT tracknum FROM trackingNumbers WHERE msisdn = ? AND productid=?`, phone, product)
	err := row.Scan(&tn)
	if err != nil {
		fmt.Println(err)
	}
	return tn
}

func cleanErrorInPolicyNumber(policyno string) string {
	if len(policyno) > 5 {
		if policyno[0:4] == "AXAM" {
			policyno = "90" + policyno[8:]
		}
	}
	return policyno
}

func Success(filename string, val []string, opt ...string) {
	var d []interface{}
	for _, v := range val {
		d = append(d, v)
	}

	var data [][]interface{}
	data = append(data, d)
	err := CreateExcelFile(filename, opt, data)
	if err != nil {
		fmt.Printf("failed to save %v", err)
		return
	}
	fmt.Println("Successfully added", err)
}

func Fail(val []string) {
	var d []interface{}
	for _, v := range val {
		d = append(d, v)
	}

	var data [][]interface{}
	data = append(data, d)
	err := CreateExcelFile("failed.xlsx", []string{"order_id", "gmt_created", "effect_date", "expire_date"}, data)
	if err != nil {
		fmt.Printf("failed to save %v", err)
		return
	}
	fmt.Println("Successfully added", err)
}

func UpdateTrackNumber2(db *sql.DB, tn, tn2 string) error {
	_, err := db.Exec("UPDATE trackingNumbers SET tracknum2 = ? WHERE tracknum = ?", tn2, tn)
	return err
}
func GetRevenueByTrxnID(db *sql.DB, trxnid string) (rev types.Revenue, err error) {
	// Define the query with a WHERE clause for trxnid and updatetype = 'modification'
	query := `
	  SELECT msisdn, payload, rev_id, productid, IFNULL(tracknumber, ''), dateadded
	  FROM revenues
	  WHERE trxnid = ? LIMIT 1`

	err = db.QueryRow(query, trxnid).Scan(&rev.Msisdn, &rev.Paylaod, &rev.Revid, &rev.ProductId, &rev.TrackNumber, &rev.Dateadded)
	if err != nil {
		if err == sql.ErrNoRows {
			return rev, fmt.Errorf("no record found with trxnid: %s", trxnid)
		}
		return rev, fmt.Errorf("failed to execute query: %v", err)
	}

	if rev.TrackNumber == "" {
		row := db.QueryRow("SELECT tracknum, tracknum2 FROM trackingNumbers WHERE msisdn = ?", rev.Msisdn)
		if row.Err() != nil {
			fmt.Println(err)
			return rev, nil
		}

		err := row.Scan(&rev.TrackNumber, &rev.TrackNumber2)
		if err == nil {
			fmt.Println(err)
			return rev, nil
		}
	}

	rev.Period = types.Duration[rev.ProductId]
	return rev, nil
}

func CreateExcelFile(filename string, headers []string, data [][]interface{}) error {
	// Try to open an existing file, or create a new one if it doesn't exist
	f, err := excelize.OpenFile(filename)
	if err != nil {
		// If the file doesn't exist, create a new file
		f = excelize.NewFile()
	}

	// Create or open a sheet named "Sheet1"
	sheetName := "Sheet1"
	index, err := f.GetSheetIndex(sheetName)
	if err != nil {
		return fmt.Errorf("failed to find sheet: %v", err)

	}
	if index == 0 {
		index, err = f.NewSheet(sheetName)
		if err != nil {
			return fmt.Errorf("failed to create or find sheet: %v", err)
		}
	}

	// Add headers dynamically, if they do not exist
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1) // i+1 is for column index, 1 is for row 1 (headers)
		if err != nil {
			return fmt.Errorf("failed to determine cell name for header: %v", err)
		}
		existingValue, err := f.GetCellValue(sheetName, cell)
		if err != nil {
			return fmt.Errorf("failed to get cell value: %v", err)
		}
		if existingValue == "" {
			f.SetCellValue(sheetName, cell, header) // Set the header only if the cell is empty
		}
	}

	// Find the last row in the sheet to append data after
	lastRow := 1
	rows, err := f.GetRows(sheetName)
	if err == nil {
		lastRow = len(rows) // Set the starting row after the last existing row
	}

	// Add new data rows starting from the row after the last one found
	for rowIndex, row := range data {
		for colIndex, cellValue := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, lastRow+rowIndex+1) // Adjust for row index
			if err != nil {
				return fmt.Errorf("failed to determine cell name for data: %v", err)
			}
			f.SetCellValue(sheetName, cell, cellValue) // Set each cell's value dynamically
		}
	}

	// Set the active sheet
	f.SetActiveSheet(index)

	// Save the Excel file
	err = f.SaveAs(filename)
	if err != nil {
		return fmt.Errorf("failed to save Excel file: %v", err)
	}

	return nil
}

func CallAPI(msisdn string) (*types.Response, error) {
	url := "https://www.axamansard.com/ecpartnerapi/api/v1/partner/get-enrollmentNo-by-phoneno/" + msisdn // Replace with actual URL

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Add the Authorization header with the token
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjMwODkiLCJuYmYiOjE3MzAzMTYwODEsImV4cCI6MTczMDQwMjQ4MSwiaWF0IjoxNzMwMzE2MDgxfQ.OAjh0V5ENPpZEFMRGTd1OHa8tgiLf3HTaIISR1m1PhY")

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Println(string(body))

	// Parse the JSON response into the Response struct
	var result types.Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	return &result, nil
}

func ReadExcel(filePath string, sheetName string) [][]string {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatalf("failed to open Excel file: %v", err)
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Fatalf("failed to get rows from sheet %s: %v", sheetName, err)
	}
	return rows
}
