# toggl-overtime

CLI tool that calculates overtime from time tracked Toggl app.

# Installation

```bash
go get -u github.com/ditrytus/toggl-overtime
```

# Usage

```
toggl-overtime --help 
```
```
Calculates overtime based on time logged with Toggl app.

Usage:
  toggl-overtime [flags]

Flags:
      --config string                config file (default is $HOME/.toggl-overtime.yaml)
  -e, --end-date string              end date of period for which to calculate overtime (default "2020-01-28 23:59:59.999999999 +0100 CET")
  -x, --exclude-dates strings        list of dates that shouldn't be count as working days
  -h, --help                         help for toggl-overtime
  -n, --non-working-days strings     list of days of the week that shouldn't be count as working days (default [Saturday,Sunday])
  -p, --projects strings             Toggl projects (default [Work])
  -s, --start-date string            start date of period for which to calculate overtime (default "2020-01-01 00:00:00 +0100 CET")
  -t, --token string                 Toggl API authorization token
      --version                      version for toggl-overtime
  -d, --work-day-duration duration   base duration of a working day (default 8h0m0s)
  -w, --workspace string             Toggl workspace (default "Personal")
```

Example output:

```
11h37m8s
```
# Sample config file

`~/.toggl-overtime.yaml`
```yaml
token: abbaba2823183bb123123
exclude-dates:
    - 2020-01-01
    - 2020-01-06
```