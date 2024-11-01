package main

import (
	"encoding/json"
	"example/axa_util/types"
	"example/axa_util/util"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := util.ConnectDB()
	if err != nil {
		fmt.Println(db)
		return
	}
	defer db.Close()

	str := []string{"2024-10-24", "2024-10-28", "2024-10-29", "2024-10-30"}

	// change transaction to csv
	// check initial transaction
	// compare the expiry of initial transaction and effect date of the new transaction
	//if it is invalid set the intitial

	// mp := map[string]bool{}

	// for _, val := range util.ReadExcel("new_correct_date.xlsx", "Sheet1") {
	// 	mp[val[0]] = false

	// }
	// fmt.Println(len(mp))

	processed := map[string]string{}

	count := 0
	for _, val := range str {
		t, _ := time.Parse("2006-01-02", val)
		row, err := db.Query(`SELECT trxnid, dateadded, payload, expirydate, msisdn, productid FROM revenues 
			WHERE network = 'PALMPAY' AND productid = 'AXA60970348' AND payload <> "" AND dateadded BETWEEN ? AND ? ORDER BY dateadded`, val, t.Add(time.Hour*24).Format("2006-01-02"))
		if err != nil {
			fmt.Println(err)
			return
		}

		for row.Next() {
			revenue := types.Revenue{}
			err := row.Scan(&revenue.Trxnid, &revenue.Dateadded, &revenue.Paylaod, &revenue.Expirydate, &revenue.Msisdn, &revenue.ProductId)
			if err != nil {
				fmt.Println(err)
				return
			}

			// _, ok := mp[revenue.Trxnid]
			// if ok {
			// 	continue
			// }

			if count < 3 {
				fmt.Println(revenue)
			}
			revTrxn := types.Transaction{}
			json.Unmarshal([]byte(revenue.Paylaod), &revTrxn)
			initEff := revTrxn.EffectiveDate

			revenue.Dateadded = util.FormatToDate(revenue.Dateadded)
			if revTrxn.EffectiveDate >= revenue.Dateadded {
				continue
			}

			if revTrxn.EffectiveDate > revenue.Dateadded {
				count++

				fmt.Println("===executed======", revTrxn.EffectiveDate, revenue.Dateadded, revenue.Trxnid)
			}

			/*
				if effective is less than the date of purchase,
				check if axa has that is earlier
				if axa has an ealier date use the date from axa.
				else check the db for previous transaction and see if we have a valid transaction.
				keep track of transaction for a particular number.
			*/

			var axares types.Response

			if _, ok := processed[revenue.ProductId+revenue.Msisdn]; !ok {
				res, err := util.CallAPI(revenue.Msisdn)
				if err != nil {
					return
				}
				axares = *res
			}

			policy := util.GetAxaPolicyByProduct(revenue.ProductId, axares.ReturnedObject)
			if policy != nil {
				revTrxn.EffectiveDate = util.GetStart(policy.PolicyEndDate, -types.Duration[revenue.ProductId])
				if revTrxn.EffectiveDate < revenue.Dateadded {
					prev := types.Revenue{}
					r := db.QueryRow("SELECT trxnid, dateadded, payload, expirydate, msisdn, productid, updatetype FROM revenues WHERE productid = ? AND msisdn = ? AND dateadded < ? ORDER BY dateadded DESC LIMIT 1",
						revenue.ProductId, revenue.Msisdn, revenue.Dateadded)
					if r.Err() != nil {
						fmt.Println(r.Err())
					}

					err = r.Scan(&prev.Trxnid, &prev.Dateadded, &prev.Paylaod, &prev.Expirydate, &prev.Msisdn, &prev.ProductId, &prev.Paytype)
					if err != nil {
						fmt.Println(err)
					}

					if prev.Trxnid == "" {
						continue
					}

					prevExpime, _ := time.Parse("2006-01-02", prev.Expirydate)
					revDateaddTme, _ := time.Parse("2006-01-02", revenue.Dateadded)
					isValid := prevExpime.Sub(revDateaddTme) <= time.Hour*24*time.Duration(types.Duration[prev.ProductId])

					if prev.Expirydate > prev.Dateadded && isValid && prevExpime.Sub(revDateaddTme) > 0 {
						// valid = true
						dadded, err := time.Parse("2006-01-02", prev.Expirydate)
						if err != nil {
							fmt.Println("error formating date added", dadded)
							continue
						}
						revTrxn.EffectiveDate = prev.Expirydate
						revTrxn.ExpiryDate = dadded.Add(time.Hour * 24 * time.Duration(types.Duration[revenue.ProductId])).Format("2006-01-02")
						fmt.Println("USING EXTENSION")
					}

					processed[revenue.ProductId+revenue.Msisdn] = revTrxn.ExpiryDate

				} else {
					revTrxn.ExpiryDate = util.GetStart(policy.PolicyEndDate, 0)
					revTrxn.EffectiveDate = util.GetStart(revTrxn.ExpiryDate, -types.Duration[revenue.ProductId])

					processed[revenue.ProductId+revenue.Msisdn] = revTrxn.ExpiryDate

				}
			} else {
				m, ok := processed[revenue.ProductId+revenue.Msisdn]
				if !ok {

					prev := types.Revenue{}
					r := db.QueryRow("SELECT trxnid, dateadded, payload, expirydate, msisdn, productid, updatetype FROM revenues WHERE productid = ? AND msisdn = ? AND dateadded < ? ORDER BY dateadded DESC LIMIT 1",
						revenue.ProductId, revenue.Msisdn, revenue.Dateadded)
					if r.Err() != nil {
						fmt.Println(r.Err())
					}

					err = r.Scan(&prev.Trxnid, &prev.Dateadded, &prev.Paylaod, &prev.Expirydate, &prev.Msisdn, &prev.ProductId, &prev.Paytype)
					if err != nil {
						fmt.Println(err)
					}

					if prev.Trxnid == "" {
						continue
					}

					prevExpime, _ := time.Parse("2006-01-02", prev.Expirydate)
					revDateaddTme, _ := time.Parse("2006-01-02", revenue.Dateadded)
					isValid := prevExpime.Sub(revDateaddTme) <= time.Hour*24*time.Duration(types.Duration[prev.ProductId])

					if prev.Expirydate > prev.Dateadded && isValid && prevExpime.Sub(revDateaddTme) > 0 {
						// valid = true
						dadded, err := time.Parse("2006-01-02", prev.Expirydate)
						if err != nil {
							fmt.Println("error formating date added", dadded)
							continue
						}
						revTrxn.EffectiveDate = prev.Expirydate
						revTrxn.ExpiryDate = dadded.Add(time.Hour * 24 * time.Duration(types.Duration[revenue.ProductId])).Format("2006-01-02")
						fmt.Println("USING EXTENSION")
					}

					processed[revenue.ProductId+revenue.Msisdn] = revTrxn.ExpiryDate

				} else {
					revTrxn.EffectiveDate = m
					revTrxn.ExpiryDate = util.GetStart(m, types.Duration[revenue.ProductId])
				}
			}

			if revTrxn.EffectiveDate < revenue.Dateadded {
				dadded, err := time.Parse("2006-01-02", revenue.Dateadded)
				if err != nil {
					fmt.Println("error formating date added", dadded)
					continue
				}
				revTrxn.EffectiveDate = dadded.Format("2006-01-02")
				revTrxn.ExpiryDate = dadded.Add(time.Hour * 24 * time.Duration(types.Duration[revenue.ProductId])).Format("2006-01-02")
			}

			util.Success("new_correct_date.xlsx", []string{revenue.Trxnid, revTrxn.EffectiveDate, revTrxn.ExpiryDate, revTrxn.TrackNumber, revenue.ProductId, revenue.Msisdn, revenue.Expirydate, initEff},
				"transaction_id", "effective_date", "expiry_date", "track_number", "productid", "msisdn", "old_expiry", "old_effective")

			fmt.Printf("transactionID: %v \n effectiveDate: %v \n expiryDate: %v \n trackNumber: %v \n ************* \n",
				revTrxn.TransactionID, revTrxn.EffectiveDate, revTrxn.ExpiryDate, revTrxn.TrackNumber)

			count++

		}

	}
	fmt.Println(count)
}
