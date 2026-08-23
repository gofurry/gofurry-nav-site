@echo off
setlocal EnableExtensions

set "ROOT=%~dp0"
if "%ROOT:~-1%"=="\" set "ROOT=%ROOT:~0,-1%"
set "BUILD_ROOT=%ROOT%\build"
set "TARGET=%~1"
if "%TARGET%"=="" set "TARGET=all"

set "GOOS=linux"
set "GOARCH=amd64"
set "CGO_ENABLED=0"

if not exist "%BUILD_ROOT%" mkdir "%BUILD_ROOT%"

if /I "%TARGET%"=="all" goto build_all
if /I "%TARGET%"=="nav-backend" goto build_nav_backend
if /I "%TARGET%"=="nav-collector" goto build_nav_collector
if /I "%TARGET%"=="game-backend" goto build_game_backend
if /I "%TARGET%"=="game-collector" goto build_game_collector
if /I "%TARGET%"=="admin" goto build_admin
if /I "%TARGET%"=="gofurry-nav-backend" goto build_nav_backend
if /I "%TARGET%"=="gofurry-nav-collector" goto build_nav_collector
if /I "%TARGET%"=="gofurry-game-backend" goto build_game_backend
if /I "%TARGET%"=="gofurry-game-collector" goto build_game_collector
if /I "%TARGET%"=="gofurry-admin" goto build_admin

echo Unknown target: %TARGET%
echo Supported targets:
echo   all
echo   nav-backend
echo   nav-collector
echo   game-backend
echo   game-collector
echo   admin
exit /b 1

:build_all
call "%~f0" nav-backend || exit /b 1
call "%~f0" nav-collector || exit /b 1
call "%~f0" game-backend || exit /b 1
call "%~f0" game-collector || exit /b 1
call "%~f0" admin || exit /b 1
echo Build completed. Artifacts are in "%BUILD_ROOT%".
exit /b 0

:build_nav_backend
echo [BUILD] gofurry-nav-backend
set "OUTPUT_DIR=%BUILD_ROOT%\gf-nav"
set "OUTPUT_BIN=%OUTPUT_DIR%\gf-nav"
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
mkdir "%OUTPUT_DIR%" || exit /b 1
pushd "%ROOT%\apps\cn\nav-backend" || exit /b 1
go build -trimpath -ldflags="-s -w" -o "%OUTPUT_BIN%" .
if errorlevel 1 (
    popd
    exit /b 1
)
popd
exit /b 0

:build_nav_collector
echo [BUILD] gofurry-nav-collector
set "OUTPUT_DIR=%BUILD_ROOT%\gf-nav-collector"
set "OUTPUT_BIN=%OUTPUT_DIR%\gf-nav-collector"
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
mkdir "%OUTPUT_DIR%" || exit /b 1
pushd "%ROOT%\apps\cn\nav-collector" || exit /b 1
go build -trimpath -ldflags="-s -w" -o "%OUTPUT_BIN%" .
if errorlevel 1 (
    popd
    exit /b 1
)
popd
exit /b 0

:build_game_backend
echo [BUILD] gofurry-game-backend
set "OUTPUT_DIR=%BUILD_ROOT%\gf-game"
set "OUTPUT_BIN=%OUTPUT_DIR%\gf-game"
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
mkdir "%OUTPUT_DIR%" || exit /b 1
pushd "%ROOT%\apps\cn\game-backend" || exit /b 1
go build -trimpath -ldflags="-s -w" -o "%OUTPUT_BIN%" .
if errorlevel 1 (
    popd
    exit /b 1
)
popd
exit /b 0

:build_game_collector
echo [BUILD] gofurry-game-collector
set "OUTPUT_DIR=%BUILD_ROOT%\gf-game-collector"
set "OUTPUT_BIN=%OUTPUT_DIR%\gf-game-collector"
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
mkdir "%OUTPUT_DIR%" || exit /b 1
pushd "%ROOT%\apps\cn\game-collector" || exit /b 1
go build -trimpath -ldflags="-s -w" -o "%OUTPUT_BIN%" .
if errorlevel 1 (
    popd
    exit /b 1
)
popd
exit /b 0

:build_admin
echo [BUILD] gofurry-admin web
pushd "%ROOT%\apps\cn\admin\web" || exit /b 1
if not exist "node_modules" (
    call npm ci
    if errorlevel 1 (
        popd
        exit /b 1
    )
)
call npm run build
if errorlevel 1 (
    popd
    exit /b 1
)
popd

echo [BUILD] gofurry-admin binary
set "OUTPUT_DIR=%BUILD_ROOT%\gofurry-admin"
set "OUTPUT_BIN=%OUTPUT_DIR%\gofurry-admin"
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
mkdir "%OUTPUT_DIR%" || exit /b 1
pushd "%ROOT%\apps\cn\admin" || exit /b 1
go build -trimpath -ldflags="-s -w" -o "%OUTPUT_BIN%" .
if errorlevel 1 (
    popd
    exit /b 1
)
popd
if not exist "%ROOT%\apps\cn\admin\internal\transport\http\webui\dist" (
    echo Source directory not found: "%ROOT%\apps\cn\admin\internal\transport\http\webui\dist"
    exit /b 1
)
mkdir "%BUILD_ROOT%\gofurry-admin\dist" || exit /b 1
xcopy "%ROOT%\apps\cn\admin\internal\transport\http\webui\dist\*" "%BUILD_ROOT%\gofurry-admin\dist\" /E /I /Y >nul
if errorlevel 1 exit /b 1
exit /b 0
