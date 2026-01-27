#!/bin/sh

# Replace environment variables in the built JavaScript files
# This allows runtime configuration of the Supabase URL and key

# Find and replace placeholder values with actual environment variables
for file in /usr/share/nginx/html/assets/*.js; do
    if [ -f "$file" ]; then
        # Replace Supabase URL placeholder
        if [ -n "$VITE_SUPABASE_URL" ]; then
            sed -i "s|http://localhost:54321|$VITE_SUPABASE_URL|g" "$file"
        fi
        # Replace Supabase Anon Key placeholder
        if [ -n "$VITE_SUPABASE_ANON_KEY" ]; then
            sed -i "s|eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0|$VITE_SUPABASE_ANON_KEY|g" "$file"
        fi
    fi
done

# Execute the main command
exec "$@"
