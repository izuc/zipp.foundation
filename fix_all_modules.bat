@echo off

REM Set the directory to start the search from (current directory in this case)
set ROOT_DIR=.

REM Use a for loop to iterate over all directories
for /R %ROOT_DIR% %%i in (.) do (
    REM Check if the directory contains a go.mod file and is not the .git directory
    if exist "%%i\go.mod" (
        echo Processing directory: %%i
        echo Fixing go.sum issues in %%i...
        cd "%%i"
        
        REM Add the missing dependencies to go.sum
        go mod tidy
        
        REM Verify the go.sum
        go mod verify
        
        echo .
    )
)

echo All directories processed.