package util

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

// RemoveWhitespace removes all spaces and tabs from a string
func RemoveWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1 // Skip whitespace characters
		}
		return r
	}, s)
}

func TotringToStringMap(s string) map[string]string {
	m := map[string]string{}
	json.Unmarshal([]byte(s), &m)
	return m
}

func JsonToString(s interface{}) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "chinonsoo:Chin2023&CC@tcp(89.107.62.131:3306)/cubecover")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return db, nil
}

func GetStart(date string, duration int64) string {
	date = FormatToDate(date)
	t, _ :=	time.Parse("2006-01-02", date)
	fmt.Println(time.Duration(duration))
	t = t.Add(time.Hour * 24 * time.Duration(duration))
	return t.Format("2006-01-02")
}

func FormatToDate(str string) string {
	if len(strings.Split(str, " ")) > 1 {
		str = strings.Split(str, " ")[0]
	} else {
		str = strings.Split(str, "T")[0]

	}
	// Define possible date formats
	dateFormats := []string{"2006-Jan-02", "02-01-2006", "02-Jan-2006", "2006-01-02"}

	// Iterate over all defined formats and try parsing the date
	for _, format := range dateFormats {
		parsedTime, err := time.Parse(format, str)
		if err == nil {
			// If parsing is successful, return the date in YYYY-MM-DD format
			return parsedTime.Format("2006-01-02")
		}
	}
	return str
}

func ConnectRedis() (*redis.Client, error) {
	// localhost:6379
	redClient := redis.NewClient(&redis.Options{
		// localhost:
		Addr:     "localhost:6379",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	return redClient, nil
}

// CompareStrings compares two strings case-insensitively and ignores whitespace, returning the percentage of match
func CompareStrings(s1, s2 string) float64 {
	// Convert both strings to lowercase to make comparison case-insensitive
	s1 = strings.ToLower(RemoveWhitespace(s1))
	s2 = strings.ToLower(RemoveWhitespace(s2))

	// Find the length of the shorter and longer strings
	len1 := len(s1)
	len2 := len(s2)
	maxLen := int(math.Max(float64(len1), float64(len2)))

	// If both strings are empty, they are considered 100% match
	if maxLen == 0 {
		return 100.0
	}

	// Count character matches
	matches := 0
	minLen := int(math.Min(float64(len1), float64(len2)))

	for i := 0; i < minLen; i++ {
		if s1[i] == s2[i] {
			matches++
		}
	}

	// Return percentage of matches relative to the longer string
	return (float64(matches) / float64(maxLen)) * 100.0
}
