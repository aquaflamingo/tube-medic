#!/bin/bash

# Exit immediately if any command fails
set -e

# ==========================================
# GLOBAL CONFIGURATION
# ==========================================
MONOREPO_PATH="/path/to/your/monorepo"
BRANCH_NAME="main" # Change to 'master' if your repos use it
# ==========================================

# ------------------------------------------
# HOMEBREW-ONLY INSTALLATION HARNESS
# ------------------------------------------
ensure_filter_repo_installed() {
    if ! command -v git-filter-repo &> /dev/null; then
        echo "🔍 'git-filter-repo' is not installed."
        echo -n "Would you like this script to attempt installing it via Homebrew? (y/n): "
        read -r user_choice

        case "$user_choice" in 
            [Yy]* ) 
                if command -v brew &> /dev/null; then
                    echo "🍺 Homebrew detected. Running: brew install git-filter-repo"
                    brew install git-filter-repo
                else
                    echo "❌ Error: Homebrew ('brew') is not installed or not in your PATH."
                    echo "Please install Homebrew first (https://brew.sh) or install git-filter-repo manually."
                    exit 1
                fi
                
                # Double-check after installation
                if ! command -v git-filter-repo &> /dev/null; then
                    echo "❌ Homebrew command finished, but 'git-filter-repo' is still missing."
                    exit 1
                fi
                echo "✅ Installation successful!"
                ;;
            * ) 
                echo "❌ Blocked. The script cannot continue without git-filter-repo."
                exit 1
                ;;
        esac
    else
        echo "✅ 'git-filter-repo' is already installed."
    fi
}

# Run the installation harness check
ensure_filter_repo_installed

# Ensure monorepo exists and is clean
if [ ! -d "$MONOREPO_PATH/.git" ]; then
    echo "❌ Error: $MONOREPO_PATH is not a valid Git repository."
    exit 1
fi

cd "$MONOREPO_PATH"
if [[ -n $(git status --porcelain) ]]; then
    echo "⚠️ Warning: Your monorepo has uncommitted changes. Please commit or stash them first."
    exit 1
fi

# ------------------------------------------
# CORE MIGRATION FUNCTION
# ------------------------------------------
merge_subrepo() {
    local sub_repo_url=$1
    local sub_repo_name=$2
    local target_dir=$3

    echo "--------------------------------------------------"
    echo "🔄 Starting migration for: $sub_repo_name"
    echo "--------------------------------------------------"

    # Create a secure temporary directory
    local temp_dir
    temp_dir=$(mktemp -d)
    
    # 1. Clone sub-repo
    echo "🚀 1/4 Cloning $sub_repo_name..."
    git clone "$sub_repo_url" "$temp_dir/$sub_repo_name"
    cd "$temp_dir/$sub_repo_name"

    # 2. Rewrite history
    echo "🪄 2/4 Rewriting history to '$target_dir'..."
    git filter-repo --to-subdirectory-filter "$target_dir"

    # 3. Fetch and Merge into Monorepo
    echo "📦 3/4 Merging history into monorepo..."
    cd "$MONOREPO_PATH"
    
    git remote add "temp-remote-$sub_repo_name" "$temp_dir/$sub_repo_name"
    git fetch "temp-remote-$sub_repo_name"

    git merge "temp-remote-$sub_repo_name/$BRANCH_NAME" --allow-unrelated-histories --no-edit -m "chore: merge $sub_repo_name history into $target_dir"

    git remote remove "temp-remote-$sub_repo_name"

    # 4. Cleanup
    echo "🧹 4/4 Cleaning up temp files..."
    rm -rf "$temp_dir"

    echo "✅ Successfully merged $sub_repo_name!"
    echo ""
}

# ==========================================
# EXECUTION - Define your 2 repos here!
# ==========================================

# Format: merge_subrepo "REPO_URL" "REPO_NAME" "TARGET_FOLDER_IN_MONOREPO"

merge_subrepo \
    "https://github.com/username/first-subrepo.git" \
    "first-subrepo" \
    "packages/first-subrepo"

merge_subrepo \
    "https://github.com/username/second-subrepo.git" \
    "second-subrepo" \
    "packages/second-subrepo"


echo "🎉 All repositories have been successfully merged into your monorepo!"
