#!/usr/bin/env bash
set -euo pipefail

ZELMA_BIN="${ZELMA_BIN:-/test/zelma}"
SESSION="${ZELMA_E2E_SESSION:-zelma-e2e-$$}"
ROOT="${TMPDIR:-/tmp}/${SESSION}"
REPO_ROOT="${ROOT}/repo"
FAKE_CODEX_BIN="${FAKE_CODEX_BIN:-/test/codex}"

export HOME="${ROOT}/home"
export CODEX_HOME="${ROOT}/codex-home"
export ZELLIJ_CONFIG_DIR="${HOME}/.config/zellij"
export ZELLIJ_CONFIG_FILE="${ZELLIJ_CONFIG_DIR}/config.kdl"
export ZELMA_ZELLIJ_BIN="${ZELMA_ZELLIJ_BIN:-zellij}"
export ZELMA_ZELLIJ_SESSION="${SESSION}"

cleanup() {
  zellij kill-session "${SESSION}" >/dev/null 2>&1 || true
  rm -rf "${ROOT}"
}
trap cleanup EXIT

log() {
  printf '[e2e] %s\n' "$*"
}

run_zelma() {
  (cd "${REPO_ROOT}" && "${ZELMA_BIN}" "$@")
}

wait_for_session() {
  for _ in $(seq 1 40); do
    if zellij list-sessions --short --no-formatting 2>/dev/null | grep -Fxq "${SESSION}"; then
      return 0
    fi
    sleep 0.25
  done
  zellij list-sessions --short --no-formatting >&2 || true
  return 1
}

wait_for_pane_cwd() {
  expected="$1"
  for _ in $(seq 1 40); do
    if zellij --session "${SESSION}" action list-panes --json --all \
      | jq -e --arg expected "${expected}" \
        '.[] | select(.is_plugin == false and .pane_cwd == $expected and .exited == false)' >/dev/null; then
      return 0
    fi
    sleep 0.25
  done
  zellij --session "${SESSION}" action list-panes --json --all >&2 || true
  return 1
}

assert_json() {
  jq -e "$1" >/dev/null
}

log "preparing isolated HOME, CODEX_HOME and repo"
mkdir -p "${HOME}" "${CODEX_HOME}" "${ZELLIJ_CONFIG_DIR}" "${REPO_ROOT}"
cat >"${ZELLIJ_CONFIG_FILE}" <<'EOF'
show_startup_tips false
simplified_ui true
pane_frames false
default_mode "normal"
EOF
git -C "${REPO_ROOT}" init --quiet
test -x "${FAKE_CODEX_BIN}"
export ZELMA_CODEX_BIN="${FAKE_CODEX_BIN}"

log "starting zellij ${SESSION}"
script -qfec "zellij attach --create-background ${SESSION}" /dev/null
wait_for_session

log "running zelma setup"
setup_json="$(run_zelma setup --json)"
printf '%s\n' "${setup_json}" \
  | assert_json '.changed == true and .gitignore_changed == true and .zelma_dir_created == true'

log "running zelma sessions create"
create_stderr="${ROOT}/create.stderr"
if ! create_json="$(run_zelma sessions create --json 2>"${create_stderr}")"; then
  cat "${create_stderr}" >&2
  zellij --session "${SESSION}" action list-panes --json --all >&2 || true
  exit 1
fi
printf '%s\n' "${create_json}" \
  | assert_json '.created == 1 and .registered == 1 and .skipped == 0'
wait_for_pane_cwd "${REPO_ROOT}"

log "checking live list after create"
list_json="$(run_zelma sessions list --live --json)"
printf '%s\n' "${list_json}" \
  | jq -e --arg repo "${REPO_ROOT}" '
      .version == 1
      and (.sessions | length) == 1
      and .sessions[0].opened_path == $repo
      and .sessions[0].live_status == "live"
      and (.sessions[0].state == "candidate" or .sessions[0].state == "active")
    ' >/dev/null

log "creating manually managed Codex pane"
manual_path="${REPO_ROOT}/manual"
mkdir -p "${manual_path}"
manual_pane="$(
  zellij --session "${SESSION}" run \
    --cwd "${manual_path}" \
    --name manual-codex \
    -- "${FAKE_CODEX_BIN}" --cd "${manual_path}"
)"
case "${manual_pane}" in
  terminal_*) ;;
  *) echo "unexpected manual pane id: ${manual_pane}" >&2; exit 1 ;;
esac
wait_for_pane_cwd "${manual_path}"

log "running zelma sessions detect"
detect_json="$(run_zelma sessions detect --json)"
printf '%s\n' "${detect_json}" \
  | jq -e '.added >= 1 and (.active + .candidate) >= 2' >/dev/null

log "checking live list after detect"
detected_list_json="$(run_zelma sessions list --live --json)"
printf '%s\n' "${detected_list_json}" \
  | jq -e --arg manual_path "${manual_path}" '
      [.sessions[] | select(.opened_path == $manual_path and .live_status == "live")]
      | length == 1
    ' >/dev/null

log "docker zellij e2e passed"
