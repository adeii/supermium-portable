@echo off
set APP_DIR=%~dp0\app
set DATA_DIR=%~dp0\data
"%APP_DIR%\zen.exe" -profile "%DATA_DIR%\profile1" -no-remote
