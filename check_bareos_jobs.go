/*
  check_bareos_jobs - Checks the last status of bareos jobs within a defined time period
  Copyright (C) 2020  Armin Bube

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
  "flag"
  "fmt"
  "time"
  "os"
  "strconv"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

// Global variables
var programName string = "check_bareos_jobs"
var version string = "0.0.1"
var exitCode int = 0
var dbType string
var dbHostname string
var dbPort string
var dbName string
var dbUser string
var dbPassword string
var licenseFlag bool
var versionFlag bool

type JobStatus struct {
  Id int `json:"JobId"`
  Name string `json:"Name"`
  Status string `json:"JobStatus"`
  StatusLong string `json:"JobStatusLong"`
  Severity int `json:"Severity"`
}

func defineParameter() {
  // Fields are name, default and description
  flag.StringVar(&dbType, "type", "mysql", "Bareos Database Type")
  flag.StringVar(&dbHostname, "host", "127.0.0.1", "Bareos Database Hostname or IP")
  flag.StringVar(&dbPort, "port", "3306", "Bareos Database Port")
  flag.StringVar(&dbName, "database", "bareos", "Bareos Database Name")
  flag.StringVar(&dbUser, "user", "bareos", "Bareos Database Username")
  flag.StringVar(&dbPassword, "password", "bareos", "Bareos Database Password")
  flag.BoolVar(&licenseFlag, "license", false, "Displays license information and quits")
  flag.BoolVar(&versionFlag, "version", false, "Displays version information and quits")

  flag.Usage = func() {
    fmt.Fprintf(os.Stderr, `
check_bareos_jobs  Copyright (C) 2020  Armin Bube
This program comes with ABSOLUTELY NO WARRANTY; for details use flag -license.
This is free software, and you are welcome to redistribute it
under certain conditions; for details use flag -license.
`)
    fmt.Fprintf(os.Stderr, "\nUsage of bin/check_bareos_jobs:\n\n")
    flag.PrintDefaults()
  }

  flag.Parse()

  if licenseFlag {
    showLicense()
  }

  if versionFlag {
    showVersion()
  }
}

func abort(message string, ec int) {
  fmt.Println("[ERROR] "+message)
  os.Exit(ec)
}

func showLicense() {
  fmt.Println(`
GNU General Public License v3.0
You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
`)
  os.Exit(0)
}

func showVersion() {
  fmt.Println(version)
  os.Exit(0)
}

func queryJobStatusList(databaseType string) *sql.Rows {
  var results *sql.Rows
  var query string = "SELECT job.JobId, job.Name, job.JobStatus, status.JobStatusLong, status.Severity FROM Job job INNER JOIN Status status ON status.JobStatus = job.JobStatus WHERE EndTime BETWEEN DATE_SUB(NOW(), INTERVAL 25 HOUR) AND NOW() AND job.JobId IN (SELECT MAX(JobId) FROM Job WHERE Name = job.Name) ORDER BY status.Severity DESC"
  switch databaseType {
    case "mysql", "MySQL":
      db, err := sql.Open("mysql", dbUser+":"+dbPassword+"@tcp("+dbHostname+":"+dbPort+")/"+dbName)

      if err != nil {
        abort("Could not connect to database: "+err.Error(), 11)
      }

      db.SetConnMaxLifetime(time.Minute * 3)
      db.SetMaxOpenConns(1)
      db.SetMaxIdleConns(1)

      defer db.Close()

      // Query job information of the last 25 hours
      var queryError error
      results, queryError = db.Query(query)

      if queryError != nil {
        abort("Could not query sql: "+queryError.Error(), 12)
      }
    case "postgres", "PostgreSQL", "postgresql":
      abort("PostgreSQL not supported", 200)
    default:
      abort("Database type unknown", 10)
  }

  return results
}

func processStatusInformation(jobStatusList *sql.Rows) (string, []string, string) {
  JobCount := 0
  OKJobCount := 0
  CriticalJobCount := 0
  WarningJobCount := 0
  extendedStatusInformation := []string{"extendedStatusInformation"}

  for jobStatusList.Next() {
    var jobstatus JobStatus
    // For each row, scan the result into the JobStatus object
    err := jobStatusList.Scan(&jobstatus.Id, &jobstatus.Name, &jobstatus.Status, &jobstatus.StatusLong, &jobstatus.Severity)
    if err != nil {
      abort("Results have different scheme than expected: "+err.Error(), 20)
    }

    // Check if jobstatus is ok, warning or critical
    // To check the meaning of the severity number query your Bareos database with "SELECT * FROM Status ORDER BY Severity ASC;"

    if jobstatus.Severity >= 25 {
      CriticalJobCount = CriticalJobCount + 1

      if extendedStatusInformation[0] == "extendedStatusInformation" {
        extendedStatusInformation[0] = "[CRITICAL] "+jobstatus.Name+": "+jobstatus.StatusLong
      } else {
        extendedStatusInformation = append(extendedStatusInformation, "[CRITICAL] "+jobstatus.Name+": "+jobstatus.StatusLong)
      }
    } else if jobstatus.Severity >= 15 {
      WarningJobCount = WarningJobCount + 1

      if extendedStatusInformation[0] == "extendedStatusInformation" {
        extendedStatusInformation[0] = "[WARNING] "+jobstatus.Name+": "+jobstatus.StatusLong
      } else {
        extendedStatusInformation = append(extendedStatusInformation, "[WARNING] "+jobstatus.Name+": "+jobstatus.StatusLong)
      }
    } else {
      OKJobCount = OKJobCount +1

      if extendedStatusInformation[0] == "extendedStatusInformation" {
        extendedStatusInformation[0] = "[OK] "+jobstatus.Name+": "+jobstatus.StatusLong
      } else {
        extendedStatusInformation = append(extendedStatusInformation, "[OK] "+jobstatus.Name+": "+jobstatus.StatusLong)
      }
    }
    JobCount = JobCount +1
  }

  statusInformation := "Jobs: "+strconv.Itoa(JobCount)+" OK: "+strconv.Itoa(OKJobCount)+" WARNING: "+strconv.Itoa(WarningJobCount)+" CRITICAL: "+strconv.Itoa(CriticalJobCount)
  performanceData := "|jobs="+strconv.Itoa(JobCount)+" ok="+strconv.Itoa(OKJobCount)+" warning="+strconv.Itoa(WarningJobCount)+" critical="+strconv.Itoa(CriticalJobCount)

  // Define the exit code related to job count and status
  if CriticalJobCount > 0 {
    exitCode = 2
  } else if WarningJobCount > 0 {
    exitCode = 1
  } else if JobCount == 0 {
    exitCode = 3
  } else {
    exitCode = 0
  }

  return statusInformation, extendedStatusInformation, performanceData
}

func printResults(statusInformation string, extendedStatusInformation []string, performanceData string) {
  // Print main status information
  fmt.Println(statusInformation)

  // Print extended status information
  if extendedStatusInformation[0] != "extendedStatusInformation" {
    for extendedStatusCount := 0; extendedStatusCount < len(extendedStatusInformation); extendedStatusCount++ {
      fmt.Println(extendedStatusInformation[extendedStatusCount])
    }
  }

  // Print performance data
  fmt.Println(performanceData)
}

func main() {
  defineParameter()
  statusInformation, extendedStatusInformation, performanceData := processStatusInformation(queryJobStatusList(dbType))
  printResults(statusInformation, extendedStatusInformation, performanceData)
  os.Exit(exitCode)
}
