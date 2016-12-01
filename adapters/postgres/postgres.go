package postgres

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/caarlos0/env"
	"github.com/jmoiron/sqlx"
	// Used pg drive on sqlx
	_ "github.com/lib/pq"

	"github.com/nuveo/prest/config"
)

const (
	pageNumberKey   = "_page"
	pageSizeKey     = "_page_size"
	defaultPageSize = "10"
)

// Conn connect on PostgreSQL
// Used sqlx
func Conn() (db *sqlx.DB) {
	cfg := config.Prest{}
	env.Parse(&cfg)

	db, err := sqlx.Connect("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable", cfg.PGUser, cfg.PGDatabase))
	if err != nil {
		panic(fmt.Sprintf("Unable to connection to database: %v\n", err))
	}
	return
}

// WhereByRequest create interface for queries + where
func WhereByRequest(r *http.Request) (whereSyntax string) {
	u, _ := url.Parse(r.URL.String())
	where := []string{}
	for key, val := range u.Query() {
		if !strings.HasPrefix(key, "_") {
			where = append(where, fmt.Sprintf("%s='%s'", key, val[0]))
		}
	}

	filter := strings.Join(where, " and ")
	paginatedQuery := paginateIfPossible(u.Query())
	whereSyntax = fmt.Sprint(filter, paginatedQuery)
	return
}

// Query process queries
func Query(SQL string, params ...interface{}) (jsonData []byte, err error) {
	db := Conn()
	rows, err := db.Queryx(SQL, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}
	jsonData, err = json.Marshal(tableData)

	return
}

func paginateIfPossible(values url.Values) (paginatedQuery string) {
	if _, ok := values[pageNumberKey]; !ok {
		paginatedQuery = ""
		return
	}
	pageNumber := values[pageNumberKey][0]
	pageSize := defaultPageSize
	if size, ok := values[pageSizeKey]; ok {
		pageSize = size[0]
	}
	paginatedQuery = fmt.Sprintf("LIMIT %s OFFSET(%s - 1) * %s", pageSize, pageNumber, pageSize)
	return
}
