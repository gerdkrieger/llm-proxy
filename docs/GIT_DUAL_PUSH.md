# 🔄 Git Dual-Push Setup

## Overview

This repository is configured to automatically push to **both GitLab and GitHub** with a single `git push` command.

---

## Configuration

### Remote Setup

```bash
origin (fetch) → git@gitlab.com:krieger-engineering/llm-proxy.git
origin (push)  → git@gitlab.com:krieger-engineering/llm-proxy.git  ✅ GitLab
origin (push)  → https://github.com/gerdkrieger/llm-proxy.git     ✅ GitHub
```

### How it Works

When you run `git push`, Git will:
1. Push to GitLab (primary)
2. Push to GitHub (backup/mirror)

Both pushes happen automatically!

---

## Usage

### Normal Workflow

```bash
# Make changes
git add .
git commit -m "Feature: Add new functionality"

# Push to BOTH GitLab and GitHub
git push

# ✅ Automatically pushed to both platforms!
```

### Specific Branch

```bash
git push origin master           # Pushes master to both
git push origin feature-branch   # Pushes feature-branch to both
```

### Force Push (CAREFUL!)

```bash
git push --force   # Force pushes to BOTH (use with caution!)
```

---

## Verify Configuration

```bash
# Check remote configuration
git remote -v

# Should show:
# origin  git@gitlab.com:krieger-engineering/llm-proxy.git (fetch)
# origin  git@gitlab.com:krieger-engineering/llm-proxy.git (push)
# origin  https://github.com/gerdkrieger/llm-proxy.git (push)
```

---

## Setup for New Repository

If you want to set up dual-push for another repository:

```bash
cd /path/to/your/repo
/path/to/llm-proxy/scripts/setup-dual-push.sh
```

Or manually:

```bash
# Set GitLab as fetch URL
git remote set-url origin git@gitlab.com:user/repo.git

# Add both push URLs
git remote set-url --add --push origin git@gitlab.com:user/repo.git
git remote set-url --add --push origin https://github.com/user/repo.git

# Verify
git remote -v
```

---

## Benefits

✅ **One command** pushes to both platforms  
✅ **GitLab as primary** (fetch from there)  
✅ **GitHub as mirror/backup**  
✅ **No manual synchronization** needed  
✅ **Consistent state** across both platforms

---

## Important Notes

### Landing Page Repository

⚠️ **IMPORTANT:** The landing page is **ONLY** in the GitLab repository!

- LLM-Proxy code → GitLab + GitHub ✅
- Landing page code → GitLab ONLY ✅

### Authentication

- **GitLab:** Uses SSH key (`~/.ssh/gkrieger`)
- **GitHub:** Uses HTTPS (may prompt for credentials on first push)

If GitHub asks for credentials:
```bash
# Use Personal Access Token (PAT) as password
Username: gerdkrieger
Password: <your-github-personal-access-token>
```

---

## Troubleshooting

### Push fails to GitHub

```bash
# Check if you're authenticated
git config --global credential.helper store

# Try pushing manually to test
git push https://github.com/gerdkrieger/llm-proxy.git master
```

### Push fails to GitLab

```bash
# Check SSH key
ssh -T git@gitlab.com

# Should show: "Welcome to GitLab, @krieger-engineering!"
```

### Remove dual-push (if needed)

```bash
# Reset to single remote (GitLab only)
git remote set-url origin git@gitlab.com:krieger-engineering/llm-proxy.git

# Verify
git remote -v
```

### Add it back

```bash
git remote set-url --add --push origin git@gitlab.com:krieger-engineering/llm-proxy.git
git remote set-url --add --push origin https://github.com/gerdkrieger/llm-proxy.git
```

---

## FAQ

### Q: What happens if one push fails?

Git will push to the first remote (GitLab), then to the second (GitHub). If the second fails, you'll see an error but the first push succeeded.

### Q: Can I push to only one remote?

Yes:
```bash
git push git@gitlab.com:krieger-engineering/llm-proxy.git master  # Only GitLab
git push https://github.com/gerdkrieger/llm-proxy.git master      # Only GitHub
```

### Q: How do I know both pushes succeeded?

Check the output:
```bash
$ git push
To gitlab.com:krieger-engineering/llm-proxy.git
   abc1234..def5678  master -> master    ✅
To https://github.com/gerdkrieger/llm-proxy.git
   abc1234..def5678  master -> master    ✅
```

### Q: Does this work with branches?

Yes! All branches and tags are pushed to both remotes.

---

## Related Documentation

- **Deployment Strategy:** [docs/deployment/REGISTRY_DEPLOYMENT.md](deployment/REGISTRY_DEPLOYMENT.md)
- **Setup Script:** [scripts/setup-dual-push.sh](../scripts/setup-dual-push.sh)

---

**Last Updated:** 2026-03-20  
**Version:** 1.0
