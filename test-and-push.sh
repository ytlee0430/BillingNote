#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Running All Tests ===${NC}"

# Run Backend Tests
echo -e "\n${YELLOW}[1/2] Running Backend Tests...${NC}"
cd backend
if go test ./... -v; then
    echo -e "${GREEN}Backend tests PASSED${NC}"
else
    echo -e "${RED}Backend tests FAILED - aborting${NC}"
    exit 1
fi
cd ..

# Run Frontend Tests
echo -e "\n${YELLOW}[2/2] Running Frontend Tests...${NC}"
cd frontend
if npm test -- --run; then
    echo -e "${GREEN}Frontend tests PASSED${NC}"
else
    echo -e "${RED}Frontend tests FAILED - aborting${NC}"
    exit 1
fi
cd ..

echo -e "\n${GREEN}=== All Tests PASSED ===${NC}"

# Check if there are changes to commit
if git diff --quiet && git diff --staged --quiet; then
    echo -e "${YELLOW}No changes to commit${NC}"
    exit 0
fi

# Auto commit and push
echo -e "\n${YELLOW}=== Auto Commit & Push ===${NC}"

# Get commit message from argument or use default
COMMIT_MSG="${1:-Auto commit: tests passed}"

git add -A
git commit -m "$COMMIT_MSG

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push

echo -e "\n${GREEN}=== Successfully committed and pushed! ===${NC}"
