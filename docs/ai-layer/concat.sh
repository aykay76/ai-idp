#!/bin/bash

# Concatenate all markdown files in the current directory in order into one file

output="all_markdown_concatenated.mdc"

touch "$output"

for file in *.md; do
    cat "$file" >> "$output"
    echo -e "\n" >> "$output" # Add a newline between files
done

echo "Concatenation complete: $output"