#!/bin/bash

# Aura CLI - Comprehensive Test Suite
# This script runs all tests in the codebase with coverage reporting

set -e

echo "🧪 Aura CLI - Running Comprehensive Test Suite"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

echo -e "${BLUE}📋 Go Version:${NC}"
go version
echo

# Create test output directory
mkdir -p test-results

# Function to run tests for a specific package
run_package_tests() {
    local package=$1
    local package_name=$(basename $package)
    
    echo -e "${BLUE}🔍 Testing package: ${package_name}${NC}"
    
    # Run tests with coverage
    if go test -v -race -coverprofile="test-results/${package_name}.out" -covermode=atomic "./$package" 2>&1; then
        echo -e "${GREEN}✅ ${package_name} tests passed${NC}"
        
        # Generate coverage report
        if [ -f "test-results/${package_name}.out" ]; then
            coverage=$(go tool cover -func="test-results/${package_name}.out" | grep total | awk '{print $3}')
            echo -e "${BLUE}📊 Coverage: ${coverage}${NC}"
        fi
    else
        echo -e "${RED}❌ ${package_name} tests failed${NC}"
        return 1
    fi
    echo
}

# Function to run tests with timeout
run_tests_with_timeout() {
    timeout 300s "$@"
    if [ $? -eq 124 ]; then
        echo -e "${RED}❌ Tests timed out after 5 minutes${NC}"
        return 1
    fi
}

echo -e "${YELLOW}🚀 Starting test execution...${NC}"
echo

# Test all packages
packages=(
    "cmd/aura"
    "internal/ai"
    "internal/cmd" 
    "internal/config"
    "internal/context"
    "internal/db"
    "assets"
)

failed_packages=()
passed_packages=()

for package in "${packages[@]}"; do
    if run_tests_with_timeout run_package_tests "$package"; then
        passed_packages+=("$package")
    else
        failed_packages+=("$package")
    fi
done

echo -e "${BLUE}📈 Generating combined coverage report...${NC}"

# Combine all coverage profiles
echo "mode: atomic" > test-results/coverage.out
for package in "${packages[@]}"; do
    package_name=$(basename $package)
    if [ -f "test-results/${package_name}.out" ]; then
        tail -n +2 "test-results/${package_name}.out" >> test-results/coverage.out
    fi
done

# Generate HTML coverage report
if [ -f "test-results/coverage.out" ]; then
    go tool cover -html=test-results/coverage.out -o test-results/coverage.html
    total_coverage=$(go tool cover -func=test-results/coverage.out | grep total | awk '{print $3}')
    echo -e "${GREEN}📊 Total Coverage: ${total_coverage}${NC}"
    echo -e "${BLUE}📋 HTML Report: test-results/coverage.html${NC}"
fi

echo
echo "🏁 Test Summary"
echo "==============="

if [ ${#passed_packages[@]} -gt 0 ]; then
    echo -e "${GREEN}✅ Passed packages (${#passed_packages[@]}):${NC}"
    for package in "${passed_packages[@]}"; do
        echo -e "   ${GREEN}• ${package}${NC}"
    done
    echo
fi

if [ ${#failed_packages[@]} -gt 0 ]; then
    echo -e "${RED}❌ Failed packages (${#failed_packages[@]}):${NC}"
    for package in "${failed_packages[@]}"; do
        echo -e "   ${RED}• ${package}${NC}"
    done
    echo
fi

# Run benchmark tests
echo -e "${BLUE}🚀 Running benchmark tests...${NC}"
go test -bench=. -benchmem ./... > test-results/benchmarks.txt 2>&1 || true
if [ -f "test-results/benchmarks.txt" ]; then
    echo -e "${BLUE}📋 Benchmark results saved to: test-results/benchmarks.txt${NC}"
fi

# Run race detection tests
echo -e "${BLUE}🏃 Running race detection tests...${NC}"
go test -race ./... > test-results/race-detection.txt 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ No race conditions detected${NC}"
else
    echo -e "${YELLOW}⚠️  Race detection results saved to: test-results/race-detection.txt${NC}"
fi

# Run vet
echo -e "${BLUE}🔍 Running go vet...${NC}"
go vet ./... > test-results/vet.txt 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ go vet passed${NC}"
else
    echo -e "${YELLOW}⚠️  go vet issues found - check test-results/vet.txt${NC}"
fi

# Final result
if [ ${#failed_packages[@]} -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}💥 Some tests failed. See details above.${NC}"
    exit 1
fi
