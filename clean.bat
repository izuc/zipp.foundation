@echo off
setlocal

echo Cleaning Go mod cache...
go clean -modcache

set ROOT_DIR=%~dp0
set MODULE_DIRS=ads app apputils autopeering codegen constraints core crypto ds ierrors kvstore lo logger objectstorage

for %%D in (%MODULE_DIRS%) do (
    echo.
    echo Processing directory: %ROOT_DIR%%%D\.
    echo Removing problematic entries from go.sum...
    cd %ROOT_DIR%%%D
    if exist go.sum (
        for /F "delims=" %%i in ('findstr /V "github.com/izuc/zipp.foundation" go.sum') do echo %%i >> temp.sum
        move /Y temp.sum go.sum
    )
)

echo.
echo All done.

endlocal
