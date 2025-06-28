@echo off
setlocal EnableDelayedExpansion

REM Aura CLI - Comprehensive Test Suite for Windows
REM This script runs all tests in the codebase with coverage reporting

echo 🧪 Aura CLI - Running Comprehensive Test Suite
echo ==============================================

REM Check if Go is installed
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go is not installed
    exit /b 1
)

echo 📋 Go Version:
go version
echo.

REM Create test output directory
if not exist test-results mkdir test-results

echo 🚀 Starting test execution...
echo.

REM Test all packages
set "packages=cmd/aura internal/ai internal/cmd internal/config internal/context internal/db assets"
set "failed_count=0"
set "passed_count=0"

for %%p in (%packages%) do (
    echo 🔍 Testing package: %%p
    
    REM Run tests with coverage
    go test -v -race -coverprofile="test-results\%%~np.out" -covermode=atomic "./%%p" 2>&1
    if %errorlevel% equ 0 (
        echo ✅ %%p tests passed
        set /a passed_count+=1
        
        REM Generate coverage report if profile exists
        if exist "test-results\%%~np.out" (
            for /f "tokens=3" %%c in ('go tool cover -func="test-results\%%~np.out" ^| findstr "total"') do (
                echo 📊 Coverage: %%c
            )
        )
    ) else (
        echo ❌ %%p tests failed
        set /a failed_count+=1
    )
    echo.
)

echo 📈 Generating combined coverage report...

REM Combine all coverage profiles
echo mode: atomic > test-results\coverage.out
for %%p in (%packages%) do (
    set "package_name=%%~np"
    if exist "test-results\!package_name!.out" (
        more +1 "test-results\!package_name!.out" >> test-results\coverage.out
    )
)

REM Generate HTML coverage report
if exist "test-results\coverage.out" (
    go tool cover -html=test-results\coverage.out -o test-results\coverage.html
    for /f "tokens=3" %%c in ('go tool cover -func=test-results\coverage.out ^| findstr "total"') do (
        echo 📊 Total Coverage: %%c
    )
    echo 📋 HTML Report: test-results\coverage.html
)

echo.
echo 🏁 Test Summary
echo ===============

echo ✅ Passed packages: %passed_count%
echo ❌ Failed packages: %failed_count%

REM Run benchmark tests
echo.
echo 🚀 Running benchmark tests...
go test -bench=. -benchmem ./... > test-results\benchmarks.txt 2>&1
if exist "test-results\benchmarks.txt" (
    echo 📋 Benchmark results saved to: test-results\benchmarks.txt
)

REM Run race detection tests
echo 🏃 Running race detection tests...
go test -race ./... > test-results\race-detection.txt 2>&1
if %errorlevel% equ 0 (
    echo ✅ No race conditions detected
) else (
    echo ⚠️  Race detection results saved to: test-results\race-detection.txt
)

REM Run vet
echo 🔍 Running go vet...
go vet ./... > test-results\vet.txt 2>&1
if %errorlevel% equ 0 (
    echo ✅ go vet passed
) else (
    echo ⚠️  go vet issues found - check test-results\vet.txt
)

REM Final result
if %failed_count% equ 0 (
    echo 🎉 All tests passed!
    exit /b 0
) else (
    echo 💥 Some tests failed. See details above.
    exit /b 1
)
