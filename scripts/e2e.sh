#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

BOLD=$'\033[1m'
ACCENT=$'\033[38;2;251;191;36m'
MUTED=$'\033[38;2;90;100;128m'
SUCCESS=$'\033[38;2;0;229;204m'
ERROR=$'\033[38;2;230;57;70m'
NC=$'\033[0m'

# Parse arguments after suite
E2E_FILTER=""
E2E_EXTRA=""
for arg in "${@:2}"; do
  case "$arg" in
    filter=*)
      E2E_FILTER="${arg#filter=}"
      ;;
    extra=*)
      E2E_EXTRA="${arg#extra=}"
      ;;
    *)
      # Backwards compatibility: treat bare argument as filter
      if [ -z "$E2E_FILTER" ]; then
        E2E_FILTER="$arg"
      fi
      ;;
  esac
done


show_filter_status() {
  if [ -n "${E2E_FILTER}" ]; then
    echo "  ${MUTED}filter: ${E2E_FILTER}${NC}"
    return
  fi

  echo "  ${MUTED}filter: none (running all scenarios in this suite)${NC}"
}

# Detect available docker compose command
COMPOSE="docker compose"
if [ -n "${PINCHTAB_COMPOSE:-}" ]; then
  COMPOSE="${PINCHTAB_COMPOSE}"
elif docker compose version >/dev/null 2>&1; then
  COMPOSE="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE="docker-compose"
else
  echo "Neither 'docker compose' nor 'docker-compose' is available" >&2
  exit 127
fi

compose_down() {
  local compose_file="$1"
  $COMPOSE -f "${compose_file}" down -v 2>/dev/null || true
}

# Always rebuild the fixtures nginx image up front so changes to
# tests/e2e/nginx/{Dockerfile,default.conf} are guaranteed to land in the
# container, regardless of whether `compose run --build` would otherwise
# rebuild this transitive dependency. This eliminates an entire class of
# "the new fixture config didn't get picked up" failure modes.
build_support_images() {
  local compose_file="$1"
  $COMPOSE -f "${compose_file}" build fixtures
}

dump_compose_failure() {
  local compose_file="$1"
  shift
  local log_prefix="$1"
  shift
  local services=("$@")

  mkdir -p tests/e2e/results
  for service in "${services[@]}"; do
    $COMPOSE -f "${compose_file}" logs "${service}" > "tests/e2e/results/${log_prefix}-${service}.log" 2>&1 || true
  done
}

show_suite_artifacts() {
  local summary_file="$1"
  local report_file="$2"
  local progress_file="$3"
  local log_prefix="$4"
  shift 4
  local services=("$@")
  local printed=0

  if [ -f "${summary_file}" ]; then
    echo ""
    echo "  ${MUTED}Summary saved to: ${summary_file}${NC}"
    printed=1
  fi

  if [ -f "${report_file}" ]; then
    echo "  ${MUTED}Report saved to: ${report_file}${NC}"
    printed=1
  fi

  if [ -f "${progress_file}" ]; then
    echo "  ${MUTED}Progress saved to: ${progress_file}${NC}"
    printed=1
  fi

  for service in "${services[@]}"; do
    local service_log="tests/e2e/results/${log_prefix}-${service}.log"
    if [ -f "${service_log}" ]; then
      echo "  ${MUTED}Logs saved to: ${service_log}${NC}"
      printed=1
    fi
  done

  if [ "${printed}" -eq 1 ]; then
    echo ""
  fi
}

show_suite_summary() {
  local compose_file="$1"
  shift
  :
}

prepare_suite_results() {
  local summary_file="$1"
  local report_file="$2"
  local progress_file="$3"
  local log_prefix="$4"

  rm -f \
    "${summary_file}" \
    "${report_file}" \
    "${progress_file}" \
    tests/e2e/results/${log_prefix}-*.log \
    tests/e2e/results/summary.txt \
    tests/e2e/results/report.md
}

run_api() {
  local compose_file="tests/e2e/docker-compose.yml"
  local summary_file="tests/e2e/results/summary-api.txt"
  local report_file="tests/e2e/results/report-api.md"
  local progress_file="tests/e2e/results/progress-api.log"
  local log_prefix="logs-api"
  echo "  ${ACCENT}${BOLD}E2E API tests (Docker)${NC}"
  show_filter_status
  echo ""
  prepare_suite_results "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}"
  build_support_images "${compose_file}"
  set +e
  local args=""
  [ -n "${E2E_FILTER}" ] && args="${args} filter=${E2E_FILTER}"
  [ -n "${E2E_EXTRA}" ] && args="${args} extra=${E2E_EXTRA}"
  $COMPOSE -f "${compose_file}" run --build --rm runner-api /bin/bash /e2e/run.sh api ${args}
  local api_exit=$?
  set -e
  if [ "${api_exit}" -ne 0 ]; then
    dump_compose_failure "${compose_file}" "${log_prefix}" runner-api pinchtab
    show_suite_artifacts "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}" runner-api pinchtab
  fi
  compose_down "${compose_file}"
  return "${api_exit}"
}

run_api_extended() {
  local compose_file="tests/e2e/docker-compose-multi.yml"
  local summary_file="tests/e2e/results/summary-api-extended.txt"
  local report_file="tests/e2e/results/report-api-extended.md"
  local progress_file="tests/e2e/results/progress-api-extended.log"
  local log_prefix="logs-api-extended"
  echo "  ${ACCENT}${BOLD}E2E API Extended tests (Docker)${NC}"
  show_filter_status
  echo ""
  prepare_suite_results "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}"
  build_support_images "${compose_file}"
  set +e
  E2E_SUITE=api E2E_EXTENDED=true E2E_SCENARIO_FILTER="${E2E_FILTER}" $COMPOSE -f "${compose_file}" up --build --abort-on-container-exit --exit-code-from runner-api runner-api
  local api_exit=$?
  set -e
  if [ "${api_exit}" -ne 0 ]; then
    dump_compose_failure "${compose_file}" "${log_prefix}" runner-api pinchtab pinchtab-secure pinchtab-medium pinchtab-full pinchtab-lite pinchtab-bridge
    show_suite_artifacts "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}" runner-api pinchtab pinchtab-secure pinchtab-medium pinchtab-full pinchtab-lite pinchtab-bridge
  fi
  compose_down "${compose_file}"
  return "${api_exit}"
}

run_cli() {
  local compose_file="tests/e2e/docker-compose.yml"
  local summary_file="tests/e2e/results/summary-cli.txt"
  local report_file="tests/e2e/results/report-cli.md"
  local progress_file="tests/e2e/results/progress-cli.log"
  local log_prefix="logs-cli"
  echo "  ${ACCENT}${BOLD}E2E CLI tests (Docker)${NC}"
  show_filter_status
  echo ""
  prepare_suite_results "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}"
  build_support_images "${compose_file}"
  set +e
  local args=""
  [ -n "${E2E_FILTER}" ] && args="${args} filter=${E2E_FILTER}"
  [ -n "${E2E_EXTRA}" ] && args="${args} extra=${E2E_EXTRA}"
  $COMPOSE -f "${compose_file}" run --build --rm runner-cli /bin/bash /e2e/run.sh cli ${args}
  local cli_exit=$?
  set -e
  if [ "${cli_exit}" -ne 0 ]; then
    dump_compose_failure "${compose_file}" "${log_prefix}" runner-cli pinchtab
    show_suite_artifacts "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}" runner-cli pinchtab
  fi
  compose_down "${compose_file}"
  return "${cli_exit}"
}

run_cli_extended() {
  local compose_file="tests/e2e/docker-compose.yml"
  local summary_file="tests/e2e/results/summary-cli-extended.txt"
  local report_file="tests/e2e/results/report-cli-extended.md"
  local progress_file="tests/e2e/results/progress-cli-extended.log"
  local log_prefix="logs-cli-extended"
  echo "  ${ACCENT}${BOLD}E2E CLI Extended tests (Docker)${NC}"
  show_filter_status
  echo ""
  prepare_suite_results "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}"
  build_support_images "${compose_file}"
  set +e
  E2E_SCENARIO_FILTER="${E2E_FILTER}" $COMPOSE -f "${compose_file}" up --build --abort-on-container-exit --exit-code-from runner-cli runner-cli
  local cli_exit=$?
  set -e
  if [ "${cli_exit}" -ne 0 ]; then
    dump_compose_failure "${compose_file}" "${log_prefix}" runner-cli pinchtab
    show_suite_artifacts "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}" runner-cli pinchtab
  fi
  compose_down "${compose_file}"
  return "${cli_exit}"
}

run_infra() {
  local compose_file="tests/e2e/docker-compose.yml"
  local summary_file="tests/e2e/results/summary-infra.txt"
  local report_file="tests/e2e/results/report-infra.md"
  local progress_file="tests/e2e/results/progress-infra.log"
  local log_prefix="logs-infra"
  echo "  ${ACCENT}${BOLD}E2E Infra tests (Docker)${NC}"
  show_filter_status
  echo ""
  prepare_suite_results "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}"
  build_support_images "${compose_file}"
  set +e
  local args=""
  [ -n "${E2E_FILTER}" ] && args="${args} filter=${E2E_FILTER}"
  [ -n "${E2E_EXTRA}" ] && args="${args} extra=${E2E_EXTRA}"
  $COMPOSE -f "${compose_file}" run --build --rm runner-api /bin/bash /e2e/run.sh infra ${args}
  local infra_exit=$?
  set -e
  if [ "${infra_exit}" -ne 0 ]; then
    dump_compose_failure "${compose_file}" "${log_prefix}" runner-api pinchtab
    show_suite_artifacts "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}" runner-api pinchtab
  fi
  compose_down "${compose_file}"
  return "${infra_exit}"
}

run_infra_extended() {
  local compose_file="tests/e2e/docker-compose-multi.yml"
  local summary_file="tests/e2e/results/summary-infra-extended.txt"
  local report_file="tests/e2e/results/report-infra-extended.md"
  local progress_file="tests/e2e/results/progress-infra-extended.log"
  local log_prefix="logs-infra-extended"
  echo "  ${ACCENT}${BOLD}E2E Infra Extended tests (Docker)${NC}"
  show_filter_status
  echo ""
  prepare_suite_results "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}"
  build_support_images "${compose_file}"
  set +e
  E2E_SUITE=infra E2E_EXTENDED=true E2E_SCENARIO_FILTER="${E2E_FILTER}" $COMPOSE -f "${compose_file}" up --build --abort-on-container-exit --exit-code-from runner-api runner-api
  local infra_exit=$?
  set -e
  if [ "${infra_exit}" -ne 0 ]; then
    dump_compose_failure "${compose_file}" "${log_prefix}" runner-api pinchtab pinchtab-secure pinchtab-medium pinchtab-full pinchtab-lite pinchtab-bridge
    show_suite_artifacts "${summary_file}" "${report_file}" "${progress_file}" "${log_prefix}" runner-api pinchtab pinchtab-secure pinchtab-medium pinchtab-full pinchtab-lite pinchtab-bridge
  fi
  compose_down "${compose_file}"
  return "${infra_exit}"
}

run_pr() {
  local api_exit=0
  local cli_exit=0
  local infra_exit=0

  run_api || api_exit=$?

  echo ""

  run_cli || cli_exit=$?

  echo ""

  run_infra || infra_exit=$?

  echo ""
  if [ "${api_exit}" -ne 0 ] || [ "${cli_exit}" -ne 0 ] || [ "${infra_exit}" -ne 0 ]; then
    echo "  ${ERROR}PR E2E suites failed${NC}"
    echo "  ${MUTED}exit codes: api=${api_exit}, cli=${cli_exit}, infra=${infra_exit}${NC}"
    return 1
  fi
  echo "  ${SUCCESS}PR E2E suites passed${NC}"
  return 0
}

run_release() {
  local api_exit=0
  local cli_exit=0
  local infra_exit=0

  run_api_extended || api_exit=$?

  echo ""

  run_cli_extended || cli_exit=$?

  echo ""

  run_infra_extended || infra_exit=$?

  echo ""
  if [ "${api_exit}" -ne 0 ] || [ "${cli_exit}" -ne 0 ] || [ "${infra_exit}" -ne 0 ]; then
    echo "  ${ERROR}Some E2E suites failed${NC}"
    echo "  ${MUTED}exit codes: api-extended=${api_exit}, cli-extended=${cli_exit}, infra-extended=${infra_exit}${NC}"
    return 1
  fi
  echo "  ${SUCCESS}All E2E suites passed${NC}"
  return 0
}

chmod -R 755 tests/e2e/fixtures/test-extension* 2>/dev/null || true

suite="${1:-release}"

case "${suite}" in
  pr)
    run_pr
    ;;
  api)
    run_api
    ;;
  api-extended)
    run_api_extended
    ;;
  cli)
    run_cli
    ;;
  cli-extended)
    run_cli_extended
    ;;
  infra)
    run_infra
    ;;
  infra-extended)
    run_infra_extended
    ;;
  release|all)
    run_release
    ;;
  # Backwards compatibility aliases
  api-fast)
    run_api
    ;;
  cli-fast)
    run_cli
    ;;
  api-full|full-api|curl)
    run_api_extended
    ;;
  cli-full|full-cli)
    run_cli_extended
    ;;
  *)
    echo "Unknown E2E suite: ${suite}" >&2
    echo "Available suites: pr, api, cli, infra, api-extended, cli-extended, infra-extended, release" >&2
    exit 1
    ;;
esac
