@echo off
cd /d "%~dp0"
docker compose run --build --rm heimdall %*
exit /b %ERRORLEVEL%
