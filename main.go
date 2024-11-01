package main

import (
	"example/axa_util/types"
	"example/axa_util/util"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	//Get the transaction msisdn, productid, dateadded, payload,
	//confirm the recent expiry from axa, if not valid and not greater the most recent payment
	//check the history of payment to validate what payment is.
	//fix the payment according the last expiry date

	db, err := util.ConnectDB()
	if err != nil {
		fmt.Println(db)
		return
	}

	row, err := db.Query("SELECT trxnid, dateadded, payload, expirydate, msisdn, productid FROM revenues WHERE network = 'PALMPAY' AND expirydate < dateadded AND productid = 'AXA60970348' ORDER BY dateadded DESC")
	if err != nil {
		fmt.Println(err)
		return
	}

	count := 0

	mp := map[string]types.Revenue{}

	for _, val := range util.ReadExcel("correct_date.xlsx", "Sheet1") {
		mp[val[0]] = types.Revenue{
			Trxnid:      val[0],
			Expirydate:  val[2],
			TrackNumber: val[3],
			Effective:   val[1],
		}
	}
	fmt.Println(len(mp))

	// return
	for row.Next() {
		revenue := types.Revenue{}
		err := row.Scan(&revenue.Trxnid, &revenue.Dateadded, &revenue.Paylaod, &revenue.Expirydate, &revenue.Msisdn, &revenue.ProductId)
		if err != nil {
			fmt.Println(err)
			return
		}

		m, ok := mp[revenue.Trxnid]
		if ok {
			util.Success("correct_date3.xlsx", []string{revenue.Trxnid, m.Effective, m.Expirydate, m.TrackNumber,revenue.ProductId, revenue.Msisdn},
				"transaction_id", "effective_date", "expiry_date", "track_number","product", "msisdn")

			continue
		}
		count++
	}
	fmt.Println(count)
}
