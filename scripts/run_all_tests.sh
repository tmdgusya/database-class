#!/bin/bash

# Run all tests for vector-db-class project

set -e  # Exit on error

echo "=================================="
echo "  Vector DB Class - Test Runner"
echo "=================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track results
TOTAL=0
PASSED=0
FAILED=0

# Function to run tests for a directory
run_tests() {
    local dir=$1
    local name=$2

    echo -e "${YELLOW}Testing: $name${NC}"
    echo "Directory: $dir"

    if [ ! -d "$dir" ]; then
        echo -e "${RED}✗ Directory not found${NC}"
        echo ""
        return 1
    fi

    cd "$dir"

    if go test -v > /tmp/test_output.txt 2>&1; then
        local test_count=$(grep -c "^=== RUN" /tmp/test_output.txt || echo "0")
        echo -e "${GREEN}✓ All tests passed ($test_count tests)${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ Tests failed${NC}"
        echo "Last 20 lines of output:"
        tail -20 /tmp/test_output.txt
        FAILED=$((FAILED + 1))
    fi

    TOTAL=$((TOTAL + 1))
    cd - > /dev/null
    echo ""
}

# Main execution
echo "Starting test suite..."
echo ""

# Test pkg utilities
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Phase 1: Shared Utilities"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

run_tests "pkg/vector" "Vector Package"
run_tests "pkg/distance" "Distance Package"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Phase 2: Flat Index"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

run_tests "01-flat/solution" "Flat Index (Solution)"

# Check if exercise is implemented
if [ -f "01-flat/exercise/flat.go" ]; then
    if ! grep -q "panic(\"not implemented\")" "01-flat/exercise/flat.go"; then
        run_tests "01-flat/exercise" "Flat Index (Exercise)"
    else
        echo -e "${YELLOW}⊘ Flat Exercise not implemented yet${NC}"
        echo ""
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Phase 3: IVF Index"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

run_tests "02-ivf/solution" "IVF Index (Solution)"

if [ -f "02-ivf/exercise/ivf.go" ]; then
    if ! grep -q "panic(\"not implemented\")" "02-ivf/exercise/ivf.go"; then
        run_tests "02-ivf/exercise" "IVF Index (Exercise)"
    else
        echo -e "${YELLOW}⊘ IVF Exercise not implemented yet${NC}"
        echo ""
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Phase 4: HNSW Index"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ -d "03-hnsw/solution" ]; then
    run_tests "03-hnsw/solution" "HNSW Index (Solution)"
else
    echo -e "${YELLOW}⊘ HNSW Solution not created yet${NC}"
    echo ""
fi

if [ -f "03-hnsw/exercise/hnsw.go" ]; then
    if ! grep -q "panic(\"not implemented\")" "03-hnsw/exercise/hnsw.go"; then
        run_tests "03-hnsw/exercise" "HNSW Index (Exercise)"
    else
        echo -e "${YELLOW}⊘ HNSW Exercise not implemented yet${NC}"
        echo ""
    fi
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Test Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Total test suites: $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  ✓ All tests passed!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 0
else
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}  ✗ Some tests failed${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 1
fi
