#!/bin/sh

RELATIVE_PATH=githooks

ln -s $RELATIVE_PATH/post-commit .git/hooks/post-commit
ln -s $RELATIVE_PATH/pre-commit .git/hooks/pre-commit
