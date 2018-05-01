#!/bin/sh

# If SOURCE is a relative path (that is, it does not start with /), then it is interpreted relative to the directory
# that TARGET is in.
# https://stackoverflow.com/questions/17737065/symlink-broken-right-after-creation
RELATIVE_PATH=../../githooks

ln -s $RELATIVE_PATH/post-commit .git/hooks/post-commit
ln -s $RELATIVE_PATH/pre-commit .git/hooks/pre-commit
