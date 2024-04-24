#!/bin/bash

# Will attempt to create symbolic links to the relevant tables
if [ $# -eq 0 ]; then 
    echo "Usage: ./link-tables.sh <table-directory>"
    exit 1 
fi

ln -s $1/*.csv .
