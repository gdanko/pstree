#/bin/bash

REPO_ROOT=$(git rev-parse --show-toplevel)
if [ -z $REPO_ROOT ]; then
    echo "Failed to determine the repo root. Cannot continue."
    exit 1
fi

if [ -z $PSTREE_VERSION ]; then
    echo "PSTREE_VERSION is not set. Cannot continue."
    exit 1
fi

# Makefile
MAKEFILE="${REPO_ROOT}/Makefile"
MAKEFILE_BACKUP="$MAKEFILE.bak"

if [[ ! -f "$MAKEFILE" ]]; then
    echo "Error: $MAKEFILE not found."
    exit 1
fi

echo "Backing up $MAKEFILE to $MAKEFILE_BACKUP"
cp $MAKEFILE $MAKEFILE_BACKUP
if [ $? -ne 0 ]; then
    echo "Failed to backup $MAKEFILE. Cannot continue"
    exit 1
fi

echo "Updating PSTREE_VERSION in $MAKEFILE to $PSTREE_VERSION"
sed -i "s/^PSTREE_VERSION\s:*=\s*.*/PSTREE_VERSION := $PSTREE_VERSION/" "$MAKEFILE"
if [ $? -ne 0 ]; then
    echo "Failed to update $MAKEFILE. Cannot continue"
    exit 1
fi

echo "Verifying the changes to $MAKEFILE"
grep -E "^PSTREE_VERSION\s*:=\s*$PSTREE_VERSION" "$MAKEFILE" 1>/dev/null
if [ $? -ne 0 ]; then
    echo "Failed to update $MAKEFILE. Cannot continue"
    exit 1
fi

# root.go
ROOT_CMD="${REPO_ROOT}/cmd/root.go"
ROOT_CMD_BACKUP="$ROOT_CMD.bak"

if [[ ! -f "$ROOT_CMD" ]]; then
    echo "Error: $ROOT_CMD not found."
    exit 1
fi

echo "Backing up $ROOT_CMD to $ROOT_CMD_BACKUP"
cp $ROOT_CMD $ROOT_CMD_BACKUP
if [ $? -ne 0 ]; then
    echo "Failed to backup $ROOT_CMD. Cannot continue"
    exit 1
fi

echo "Updating version in $ROOT_CMD to $PSTREE_VERSION"
sed -i "s/^\s*version\s*string\s*=\s*\".*\"/\tversion               string = \"$PSTREE_VERSION\"/" "$ROOT_CMD"
if [ $? -ne 0 ]; then
    echo "Failed to update $ROOT_CMD. Cannot continue"
    exit 1
fi

echo Reformatting $ROOT_CMD
go fmt $ROOT_CMD

echo "Verifying the changes to $ROOT_CMD"
grep -E "^\s*version\s*string\s*=\s*\"$PSTREE_VERSION\"" "$ROOT_CMD" 1>/dev/null
if [ $? -ne 0 ]; then
    echo "Failed to update $ROOT_CMD. Cannot continue"
    exit 1
fi
