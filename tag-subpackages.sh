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
  "plugins/p_academicrecords_courses"
  "plugins/p_academicrecords_programs"
  "plugins/p_announcements"
  "plugins/p_announcements_semesters"
  "plugins/p_assignmentresults"
  "plugins/p_assignments"
  "plugins/p_assignments_semesters"
  "plugins/p_contacts"
  "plugins/p_courses"
  "plugins/p_courses_teachers"
  "plugins/p_dashboard"
  "plugins/p_filesystem"
  "plugins/p_nirmancampus_programs"
  "plugins/p_nirmancampus_studentapplications"
  "plugins/p_nirmancampus_students"
  "plugins/p_nirmancampus_users"
  "plugins/p_nirmancampus_website"
  "plugins/p_otp"
  "plugins/p_programs"
  "plugins/p_pwa"
  "plugins/p_semesters"
  "plugins/p_students"
  "plugins/p_teachers"
  "plugins/p_totschool_appointments"
  "plugins/p_totschool_proposals"
  "plugins/p_totschool_tally"
  "plugins/p_totschool_users"
  "plugins/p_users"
  "registry"
  "views"
)

usage() {
  echo "Usage: $0 {major|minor|patch}" >&2
  exit 1
}

if [[ $# -ne 1 ]]; then
  usage
fi

increment_type="$1"
if [[ "$increment_type" != "major" && "$increment_type" != "minor" && "$increment_type" != "patch" ]]; then
  usage
fi

latest_tag="$(git tag --list 'v*' --sort=-version:refname | sed -n '1p')"
if [[ -z "$latest_tag" ]]; then
  echo "No tags matching v* were found." >&2
  exit 1
fi

if [[ ! "$latest_tag" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "Latest tag '$latest_tag' is not a valid semver tag like v1.2.3." >&2
  exit 1
fi

major="${BASH_REMATCH[1]}"
minor="${BASH_REMATCH[2]}"
patch="${BASH_REMATCH[3]}"

case "$increment_type" in
  major)
    ((major += 1))
    minor=0
    patch=0
    ;;
  minor)
    ((minor += 1))
    patch=0
    ;;
  patch)
    ((patch += 1))
    ;;
esac

next_version="v${major}.${minor}.${patch}"

echo "Latest tag: ${latest_tag}"
echo "Next version: ${next_version}"

for subpackage in "${SUBPACKAGES[@]}"; do
  new_tag="${subpackage}/${next_version}"
  if git rev-parse --verify --quiet "refs/tags/${new_tag}" >/dev/null; then
    echo "Tag already exists: ${new_tag}" >&2
    continue
  fi
  git tag "${new_tag}"
  echo "Created tag: ${new_tag}"
done
