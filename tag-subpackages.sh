#!/usr/bin/env bash

set -euo pipefail

readonly SUBPACKAGES=(
  "components"
  "deployments/lariv"
  "deployments/nirmancampus"
  "deployments/totschool_lago"
  "getters"
  "lago"
  "plugins/p_academicrecords"
  "plugins/p_nirmancampus_announcements"
  "plugins/p_nirmancampus_assignmentresults"
  "plugins/p_nirmancampus_assignments"
  "plugins/p_contacts"
  "plugins/p_nirmancampus_courses"
  "plugins/p_dashboard"
  "plugins/p_filesystem"
  "plugins/p_nirmancampus_programs"
  "plugins/p_nirmancampus_studentapplications"
  "plugins/p_nirmancampus_students"
  "plugins/p_nirmancampus_users"
  "plugins/p_nirmancampus_website"
  "plugins/p_otp"
  "plugins/p_pwa"
  "plugins/p_nirmancampus_sessions"
  "plugins/p_totschool_appointments"
  "plugins/p_totschool_proposals"
  "plugins/p_totschool_tally"
  "plugins/p_totschool_users"
  "plugins/p_users"
  "registry"
  "views"
)

usage() {
  echo "Usage: $0 <tag>" >&2
  exit 1
}

if [[ $# -ne 1 ]]; then
  usage
fi

tag="$1"

for subpackage in "${SUBPACKAGES[@]}"; do
  new_tag="${subpackage}/${tag}"
  if git rev-parse --verify --quiet "refs/tags/${new_tag}" >/dev/null; then
    echo "Tag already exists: ${new_tag}" >&2
    continue
  fi
  git tag "${new_tag}"
  echo "Created tag: ${new_tag}"
done
