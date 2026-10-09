# Setup on Your Machine (Kali)

## 1. Unzip this wherever you keep projects

```bash
mkdir -p ~/projects
cd ~/projects
unzip racedetect.zip
cd racedetect
```

## 2. Make sure Go is installed

```bash
go version
```

If that fails:

```bash
sudo apt update
sudo apt install golang-go
```

## 3. Fetch the one dependency

This is a Go library we need (`x/tools/go/cfg`). On your machine, with
normal internet, this is a single simple command — the complicated
workaround earlier was only needed because Claude's sandbox has a
restricted network, which your machine doesn't:

```bash
go mod tidy
```

## 4. Run it — this is the moment of truth

```bash
go run ./cmd/detector corpus/
```

You should see the same output we saw while building this: violations
correctly flagged in `missing_unlock/`, nothing flagged in `clean/` or
`defer_ok/`.

Then try the real-world bugs:

```bash
go run ./cmd/detector corpus/real-world/
```

## 5. Turn it into a real GitHub repo

```bash
git init
git add .
git commit -m "racedetect: Stage 0 corpus + Stage 1 parser + Stage 2 CFG reachability check"
```

Then create an empty repo on github.com (don't let GitHub add a
README — we already have files), and:

```bash
git remote add origin https://github.com/<your-username>/racedetect.git
git branch -M main
git push -u origin main
```

## If anything in step 3 or 4 fails

Paste the exact error back to Claude in a new message — don't try to
debug it alone first. That's faster for both of us.
