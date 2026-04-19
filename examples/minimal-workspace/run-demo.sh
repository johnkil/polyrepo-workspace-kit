#!/usr/bin/env sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)

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
cp -R "$script_dir/workspace" "$workspace"
cp -R "$script_dir/demo-repos/app-web" "$demo_repos/app-web"
cp -R "$script_dir/demo-repos/shared-schema" "$demo_repos/shared-schema"

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

change_output=$(wkit --workspace "$workspace" change new schema-rollout --title "payload field rollout")
printf '%s\n' "$change_output"
change_id=$(printf '%s\n' "$change_output" | awk '{print $2}')

wkit --workspace "$workspace" scenario pin schema-rollout --change "$change_id"
wkit --workspace "$workspace" scenario run schema-rollout

wkit --workspace "$workspace" install plan portable app-web
wkit --workspace "$workspace" install apply portable app-web --yes
wkit --workspace "$workspace" validate

printf '\nDemo workspace: %s\n' "$workspace"
printf 'Demo repos: %s\n' "$demo_repos"
