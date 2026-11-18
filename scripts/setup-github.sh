#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# scripts/setup-github.sh — one-shot, idempotent GitHub project setup for
# ColdCase's agile workflow: labels, milestones, a user-story backlog,
# a Kanban project board and branch protection on main.
#
# Requirements: gh CLI authenticated (gh auth status), push access to $REPO.
# Safe to re-run: existing labels/milestones/issues/board columns are reused.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

REPO="${GITHUB_REPO:-Fakechippies/ColdCase}"
OWNER="${GITHUB_OWNER:-Fakechippies}"
TODAY="$(date +%F)"
DUE1="$(date -d '+14 days' +%F 2>/dev/null || date -v+14d +%F)"
DUE2="$(date -d '+28 days' +%F 2>/dev/null || date -v+28d +%F)"
DUE3="$(date -d '+42 days' +%F 2>/dev/null || date -v+42d +%F)"

echo "==> Configuring $REPO (milestones due $DUE1 / $DUE2 / $DUE3)"

# ─── labels ──────────────────────────────────────────────────────────────────
label() { # name description color
  if gh label view "$1" --repo "$REPO" >/dev/null 2>&1; then
    gh label edit "$1" --repo "$REPO" --description "$2" --color "$3" >/dev/null
  else
    gh label create "$1" --repo "$REPO" --description "$2" --color "$3" >/dev/null
  fi
}

label "user-story" "A feature described as role/goal/benefit" "0e8a16"
label "bug"        "Something is broken"                      "d73a4a"
label "tech-debt"  "Internal quality or performance work"     "fbca04"
label "security"   "Security-relevant work"                   "ee0701"
label "ci-cd"      "Pipelines, releases, tooling"             "1d76db"
label "docs"       "Documentation"                            "5319e7"
label "sprint-1"   "Sprint 1 (current cycle)"                 "c2e0c6"
label "sprint-2"   "Sprint 2 (next cycle)"                    "bfd4f2"
label "sprint-3"   "Sprint 3 (future)"                        "d4c5f9"
echo "  labels ok"

# ─── milestones ──────────────────────────────────────────────────────────────
milestone() { # title due_date description
  local n
  n="$(gh api "repos/$REPO/milestones" --jq ".[] | select(.title==\"$1\") | .number" 2>/dev/null || true)"
  if [ -n "$n" ]; then
    echo "  milestone '$1' exists (#$n)"
  else
    gh api "repos/$REPO/milestones" -f title="$1" -f due_on="${2}T12:00:00Z" \
       -f description="$3" --jq '.number' | xargs -I{} echo "  milestone '$1' -> #{}"
  fi
}

milestone "Sprint 1" "$DUE1" "First sprint: CI/CD foundation, security scanning, signed releases, console lifecycle"
milestone "Sprint 2" "$DUE2" "Second sprint: session hardening, catalog metadata, performance"
milestone "Sprint 3" "$DUE3" "Third sprint: container hardening, evidence workflow, release UX"

# ─── issues ──────────────────────────────────────────────────────────────────
create_issue() { # title label(s) milestone body
  local num
  num="$(gh issue list --repo "$REPO" --search "$1 in:title" --limit 1 --json number --jq '.[0].number // empty' 2>/dev/null || true)"
  if [ -n "$num" ]; then
    echo "  issue exists: #$num $1"
    echo "$num"
  else
    num="$(gh issue create --repo "$REPO" --title "$1" --label "$2" --milestone "$3" --body "$4" | grep -oE '[0-9]+$')"
    echo "  created: #$num $1"
    echo "$num"
  fi
}

story() { # role goal benefit
  printf '**As a %s, I want %s so that %s.**\n' "$1" "$2" "$3"
}

ISSUES_DONE=""

n=$(create_issue "Set up CI: lint, test and build matrix for every push" \
  "ci-cd,sprint-1" "Sprint 1" "$(story 'developer' 'every push and PR vetted, linted, tested and cross-built' 'breakage is caught before merge')"

  printf '%s\n\n**Acceptance criteria**\n- [x] CI runs on push to main and on all PRs\n- [x] Jobs: go vet, golangci-lint, go test -race -cover, go build\n- [x] Build matrix: ubuntu, macOS, Windows; Go version taken from go.mod\n- [x] Go module cache; coverage uploaded as an artifact\n- [x] Minimal permissions (contents: read)\n- [x] Unit tests exist for session crypto and catalog\n\n**Delivered by** #PR (feat/ci-cd-agile)\n')
  ISSUES_DONE="$ISSUES_DONE $n"

n=$(create_issue "Security scanning: govulncheck, CodeQL and Dependabot" \
  "security,ci-cd,sprint-1" "Sprint 1" "$(story 'maintainer' 'automated vulnerability and static security analysis' 'known CVEs and risky patterns are caught without manual effort')"

  printf '%s\n\n**Acceptance criteria**\n- [x] govulncheck ./... on Go-touching PRs and a weekly cron\n- [x] CodeQL for Go on PRs and a weekly cron, least-privilege token\n- [x] Dependabot for gomod, github-actions and docker, weekly\n\n**Delivered by** #PR (feat/ci-cd-agile)\n')
  ISSUES_DONE="$ISSUES_DONE $n"

n=$(create_issue "Signed multi-platform releases with Ed25519" \
  "security,ci-cd,sprint-1" "Sprint 1" "$(story 'investigator' 'verifiable release artifacts' 'I can prove the binary I download is the binary CI built')"

  printf '%s\n\n**Acceptance criteria**\n- [x] GoReleaser builds linux/darwin/windows on amd64+arm64\n- [x] checksums.txt signed with Ed25519 in the repo'"'"'s own signature format\n- [x] Private key only in the COLDCASE_SIGNING_KEY Actions secret\n- [x] Public key committed (keys/release_public.key) + verification docs\n- [x] Verified end-to-end locally: goreleaser snapshot + sign + verify + tamper-rejection\n\n**Delivered by** #PR (feat/ci-cd-agile)\n')
  ISSUES_DONE="$ISSUES_DONE $n"

n=$(create_issue "Detached console lifecycle: serve up/down/status/run" \
  "user-story,sprint-1" "Sprint 1" "$(story 'investigator' 'start and stop the web console without holding a terminal' 'the console runs like a service, not like a foreground process')"

  printf '%s\n\n**Acceptance criteria**\n- [x] coldcase serve daemonizes and returns immediately\n- [x] serve up / down / status / run subcommands\n- [x] Graceful SIGTERM shutdown (verified ~250ms), SIGKILL fallback\n- [x] down refuses to kill non-ColdCase pids; stale pid files cleaned\n\n**Delivered by** commit 007ebc1 on main\n')
  ISSUES_DONE="$ISSUES_DONE $n"

n=$(create_issue "Web console overhaul: dashboard, catalog, upload runner" \
  "user-story,sprint-1" "Sprint 1" "$(story 'investigator' 'a usable, readable console' 'dashboard telemetry, tool discovery and evidence upload are all first-class')"

  printf '%s\n\n**Acceptance criteria**\n- [x] Read-only dashboard with charts (availability, states, activity, usage)\n- [x] Run Tool tab with file upload ({file} arg injection) and live output\n- [x] Tool catalog with icons, chips and availability badges\n- [x] Availability page with donut, coverage bars, filterable table\n- [x] Black/orange design system documented in design-system/coldcase-console\n\n**Delivered by** commit c54eeb7 on main\n')
  ISSUES_DONE="$ISSUES_DONE $n"

create_issue "Encrypt session output and evidence at rest" \
  "user-story,sprint-2" "Sprint 2" "$(story 'investigator' 'session output encrypted at rest' 'sealed cases cannot leak confidential output even if the disk is seized')

**Acceptance criteria**
- [ ] Logger.SaveOutput encrypts when the session is marked encrypted (currently a stub)
- [ ] AES-256-GCM with a key derived from the session passphrase (PBKDF2 exists)
- [ ] Verify/decrypt available in the CLI and web console
- [ ] No plaintext copies left behind (temp files, output files)"

create_issue "Per-tool metadata in the catalog" \
  "user-story,tech-debt,sprint-2" "Sprint 2" "$(story 'tool developer' 'explicit per-tool metadata (icon, docs link, examples)' 'the UI stops relying on filename heuristics and stays exact as the catalog grows')

**Acceptance criteria**
- [ ] Tool struct gains optional icon/docs/examples fields
- [ ] Web UI uses the metadata; heuristic icon matching removed
- [ ] TestRegistryIntegrity extended to cover the new fields"

create_issue "Record MD5 alongside SHA256 for evidence files" \
  "user-story,tech-debt,sprint-2" "Sprint 2" "$(story 'investigator' 'legacy MD5 digests recorded with evidence' 'older case files and tooling remain cross-checkable')

**Acceptance criteria**
- [ ] session.HashFile computes both SHA256 and MD5 (MD5 is currently a stub)
- [ ] Evidence records, CLI and web UI show both digests
- [ ] gosec justification noted where MD5 import is required (legacy interop)"

create_issue "Speed up console page loads (parallel availability checks)" \
  "tech-debt,sprint-2" "Sprint 2" "$(story 'console user' 'fast dashboard and catalog loads' 'availability probes do not serialize page rendering')

**Acceptance criteria**
- [ ] Binary availability probes run concurrently with a bounded worker pool
- [ ] Dashboard/tools/status pages show no measurable regression under 111 tools
- [ ] Probe results cached per request cycle"

create_issue "Release page shows verification instructions and public key" \
  "ci-cd,sprint-2" "Sprint 2" "$(story 'new user' 'the release page itself explains how to verify' 'verifying a release takes one minute without reading the repo')

**Acceptance criteria**
- [ ] Release workflow appends verify instructions + public key to the release body
- [ ] checksums.txt.sig attached to every release when the secret is set"

create_issue "Harden container image (non-root user, pinned digests)" \
  "security,sprint-3" "Sprint 3" "$(story 'defender' 'the forensics image runs least-privilege' 'a compromise inside the container cannot escalate to root on the host mounts')

**Acceptance criteria**
- [ ] Dockerfile gains a non-root USER for the runtime stage
- [ ] Base images pinned by digest where practical
- [ ] Tools verified to still run (path remapping unaffected)"

create_issue "Evidence upload de-duplication" \
  "user-story,sprint-3" "Sprint 3" "$(story 'investigator' 'identical uploads not stored twice' 'evidence storage stays small during long investigations')

**Acceptance criteria**
- [ ] SHA256 computed while streaming the upload
- [ ] Identical files reuse the stored copy and link the session to it"

create_issue "Tag-build reproducibility check in release CI" \
  "ci-cd,sprint-3" "Sprint 3" "$(story 'auditor' 'reproducible release builds' 'the checksums I verify are trustworthy beyond CI hygiene')

**Acceptance criteria**
- [ ] Release workflow re-builds and compares at least the linux-amd64 checksum
- [ ] Mismatch fails the release"

create_issue "Session export report templates" \
  "user-story,sprint-3" "Sprint 3" "$(story 'auditor' 'professional HTML/Markdown export templates' 'reports can be attached to case files without reformatting')

**Acceptance criteria**
- [ ] Export templates aligned with the console design system
- [ ] Include evidence table, signature status and verification banner"

echo "  issues ok"

# ─── project board ───────────────────────────────────────────────────────────
OWNER_ID="$(gh api graphql -f query='{viewer{id}}' --jq '.data.viewer.id')"

EXISTING="$(gh project list --owner "$OWNER_ID" --format json 2>/dev/null | python3 -c 'import json,sys; d=json.load(sys.stdin); print(next((p["id"] for p in d if p["title"]=="ColdCase Delivery Board"), ""))' || true)"
if [ -z "$EXISTING" ]; then
  PROJECT_ID="$(gh project create --owner "$OWNER_ID" --title "ColdCase Delivery Board" --format json --jq '.id')"
else
  PROJECT_ID="$EXISTING"
fi
echo "  board: $PROJECT_ID"

# Customize the Status single-select field to Kanban columns.
FIELD_ID="$(gh api graphql -f query='query($p:ID!){node(id:$p){... on ProjectV2{field(name:"Status"){... on ProjectV2SingleSelectField{id options{id name}}}}}}' -f p="$PROJECT_ID" --jq '.data.node.field.id')"
OPTS_JSON="$(gh api graphql -f query='query($p:ID!){node(id:$p){... on ProjectV2{field(name:"Status"){... on ProjectV2SingleSelectField{options{id name}}}}}}' -f p="$PROJECT_ID" --jq '.data.node.field.options')"
OPT_TODO="$(echo "$OPTS_JSON" | python3 -c 'import json,sys; print(next((o["id"] for o in json.load(sys.stdin) if o["name"]=="Todo"), ""))')"
OPT_INPROG="$(echo "$OPTS_JSON" | python3 -c 'import json,sys; print(next((o["id"] for o in json.load(sys.stdin) if o["name"]=="In Progress"), ""))')"
OPT_DONE="$(echo "$OPTS_JSON" | python3 -c 'import json,sys; print(next((o["id"] for o in json.load(sys.stdin) if o["name"]=="Done"), ""))')"

OPTIONS="["
[ -n "$OPT_TODO" ]    && OPTIONS+='{"id":"'"$OPT_TODO"'","name":"Backlog"},'
OPTIONS+='{\"name\":\"Sprint Backlog\"},'
[ -n "$OPT_INPROG" ]  && OPTIONS+='{"id":"'"$OPT_INPROG"'","name":"In Progress"},'
OPTIONS+='{\"name\":\"In Review\"},'
[ -n "$OPT_DONE" ]    && OPTIONS+='{"id":"'"$OPT_DONE"'","name":"Done"}'
OPTIONS="${OPTIONS%,}]"

gh api graphql -f query='mutation($field:ID!,$options:[ProjectV2SingleSelectFieldOptionInput!]!){updateProjectV2SingleSelectField(input:{fieldId:$field,options:$options}){projectV2SingleSelectField{id}}}' \
  -f field="$FIELD_ID" -f options="$OPTIONS" >/dev/null
echo "  status columns ok"

opt_for() { # column name -> option id (query fresh after update)
  gh api graphql -f query='query($p:ID!){node(id:$p){... on ProjectV2{field(name:"Status"){... on ProjectV2SingleSelectField{options{id name}}}}}}' \
    -f p="$PROJECT_ID" --jq '.data.node.field.options[] | select(.name=="'"$1"'") | .id'
}

add_to_board() { # issue_number column
  local node_id item_id
  node_id="$(gh issue view "$1" --repo "$REPO" --json id --jq '.id')"
  item_id="$(gh api graphql -f query='mutation($p:ID!,$c:ID!){addProjectV2ItemById(input:{projectId:$p,contentId:$c}){item{id}}}' \
    -f p="$PROJECT_ID" -f c="$node_id" --jq '.data.addProjectV2ItemById.item.id')"
  gh api graphql -f query='mutation($p:ID!,$i:ID!,$f:ID!,$o:String!){updateProjectV2ItemFieldValue(input:{projectId:$p,itemId:$i,fieldId:$f,value:{singleSelectOptionId:$o}}){projectV2Item{id}}}' \
    -f p="$PROJECT_ID" -f i="$item_id" -f f="$FIELD_ID" \
    -f o="$(opt_for "$2")" >/dev/null
}

for num in $ISSUES_DONE; do add_to_board "$num" "Done"; done
# Sprint 2 / 3 issues were created above; map by searching labels.
for num in $(gh issue list --repo "$REPO" --label sprint-2 --json number --jq '.[].number'); do add_to_board "$num" "Sprint Backlog"; done
for num in $(gh issue list --repo "$REPO" --label sprint-3 --json number --jq '.[].number'); do add_to_board "$num" "Backlog"; done
echo "  board columns populated"

# ─── branch protection ───────────────────────────────────────────────────────
if gh api "repos/$REPO/branches/main/protection" >/dev/null 2>&1; then
  echo "  branch protection already configured"
else
  gh api -X PUT "repos/$REPO/branches/main/protection" --input - <<'JSON' >/dev/null
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["go vet", "golangci-lint", "test (ubuntu-latest)", "build (ubuntu-latest)"]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": null,
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false
}
JSON
  echo "  branch protection on main: PRs + required checks + no force pushes"
fi

echo "==> Done. Board: https://github.com/users/$OWNER/projects (ColdCase Delivery Board)"
