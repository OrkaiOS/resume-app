#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "${ROOT}" ]]; then echo "error: must run inside a git repository" >&2; exit 1; fi
cd "${ROOT}"
# CSVs written to .data/project-status/ (not gitignored, visible to orkai index).
OUT_DIR="${ROOT}/.data/project-status"
mkdir -p "${OUT_DIR}"
OUT_COMMITS="${OUT_DIR}/commits.csv"
OUT_FILES="${OUT_DIR}/commit_files.csv"
python3 - "${OUT_COMMITS}" "${OUT_FILES}" << 'PY'
from __future__ import annotations
import csv, re, subprocess, sys
from datetime import datetime, timezone
commits_path, files_path = sys.argv[1], sys.argv[2]
CONVENTIONAL = re.compile(r"^(?P<type>feat|fix|docs|test|chore|refactor|perf|ci|build|style|revert)(?:\((?P<scope>[^)]+)\))?(?P<breaking>!)?:")
COMMIT_FIELDS = ["commit_sha","committed_at","author_name","author_email","subject","commit_type","scope","additions","deletions","files_changed","net_lines","is_merge"]
FILE_FIELDS = ["commit_sha","committed_at","path","area","extension","additions","deletions","change_kind"]
def parse_type_scope(subject):
    m = CONVENTIONAL.match(subject)
    if not m: return "other", ""
    return m.group("type"), m.group("scope") or ""
def area_and_ext(path):
    path = path.strip()
    if not path: return "", ""
    parts = path.split("/")
    area = parts[0] if len(parts) > 1 else "."
    base = parts[-1]
    if "." in base and not base.startswith("."): ext = "." + base.rsplit(".", 1)[-1]
    elif base.startswith(".") and base.count(".") >= 1: ext = base
    else: ext = ""
    return area, ext
def change_kind(add_s, del_s):
    if add_s == "-" or del_s == "-": return "binary"
    add_n, del_n = int(add_s), int(del_s)
    if add_n == 0 and del_n > 0: return "deleted"
    if add_n > 0 and del_n == 0: return "added"
    return "modified"
def is_merge_commit(parents, subject):
    if not parents.strip(): return False
    if len(parents.split()) > 1: return True
    return subject.lower().startswith("merge ")
raw = subprocess.check_output(["git","log","--reverse","--numstat","--format=COMMIT\x1f%H\x1f%aI\x1f%an\x1f%ae\x1f%s\x1f%P"],text=True,cwd=None)
commits = []
file_rows = []
cur = None
cur_add = cur_del = cur_files = 0
for line in raw.splitlines():
    if line.startswith("COMMIT"):
        if cur is not None:
            ctype, scope = parse_type_scope(cur["subject"])
            cur.update({"commit_type":ctype,"scope":scope,"additions":cur_add,"deletions":cur_del,"files_changed":cur_files,"net_lines":cur_add-cur_del,"is_merge":is_merge_commit(cur.pop("_parents"),cur["subject"])})
            commits.append(cur)
        parts = line.split("\x1f")
        cur = {"commit_sha":parts[1],"committed_at":parts[2],"author_name":parts[3],"author_email":parts[4],"subject":parts[5],"_parents":parts[6] if len(parts)>6 else ""}
        cur_add = cur_del = cur_files = 0
        continue
    if not line.strip() or cur is None: continue
    cols = line.split("\t", 2)
    if len(cols) < 3: continue
    add_s, del_s, path = cols[0], cols[1], cols[2]
    try: add_n = 0 if add_s == "-" else int(add_s); del_n = 0 if del_s == "-" else int(del_s)
    except ValueError: add_n = del_n = 0
    cur_add += add_n; cur_del += del_n; cur_files += 1
    area, ext = area_and_ext(path)
    file_rows.append({"commit_sha":cur["commit_sha"],"committed_at":cur["committed_at"],"path":path,"area":area,"extension":ext,"additions":add_n,"deletions":del_n,"change_kind":change_kind(add_s,del_s)})
if cur is not None:
    ctype, scope = parse_type_scope(cur["subject"])
    cur.update({"commit_type":ctype,"scope":scope,"additions":cur_add,"deletions":cur_del,"files_changed":cur_files,"net_lines":cur_add-cur_del,"is_merge":is_merge_commit(cur.pop("_parents"),cur["subject"])})
    commits.append(cur)
generated_at = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
# CSV has no native boolean type; orkai infers bool columns from lowercase 'true'/'false' strings.
# Write explicit lowercase strings so the is_merge column is always typed as bool by the indexer.
with open(commits_path,"w",newline="",encoding="utf-8") as f:
    w = csv.DictWriter(f,fieldnames=COMMIT_FIELDS,lineterminator="\n")
    w.writeheader()
    for r in commits:
        r["is_merge"] = "true" if r["is_merge"] else "false"
        w.writerow(r)
with open(files_path,"w",newline="",encoding="utf-8") as f:
    w = csv.DictWriter(f,fieldnames=FILE_FIELDS,lineterminator="\n")
    w.writeheader(); w.writerows(file_rows)
print(f"generated_at={generated_at}")
print(f"commits={len(commits)} -> {commits_path}")
print(f"commit_files={len(file_rows)} -> {files_path}")
if commits: print(f"date_range={commits[0]['committed_at'][:10]} .. {commits[-1]['committed_at'][:10]}")
PY
echo "out_dir=${OUT_DIR}"
echo "Done. From the repo root run: orkai index analytics ."