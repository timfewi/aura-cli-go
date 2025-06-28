# Git Style Guide - Quick Reference

## Introduction

Essential Git conventions for version control in Python projects. These guidelines ensure clean commit history, effective collaboration, and maintainable code versioning.

**Key Principle**: Clear, descriptive commits tell the story of your project's evolution.

## Commit Messages

### Format Structure

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature for the user
- `fix`: Bug fix for the user
- `docs`: Documentation changes
- `style`: Code formatting (no production code change)
- `refactor`: Code restructuring without changing external behavior
- `test`: Adding or updating tests
- `chore`: Maintenance tasks, dependency updates

### Examples

```bash
# Good commit messages
feat(scraper): add price extraction for Amazon products
fix(parser): handle missing product description gracefully
docs(readme): update installation instructions
refactor(utils): extract common validation functions
test(scraper): add unit tests for rate limiting
chore(deps): update requests to v2.28.0

# Poor commit messages
fix stuff
update
changes
wip
```

### Best Practices

- Use present tense: "add feature" not "added feature"
- Keep subject line under 50 characters
- Capitalize subject line
- No period at end of subject line
- Use body to explain what and why, not how
- Separate subject from body with blank line

## Branching Strategy

### Branch Naming

```bash
# Feature branches
feature/amazon-product-scraper
feature/rate-limiting-system
feature/data-validation

# Bug fix branches
fix/price-parsing-error
fix/memory-leak-selenium
hotfix/critical-timeout-issue

# Documentation branches
docs/api-documentation
docs/setup-guide

# Maintenance branches
chore/update-dependencies
chore/cleanup-unused-imports
```

### Branch Types

- `main/master`: Production-ready code
- `develop`: Integration branch for features
- `feature/*`: New functionality
- `fix/*`: Bug fixes
- `hotfix/*`: Critical fixes for production
- `docs/*`: Documentation updates
- `chore/*`: Maintenance tasks

### Workflow

```bash
# Create feature branch from main
git checkout main
git pull origin main
git checkout -b feature/new-scraper

# Work on feature, commit regularly
git add .
git commit -m "feat(scraper): implement basic structure"

# Push branch and create PR
git push origin feature/new-scraper

# After PR approval, merge and cleanup
git checkout main
git pull origin main
git branch -d feature/new-scraper
```

## File Management

### .gitignore Best Practices

```gitignore
# Python
__pycache__/
*.pyc
*.pyo
*.egg-info/
.venv/
.pytest_cache/

# Scraped data (project-specific)
data/
output/
scraped_data/
*.csv
*.json

# Logs and temporary files
*.log
*.tmp
.DS_Store

# IDE files
.vscode/
.idea/
*.swp

# Environment variables
.env
.env.local
```

### What to Track

- ✅ Source code (.py files)
- ✅ Configuration files
- ✅ Documentation
- ✅ Requirements/dependency files
- ✅ Tests
- ✅ README and setup instructions

### What NOT to Track

- ❌ Compiled files (**pycache**, \*.pyc)
- ❌ Virtual environments
- ❌ IDE-specific files
- ❌ Scraped data/output files
- ❌ Log files
- ❌ Credentials/secrets
- ❌ Large binary files

## Daily Workflow

### Starting Work

```bash
# Update local main branch
git checkout main
git pull origin main

# Create or switch to feature branch
git checkout -b feature/task-name
# or
git checkout feature/existing-branch
git pull origin feature/existing-branch
```

### During Development

```bash
# Check status frequently
git status

# Stage specific files (preferred over git add .)
git add src/scraper.py
git add tests/test_scraper.py

# Commit with descriptive message
git commit -m "feat(scraper): add retry mechanism for failed requests"

# Push regularly to backup work
git push origin feature/task-name
```

### Code Review Process

```bash
# Before creating PR
git fetch origin
git rebase origin/main  # Keep history clean

# Create PR through GitHub/GitLab interface
# Address review feedback
git add .
git commit -m "fix(scraper): address code review comments"
git push origin feature/task-name
```

## Advanced Git Practices

### Interactive Rebase

```bash
# Clean up commit history before PR
git rebase -i HEAD~3

# Squash related commits
# Edit commit messages
# Reorder commits if needed
```

### Stashing Changes

```bash
# Save work in progress
git stash push -m "WIP: debugging rate limit issue"

# Switch branches, do other work
git checkout main
git pull origin main

# Return to work
git checkout feature/branch
git stash pop
```

### Cherry-picking

```bash
# Apply specific commit to current branch
git cherry-pick abc123def

# Apply commit with new message
git cherry-pick abc123def --edit
```

## Web Scraping Specific Guidelines

### Handling Scraped Data

```bash
# Never commit scraped data to repository
echo "data/" >> .gitignore
echo "output/" >> .gitignore
echo "*.csv" >> .gitignore

# Use data directories with .gitkeep for structure
mkdir -p data/raw data/processed
touch data/raw/.gitkeep data/processed/.gitkeep
git add data/raw/.gitkeep data/processed/.gitkeep
```

### Configuration Management

```bash
# Track template configs, not actual configs
git add config.template.json
git ignore config.json

# Use environment variables for secrets
echo ".env" >> .gitignore
# Track .env.example instead
git add .env.example
```

### Version Tagging

```bash
# Tag releases for scraper versions
git tag -a v1.0.0 -m "Initial scraper release"
git tag -a v1.1.0 -m "Add support for multiple sites"
git push origin --tags
```

## Collaboration Best Practices

### Pull Request Guidelines

- Keep PRs focused and small
- Write descriptive PR titles and descriptions
- Include testing notes
- Link to relevant issues
- Request appropriate reviewers
- Respond to feedback promptly

### Code Review Checklist

- ✅ Code follows project style guide (PEP 8)
- ✅ Tests are included and passing
- ✅ Documentation is updated
- ✅ No sensitive data is committed
- ✅ Performance implications considered
- ✅ Error handling is appropriate

### Merge Strategies

```bash
# For feature branches (preferred)
git merge --no-ff feature/branch

# For hotfixes (fast-forward)
git merge feature/hotfix

# Squash merge for cleanup
git merge --squash feature/small-fix
```

## Troubleshooting Common Issues

### Fixing Commit Messages

```bash
# Fix last commit message
git commit --amend -m "fix(scraper): correct typo in error handling"

# Fix older commit messages
git rebase -i HEAD~3
# Change 'pick' to 'reword' for commits to edit
```

### Undoing Changes

```bash
# Unstage files
git reset HEAD file.py

# Discard working directory changes
git checkout -- file.py

# Undo last commit (keep changes)
git reset --soft HEAD~1

# Undo last commit (discard changes)
git reset --hard HEAD~1
```

### Resolving Merge Conflicts

```bash
# When conflicts occur during merge/rebase
git status  # Shows conflicted files

# Edit files to resolve conflicts
# Look for <<<<<<< ======= >>>>>>> markers

# After resolving conflicts
git add conflicted-file.py
git commit  # For merge
# or
git rebase --continue  # For rebase
```

## Git Hooks for Web Scraping

### Pre-commit Hook Example

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Check for secrets in commits
if git diff --cached --name-only | xargs grep -l "api_key\|password\|secret"; then
    echo "Error: Potential secrets found in staged files"
    exit 1
fi

# Run linting
python -m flake8 src/
if [ $? -ne 0 ]; then
    echo "Error: Code style violations found"
    exit 1
fi
```

## Quick Reference Commands

### Daily Commands

```bash
git status                    # Check working directory status
git add file.py              # Stage specific file
git commit -m "message"      # Commit with message
git push origin branch       # Push to remote
git pull origin main         # Pull latest changes
git checkout -b new-branch   # Create and switch to branch
git merge branch-name        # Merge branch
git log --oneline            # View commit history
```

### Emergency Commands

```bash
git stash                    # Save work quickly
git reset --hard HEAD        # Discard all changes
git clean -fd               # Remove untracked files
git reflog                  # View reference history
git bisect start            # Find problematic commit
```

## Quick Checklist

### Before Committing

- ✅ Code is tested and working
- ✅ No sensitive data included
- ✅ Commit message is descriptive
- ✅ Only relevant files are staged
- ✅ Code follows project style guide

### Before Pushing

- ✅ Local branch is up to date
- ✅ All tests pass
- ✅ No merge conflicts
- ✅ Commit history is clean
- ✅ PR/issue references included

### Repository Health

- ✅ .gitignore is comprehensive
- ✅ README is up to date
- ✅ Dependencies are documented
- ✅ No large files in history
- ✅ Regular cleanup of old branches

---

_Use clear, descriptive commits. Keep branches focused. Never commit secrets or scraped data. Collaborate effectively through pull requests._
