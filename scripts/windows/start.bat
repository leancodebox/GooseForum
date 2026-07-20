@echo off
cd /d "%~dp0"

GooseForum.exe serve
if errorlevel 1 (
  echo.
  echo GooseForum exited with an error.
  pause
)
