# Calendar Utility Node For Timesheets
Go based portable utility to automatically create timesheets for EPCC employees, both full time and part time. It uses fyne for GUI, sqlite for built in DB.

# Starting development
To develop, please install the go command line utils using any package manager you wish. Once installed, you will need the following:

- `go` language tools
- `MinGW64` (for windows compilation of gui)
- `gcc` (c compiler, mainly for gui compilation)
- `libgl1-mesa-dev`, `xorg-dev` (linux dev)
- `xcode-select --install` (mac devs only)

# Structure:
## Database operations
- All DB operations must be done through SQL queries for SQLite.
- All DB operations are to be done on the [`db`](./db/) directory.
- Check the table queries in [`repository.go`](./db/repository.go) to understand how data is stored.
- For more details on how data is modeled, please check out the go files in the [`models`](./models/) directory.

## UI
- All UI operations are done using the fyne UI framework and should be done in the [`gui`](./gui/) directory.
- Every tab of the application should be in its individual file (profile, calendar, etc)
- All the tabs shuld be finally appended to [`main.go`](./main.go)

# Compilation
## For local testing
Type in terminal
```bash
go run main.go
```
If there's any issues (especially windows) try this in the `MinGW` terminal

## For version release
- Create a branch with the following naming convention
```bash
branch version0.1.0 or tag v0.1.0
```
Once done, push to the origin repo. Wait a couple of minutes for the compilation finishes. After so, you can go to releases (right hand side of the main page of the repo (next to README)) and download it there.