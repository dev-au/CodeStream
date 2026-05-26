#!/bin/bash

NEW_NAME=$1

if [ -z "$NEW_NAME" ]; then
    echo "Usage: ./scripts/init.sh <new-project-name>"
    exit 1
fi

# Move to project root
cd "$(dirname "$0")/.." || exit

echo "Replacing gotemplate with $NEW_NAME..."

TARGET="gotemplate"

find . -type f \
    -not -path "./.git/*" \
    -not -path "./scripts/*" \
    | xargs sed -i '' "s|$TARGET|$NEW_NAME|g" 2>/dev/null || \
find . -type f \
    -not -path "./.git/*" \
    -not -path "./scripts/*" \
    | xargs sed -i "s|$TARGET|$NEW_NAME|g"

if [ -f "go.mod" ]; then
    go mod tidy
fi

if [ -d ".git" ]; then
    echo "⚠️  Existing .git directory found. Skipping 'git init' to protect your history."
    echo "   If you want a fresh start, run 'rm -rf .git && git init' manually."
else
    echo "🌱 No .git found. Initializing new repository..."
    git init
    echo "✅ Git initialized!"
    echo "🪝 Setting up git hooks directly..."
    
    HOOK_FILE=".git/hooks/pre-commit"
    
    cat <<EOF > $HOOK_FILE
#!/bin/bash
echo "Running pre-commit checks..."
make precommit
if [ \$? -ne 0 ]; then
    echo "❌ Pre-commit checks failed. Commit aborted."
    exit 1
fi
EOF
    chmod +x $HOOK_FILE
    echo "✅ Hook installed successfully at $HOOK_FILE"

fi



echo "✅ Initialization complete!"