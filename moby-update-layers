#!/bin/bash

set -e

list_layers() {
	git config --name-only --get-regexp 'layer\.[^.]*\.url' |
	sed -e 's/^layer\.//' -e 's/\.url$//' |
	{
		while read layer; do
			if [ $(git config --get layer.$layer.skip) ]; then
				echo >&2 "[$layer] Skipping"
				continue
			fi
			echo "$layer"
		done
	}
}

update_all_sources() {
	echo "--- Checking layer remote sources for updates"
	list_layers | {
		while read layer; do
			update_source $layer
		done
	}
}

update_source() {
	layer=$1
	url=$(git config --get layer.$layer.url)
	branch=$(git config --get layer.$layer.branch)
	echo "[$layer] fetching $url $branch"
	git fetch $url $branch:layers/$layer/src
}


update_all_caches() {
	echo "--- Updating layer caches"
	list_layers | {
		while read layer; do
			update_cache $layer
		done
	}
}

update_cache() {
	layer=$1
	src=layers/$layer/src
	tmp=layers/$layer/tmp/$RANDOM
	cache=layers/$layer/cache
	git branch $tmp $src
	mountpoint=$(git config --get layer.$layer.mountpoint)
	transform_branch $tmp $mountpoint
	echo "[$layer] moving $tmp to $cache"
	git branch -m -f $tmp $cache
}

transform_branch() {
	branch=$1
	mountpoint=$2

	if [ -z "$mountpoint" ]; then
		return
	fi
	backup
	git filter-branch -f --index-filter "
		git ls-files -s |
		{
			while read line; do
				echo \"\${line/	/	$mountpoint/}\"
			done
		} |
		GIT_INDEX_FILE=\$GIT_INDEX_FILE.new git update-index --index-info &&
		if test -f \"\$GIT_INDEX_FILE.new\"; then
			mv \$GIT_INDEX_FILE.new \$GIT_INDEX_FILE;
		fi
	" $branch
	restore
}

backup() {
	echo "--- Backing up work tree"
	git stash save -u --all
	git update-ref LAYER_UPDATE_HEAD HEAD
}

restore() {
	echo "--- Restoring work tree"
	git update-ref HEAD LAYER_UPDATE_HEAD
	git checkout -f HEAD
	git stash pop
}

update_all_sources
update_all_caches
