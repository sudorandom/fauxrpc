#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Define source and destination directories
NODE_MODULES_DIR="./node_modules"
ASSETS_DIR="./assets"

# Ensure assets directory exists
mkdir -p "${ASSETS_DIR}"

# Function to copy and handle source maps
copy_and_handle_sourcemap() {
    local src_js="$1"
    local dest_js="$2"
    local dest_map="$3"

    cp "${src_js}" "${dest_js}"

    if [ -f "${src_js}.map" ]; then
        cp "${src_js}.map" "${dest_map}"
        # Adjust sourceMappingURL in the copied JS file to be relative
        sed -i.bak "s|//# sourceMappingURL=.*|//# sourceMappingURL=$(basename "${dest_map}")|g" "${dest_js}"
        rm "${dest_js}.bak" # Remove backup file
    else
        # If no source map, remove any existing sourceMappingURL comment
        sed -i.bak 's|//# sourceMappingURL=.*||g' "${dest_js}"
        rm "${dest_js}.bak" # Remove backup file
    fi
}

# Copy htmx-ext-sse.min.js
cp "${NODE_MODULES_DIR}/htmx-ext-sse/dist/sse.min.js" "${ASSETS_DIR}/htmx-ext-sse.min.js"

# Copy htmx.min.js
copy_and_handle_sourcemap \
    "${NODE_MODULES_DIR}/htmx.org/dist/htmx.min.js" \
    "${ASSETS_DIR}/htmx.min.js" \
    "${ASSETS_DIR}/htmx.min.js.map"

# Copy _hyperscript.min.js
copy_and_handle_sourcemap \
    "${NODE_MODULES_DIR}/hyperscript.org/dist/_hyperscript.min.js" \
    "${ASSETS_DIR}/hyperscript.min.js" \
    "${ASSETS_DIR}/hyperscript.min.js.map"

# Generate Tailwind CSS
npx tailwindcss -i ./src/input.css -o ./assets/tailwind.css --minify

echo "Assets generated successfully."