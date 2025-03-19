#!/usr/bin/env bash
set -euo pipefail

########################################
# GLOBALS & DEFAULTS
########################################
SKIP_DOMAIN_ROOT_FILES="true"   # Skip creating files in domain/ root by default
EXCLUDES=()
PROJECT_ROOT=""
JSON_FILE=""
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

########################################
# HELP / USAGE
########################################
show_help() {
cat <<EOF
Usage: $0 [options]

Options:
  -h, --help            Show this help and exit.
  --json <filename>     Use the specified JSON file from the current directory.
  --root <path>         Specify a custom project root path (default: current dir).
  -e, --exclude <name>  Exclude the specified directory (and subdirectories).

Behavior:
  1. If --json is provided, the script checks ./<filename> in the current directory.
     If that file does not exist, it errors out.
  2. By default (no flags), the script:
       - Uses the current directory as PROJECT_ROOT
       - Skips creating domain root files (no api.go, service.go, repo.go in domain/)
       - Does create files in domain subdirectories (service.go, repo.go, <subdir>.go).
  3. 'interfaces/api' is renamed to 'interfaces/api'.
  4. For each top-level subfolder in domain/ that is NOT 'mock', the script generates
     <subfolder>.go, service.go, and repo.go in that folder, plus a <subfolder>_api.go
     in interfaces/api.

EOF
}

########################################
# PARSE ARGS
########################################
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      show_help
      exit 0
      ;;
    --json)
      if [[ -n "${2:-}" ]]; then
        JSON_FILE="$2"
        shift 2
      else
        echo "Error: --json requires a filename."
        exit 1
      fi
      ;;
    --root)
      if [[ -n "${2:-}" ]]; then
        PROJECT_ROOT="$2"
        shift 2
      else
        echo "Error: --root requires a path."
        exit 1
      fi
      ;;
    -e|--exclude)
      if [[ -n "${2:-}" ]]; then
        EXCLUDES+=( "$2" )
        shift 2
      else
        echo "Error: --exclude requires a directory name."
        exit 1
      fi
      ;;
    *)
      echo "Unknown option: $1"
      show_help
      exit 1
      ;;
  esac
done

########################################
# DETERMINE PROJECT ROOT
########################################
if [[ -z "$PROJECT_ROOT" ]]; then
  PROJECT_ROOT="$(pwd)"
fi
echo "Using project root: $PROJECT_ROOT"

########################################
# IF --json IS PROVIDED, VERIFY FILE IN CURRENT DIR
########################################
if [[ -n "$JSON_FILE" ]]; then
  # Check if the file is in the *current directory*
  if [[ ! -f "./$JSON_FILE" ]]; then
    echo "Error: JSON file '$JSON_FILE' not found in current directory."
    exit 1
  fi
  echo "JSON file '$JSON_FILE' found in current directory."
fi

########################################
# HELPER TO CREATE DIRS
########################################
create_dir_if_needed() {
  local dir_path="$1"
  # Exclude check
  for e in "${EXCLUDES[@]}"; do
    if [[ "$dir_path" == *"$e"* ]]; then
      echo "Skipping excluded directory: $dir_path"
      return
    fi
  done
  if [[ ! -d "$dir_path" ]]; then
    mkdir -p "$dir_path"
    echo "Created: $dir_path"
  else
    echo "Directory already exists: $dir_path"
  fi
}

########################################
# CREATE BASIC DIRS
########################################
create_dir_if_needed "$PROJECT_ROOT/cmd/node"
create_dir_if_needed "$PROJECT_ROOT/docs/api"
create_dir_if_needed "$PROJECT_ROOT/docs/contributions"
create_dir_if_needed "$PROJECT_ROOT/docs/credits"
create_dir_if_needed "$PROJECT_ROOT/docs/examples"
create_dir_if_needed "$PROJECT_ROOT/docs/guides"
create_dir_if_needed "$PROJECT_ROOT/docs/images"
create_dir_if_needed "$PROJECT_ROOT/docs/sponsor"

# Domain subdirs
create_dir_if_needed "$PROJECT_ROOT/domain/auth/mockAuth"
create_dir_if_needed "$PROJECT_ROOT/domain/did/mockDID"
create_dir_if_needed "$PROJECT_ROOT/domain/proof/mockProof"

create_dir_if_needed "$PROJECT_ROOT/domain/keysManager/ephemeral"

create_dir_if_needed "$PROJECT_ROOT/infrastructure/chain"
create_dir_if_needed "$PROJECT_ROOT/infrastructure/common"
create_dir_if_needed "$PROJECT_ROOT/infrastructure/oauth"
create_dir_if_needed "$PROJECT_ROOT/infrastructure/repository"
create_dir_if_needed "$PROJECT_ROOT/infrastructure/transaction/mockTx"
create_dir_if_needed "$PROJECT_ROOT/infrastructure/zk/groth16"
create_dir_if_needed "$PROJECT_ROOT/infrastructure/clients/redis"

# Rename 'interfaces/api' to 'interfaces/api'
create_dir_if_needed "$PROJECT_ROOT/interfaces/api"
create_dir_if_needed "$PROJECT_ROOT/interfaces/events"
create_dir_if_needed "$PROJECT_ROOT/interfaces/rpc"

create_dir_if_needed "$PROJECT_ROOT/internal/migrations"
create_dir_if_needed "$PROJECT_ROOT/internal/middleware"
create_dir_if_needed "$PROJECT_ROOT/internal/server"
create_dir_if_needed "$PROJECT_ROOT/internal/bootstrapper"
create_dir_if_needed "$PROJECT_ROOT/internal/config"


create_dir_if_needed "$PROJECT_ROOT/scripts"

########################################
# CREATE BASIC GO FILES FOR NON-DOMAIN DIRS
########################################
declare -a non_domain_dirs=(
  "$PROJECT_ROOT/infrastructure/chain"
  "$PROJECT_ROOT/infrastructure/common"
  "$PROJECT_ROOT/infrastructure/oauth"
  "$PROJECT_ROOT/infrastructure/repository"
  "$PROJECT_ROOT/infrastructure/transaction"
  "$PROJECT_ROOT/infrastructure/transaction/mockTx"
  "$PROJECT_ROOT/infrastructure/zk/groth16"
  "$PROJECT_ROOT/interfaces/api"
  "$PROJECT_ROOT/interfaces/events"
  "$PROJECT_ROOT/interfaces/rpc"
  "$PROJECT_ROOT/cmd/node"
  "$PROJECT_ROOT/internal/migrations"
  "$PROJECT_ROOT/internal/middleware"
  "$PROJECT_ROOT/internal/server"
  "$PROJECT_ROOT/internal/bootstrapper"
  "$PROJECT_ROOT/internal/config"
  "$PROJECT_ROOT/infrastructure/clients/redis"
)

for dir in "${non_domain_dirs[@]}"; do
  # Exclude check
  skip="false"
  for e in "${EXCLUDES[@]}"; do
    if [[ "$dir" == *"$e"* ]]; then
      skip="true"
      break
    fi
  done
  if [[ "$skip" == "true" ]]; then
    echo "Skipping excluded path: $dir"
    continue
  fi

  pkg=$(basename "$dir" | tr '[:upper:]' '[:lower:]')
  file="$dir/${pkg}.go"
  if [[ ! -f "$file" ]]; then
    echo "package $pkg" > "$file"
    echo "Created $file with package declaration 'package $pkg'"
  else
    echo "File $file already exists; skipping."
  fi
done

########################################
# CREATE README.md in project root if missing
########################################
root_readme="$PROJECT_ROOT/README.md"
if [[ ! -f "$root_readme" ]]; then
  {
    echo "# Project Root"
    echo ""
    echo "This is the root README file for the project."
  } > "$root_readme"
  echo "Created $root_readme."
else
  echo "$root_readme already exists; skipping."
fi

########################################
# CREATE README.md in docs directory if missing
########################################
docs_readme="$PROJECT_ROOT/docs/README.md"
if [[ ! -f "$docs_readme" ]]; then
  {
    echo "# Documentation"
    echo ""
    echo "This directory contains documentation and guides."
  } > "$docs_readme"
  echo "Created $docs_readme."
else
  echo "$docs_readme already exists; skipping."
fi

########################################
# CREATE FILES IN DOMAIN SUBDIRECTORIES
# Skip "domain/" root if SKIP_DOMAIN_ROOT_FILES=true
# But do create subdomain files: service.go, repo.go, <subdir>.go
########################################

find_domain_subdirs() {
  find "$PROJECT_ROOT/domain" -mindepth 1 -type d
}

while IFS= read -r subdir; do
  # Check exclude
  skip="false"
  for e in "${EXCLUDES[@]}"; do
    if [[ "$subdir" == *"$e"* ]]; then
      skip="true"
      break
    fi
  done
  if [[ "$skip" == "true" ]]; then
    echo "Skipping excluded domain path: $subdir"
    continue
  fi

  base=$(basename "$subdir")
  lower_base=$(echo "$base" | tr '[:upper:]' '[:lower:]')

  # If this is domain root => skip if SKIP_DOMAIN_ROOT_FILES is true
  if [[ "$subdir" == "$PROJECT_ROOT/domain" ]]; then
    if [[ "$SKIP_DOMAIN_ROOT_FILES" == "true" ]]; then
      echo "Skipping root domain files creation (user default)."
    fi
    continue
  fi

  # If subdir name contains "mock", only create <basename>.go
  if [[ "$base" =~ [Mm][Oo][Cc][Kk] ]]; then
    file="$subdir/${lower_base}.go"
    if [[ ! -f "$file" ]]; then
      echo "package $lower_base" > "$file"
      echo "Created $file for mock directory with package declaration 'package $lower_base'."
    else
      echo "File $file already exists; skipping."
    fi
  else
    # For normal subdirectories => create: service.go, repo.go, <subdir>.go
    for fname in service.go repo.go "${lower_base}.go"; do
      target="$subdir/$fname"
      if [[ ! -f "$target" ]]; then
        echo "package $lower_base" > "$target"
        echo "Created $target in $subdir with package declaration 'package $lower_base'."
      else
        echo "File $target already exists; skipping."
      fi
    done
  fi
done < <(find_domain_subdirs)

########################################
# CREATE HANDLER FILES IN interfaces/api
#   For each top-level dir under domain (NOT mock), produce <domain-subdir>_api.go
########################################
for d in "$PROJECT_ROOT/domain"/*/; do
  # E.g. if domain/did => base=did => create did_api.go
  if [[ -d "$d" ]]; then
    base=$(basename "$d")
    lower_base=$(echo "$base" | tr '[:upper:]' '[:lower:]')

    # Exclude or mock check
    if [[ "$base" =~ [Mm][Oo][Cc][Kk] ]]; then
      echo "Skipping handler creation for $d (mock)."
      continue
    fi
    skip="false"
    for e in "${EXCLUDES[@]}"; do
      if [[ "$d" == *"$e"* ]]; then
        skip="true"
        break
      fi
    done
    if [[ "$skip" == "true" ]]; then
      echo "Skipping creation of handler for excluded directory $d"
      continue
    fi

    handler_file="$PROJECT_ROOT/interfaces/api/${lower_base}_api.go"
    if [[ ! -f "$handler_file" ]]; then
      {
        echo "package api"
        echo ""
        echo "// Handler for the '$base' domain."
        echo "func Handle${base^}() {"
        echo "    // TODO: Implement handler logic for domain '$base'."
        echo "}"
      } > "$handler_file"
      echo "Created $handler_file for domain '$base'."
    else
      echo "Handler file $handler_file already exists; skipping."
    fi
  fi
done

echo "Script completed successfully."
