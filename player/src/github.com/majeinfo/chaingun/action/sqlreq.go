package action

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"database/sql"
	"github.com/go-sql-driver/mysql"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"

	log "github.com/sirupsen/logrus"
)

const (
	REPORTER_SQL string = "SQL"
	SQL_ERR             = 500
)

// DoSQLRequest accepts a SQLAction and a one-way channel to write the results to.
func DoSQLRequest(sqlAction SQLAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
	var trace_req string

	// Applies variable to the statement
	stmt := SubstParams(sessionMap, sqlAction.Statement, vulog)

	if must_trace_request {
		trace_req = fmt.Sprintf("%s %s", sqlAction.Server, stmt)
	} else {
		vulog.Debugf("New Request: URL: %s, Request: %s", sqlAction.Server, stmt)
	}

	// Special case for MySQL
	server := sqlAction.Server
	if sqlAction.DBDriver == "mysql" {
		server += "/" + sqlAction.Database
	}

	// Try to substitute the server name by an IP address
	if !disable_dns_cache {
		if config, err := mysql.ParseDSN(server); err == nil {
			log.Debugf("%v", config)
			server = config.Addr
			if addr, status := utils.GetServerAddress(server); status == true {
				config.Addr = addr
				server = config.FormatDSN()
			}
		}
	}

	db, err := sql.Open(sqlAction.DBDriver, server)
	if err != nil {
		vulog.Errorf("SQL Open failed: %s", err)
		sampleReqResult := buildSampleResult(REPORTER_SQL, sessionMap["UID"], 0, reporter.NETWORK_ERROR, 0, sqlAction.Title, err.Error())
		resultsChannel <- sampleReqResult
		return false
	}
	defer db.Close()

	var start time.Time = time.Now()

	// SELECT implies Query(), other statements implies Exec()
	if strings.Index(strings.ToLower(stmt), "select") == 0 {
		rows, err := db.Query(stmt)
		if err != nil {
			vulog.Errorf("SQL Statement failed: %s: %s", stmt, err)
			sampleReqResult := buildSampleResult(REPORTER_SQL, sessionMap["UID"], 0, SQL_ERR, 0, sqlAction.Title, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
			/*
				err := rows.Scan(&id, &name)
				if err != nil {
					log.Fatal(err)
				}
				log.Println(id, name)
			*/
		}
		sessionMap[config.SQL_ROW_COUNT] = strconv.Itoa(int(count))
		err = rows.Err()
		if err != nil {
			vulog.Errorf("rows.Next() returns an error: %v", err)
		}
	} else {
		res, err := db.Exec(stmt)
		if err != nil {
			vulog.Errorf("SQL Statement failed: %s: %s", stmt, err)
			sampleReqResult := buildSampleResult(REPORTER_SQL, sessionMap["UID"], 0, SQL_ERR, 0, sqlAction.Title, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		if rowCnt, err := res.RowsAffected(); err != nil {
			sessionMap[config.SQL_ROW_COUNT] = strconv.Itoa(int(rowCnt))
		}
	}

	elapsed := time.Since(start)
	statusCode := 0

	if must_trace_request {
		vulog.Infof("%s", trace_req)
	}
	if must_display_srv_resp {
		vulog.Debugf("")
	}

	valid := true

	// if action specifies response action, parse using regexp/jsonpath
	/*
		if valid && len(response) > 0 && !processResult(sqlAction.ResponseHandlers, sessionMap, vulog, response, nil) {
			valid = false
		}
	*/

	sampleReqResult := buildSampleResult(REPORTER_SQL, sessionMap["UID"], 0, statusCode, elapsed.Nanoseconds(), sqlAction.Title, "")
	resultsChannel <- sampleReqResult
	return valid
}
