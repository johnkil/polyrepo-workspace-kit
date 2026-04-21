#!/usr/bin/env sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
minimal_dir="$repo_root/examples/minimal-workspace"

wkit() {
  if [ "${WKIT_BIN:-}" ]; then
    "$WKIT_BIN" "$@"
  else
    (cd "$repo_root" && go run ./cmd/wkit "$@")
  fi
}

tmp_root=$(mktemp -d)
workspace="$tmp_root/workspace"
demo_repos="$tmp_root/repos"

mkdir -p "$demo_repos"
cp -R "$minimal_dir/workspace" "$workspace"
cp -R "$minimal_dir/demo-repos/app-web" "$demo_repos/app-web"
cp -R "$minimal_dir/demo-repos/shared-schema" "$demo_repos/shared-schema"

setup_repo() {
  repo_path=$1
  chmod +x "$repo_path/bin/test"
  git -C "$repo_path" init >/dev/null
  git -C "$repo_path" config user.email test@example.com
  git -C "$repo_path" config user.name "Test User"
  git -C "$repo_path" add .
  git -C "$repo_path" commit -m init >/dev/null
}

setup_repo "$demo_repos/app-web"
setup_repo "$demo_repos/shared-schema"

wkit --workspace "$workspace" bind set app-web "$demo_repos/app-web"
wkit --workspace "$workspace" bind set shared-schema "$demo_repos/shared-schema"
wkit --workspace "$workspace" validate

change_output=$(wkit --workspace "$workspace" change new schema-rollout --title "payload field rollout with drift")
printf '%s\n' "$change_output"
change_id=$(printf '%s\n' "$change_output" | awk '{print $2}')

wkit --workspace "$workspace" scenario pin schema-rollout --change "$change_id"

printf '\nSimulating drift in shared-schema after the scenario was pinned...\n'
printf '\n## payload v4\n\nRemoved legacy payload field.\n' >> "$demo_repos/shared-schema/README.md"
git -C "$demo_repos/shared-schema" add README.md
git -C "$demo_repos/shared-schema" commit -m "simulate schema drift" >/dev/null

printf 'Simulating a local app-web validation failure...\n'
cat > "$demo_repos/app-web/bin/test" <<'SCRIPT'
#!/usr/bin/env sh
set -eu
echo "checking app-web against pinned schema"
echo "app-web contract check failed: payload field customer_id is missing" >&2
echo "hint: regenerate the client after reconciling shared-schema drift" >&2
exit 7
SCRIPT
chmod +x "$demo_repos/app-web/bin/test"

set +e
status_output=$(wkit --workspace "$workspace" scenario status schema-rollout 2>&1)
status_code=$?
set -e
printf '%s\n' "$status_output"
if [ "$status_code" -ne 4 ]; then
  printf 'expected scenario status to exit 4 for drift, got %s\n' "$status_code" >&2
  exit 1
fi

set +e
run_output=$(wkit --workspace "$workspace" scenario run schema-rollout 2>&1)
run_code=$?
set -e
printf '%s\n' "$run_output"
if [ "$run_code" -ne 5 ]; then
  printf 'expected scenario run to exit 5 for command failure, got %s\n' "$run_code" >&2
  exit 1
fi

wkit --workspace "$workspace" validate

printf '\nFailure demo produced the expected drift and command failure evidence.\n'
printf 'Demo workspace: %s\n' "$workspace"
printf 'Demo repos: %s\n' "$demo_repos"
