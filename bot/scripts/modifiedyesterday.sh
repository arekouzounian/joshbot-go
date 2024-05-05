# sets the modification date on a file to yesterday 

if [ $# -ne 1 ]; then 
	echo "You must specify which file to modify as an argument"
	exit 1
fi


touch -t $(date -v -1d "+%Y%m%d%H%M.%S") $1
