# check_bareos_jobs

## Description
It is a monitoring check for Nagios / NRPE to read the status of the last finished bareos jobs.
This check connects to the database and returns the abount of jobs sorted by status.

Currently onyl MySQL databases are supported but PostgreSQL support will follow soon.

## Usage

```
check_bareos_jobs -host <dbhost> -database <dbname> -user <dbusername> -password <dbpassword>
```

Check help for all options
```
check_bareos_jobs -h

check_bareos_jobs  Copyright (C) 2020  Armin Bube
This program comes with ABSOLUTELY NO WARRANTY; for details use flag -license.
This is free software, and you are welcome to redistribute it
under certain conditions; for details use flag -license.

Usage of bin/check_bareos_jobs:

  -database string
    	Bareos Database Name (default "bareos")
  -host string
    	Bareos Database Hostname or IP (default "127.0.0.1")
  -license
    	Displays license information and quits
  -password string
    	Bareos Database Password (default "bareos")
  -port string
    	Bareos Database Port (default "3306")
  -type string
    	Bareos Database Type (default "mysql")
  -user string
    	Bareos Database Username (default "bareos")
  -version
    	Displays version information and quits
```

### Example

```
check_bareos_jobs -host localhost -database bareos -user bareos_ro -password securePassword
```

## Author

Armin Bube
