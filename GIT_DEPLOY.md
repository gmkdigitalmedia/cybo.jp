# Deploying to GitHub

## GitHub Repository Setup

Your repository: https://github.com/gmkdigitalmedia/cybo.jp.git

## Initial Upload Commands

```bash
# Navigate to your project directory
cd /mnt/c/Users/ibm/Documents/cybo

# Initialize git repository (if not already done)
git init

# Add the remote repository
git remote add origin https://github.com/gmkdigitalmedia/cybo.jp.git

# Stage all files
git add .

# Create initial commit
git commit -m "Initial commit: Professional cytology viewer with Xupra.ai x Cybo.jp branding"

# Push to GitHub
git push -u origin main

# If the above fails with "main" branch, try "master"
git push -u origin master
```

## Subsequent Updates

After making changes to your code:

```bash
# Check what files have changed
git status

# Stage all changes
git add .

# Commit with a descriptive message
git commit -m "Description of your changes"

# Push to GitHub
git push
```

## Specific Update Examples

### After UI Changes
```bash
git add web/
git commit -m "Update viewer UI and styling"
git push
```

### After Backend Changes
```bash
git add internal/ cmd/ pkg/
git commit -m "Update backend API and processing"
git push
```

### After Documentation Updates
```bash
git add *.md
git commit -m "Update documentation"
git push
```

## Troubleshooting

### If you get authentication errors:

1. **Use Personal Access Token (PAT)**:
   - Go to GitHub Settings > Developer Settings > Personal Access Tokens
   - Generate new token with "repo" permissions
   - Use the token as your password when prompted

2. **Configure Git credentials**:
```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

### If remote already exists:
```bash
git remote remove origin
git remote add origin https://github.com/gmkdigitalmedia/cybo.jp.git
```

### If you need to force push (use carefully):
```bash
git push -f origin main
```

## Creating a .gitignore

Create a file named `.gitignore` in your project root:

```
# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
env/
venv/
ENV/
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg

# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
bin/
vendor/

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Logs
*.log

# Environment
.env
config.env

# Data
data/
*.db
*.sqlite
```

## Branch Strategy

### Create a development branch:
```bash
git checkout -b development
git push -u origin development
```

### Switch between branches:
```bash
# Switch to main
git checkout main

# Switch to development
git checkout development
```

### Merge development into main:
```bash
git checkout main
git merge development
git push
```

## Quick Reference

```bash
# Clone repository (for new machine)
git clone https://github.com/gmkdigitalmedia/cybo.jp.git

# Pull latest changes
git pull

# View commit history
git log --oneline

# View current branch
git branch

# Discard local changes
git checkout -- filename.ext

# View differences
git diff
```

## Complete Workflow Example

```bash
# 1. Make changes to your files
# ... edit files ...

# 2. Check what changed
git status

# 3. Stage changes
git add .

# 4. Commit with message
git commit -m "Add login page with Xupra.ai x Cybo.jp branding"

# 5. Push to GitHub
git push

# Done!
```

## GitHub Pages Setup (Optional)

If you want to host the static files on GitHub Pages:

```bash
# Create gh-pages branch
git checkout -b gh-pages

# Push to gh-pages
git push -u origin gh-pages

# Go to GitHub repository settings
# Enable GitHub Pages from gh-pages branch
# Your site will be at: https://gmkdigitalmedia.github.io/cybo.jp/
```

## Notes

- Always pull before making changes if working from multiple locations
- Write clear, descriptive commit messages
- Commit frequently with small, focused changes
- Never commit sensitive data (passwords, API keys, etc.)
- Use `.gitignore` to exclude unnecessary files
