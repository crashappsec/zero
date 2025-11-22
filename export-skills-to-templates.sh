#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Export Claude Code Skills to Portable Templates
# Converts .claude/skills/ to standalone, reusable prompt templates
#############################################################################

set -euo pipefail

SKILLS_DIR="${1:-.claude/skills}"
OUTPUT_DIR="${2:-$HOME/claude-templates}"

echo "=== Claude Skills to Templates Exporter ==="
echo "Source: $SKILLS_DIR"
echo "Output: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate README for templates directory
cat > "$OUTPUT_DIR/README.md" << 'EOF'
# Claude Prompt Templates Library

This directory contains portable prompt templates that can be used across:
- Claude Desktop
- Claude Web Interface
- Claude API
- Claude Code CLI

## Structure

Each template is a self-contained markdown file with:
- Clear purpose and description
- Input requirements
- The prompt template
- Usage examples for different interfaces
- Expected output format

## Usage

### Claude Desktop / Web
1. Open the template file
2. Copy the prompt section
3. Replace any `[PLACEHOLDERS]` with your actual values
4. Paste into Claude

### Claude Code
```bash
# Use directly with the cat command
claude "$(cat template-name.md | sed 's/\[PLACEHOLDER\]/actual-value/')"
```

### Claude API
```bash
# Include in API request
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -d "{
    \"model\": \"claude-sonnet-4-5-20250929\",
    \"messages\": [{\"role\": \"user\", \"content\": \"$(cat template.md)\"}]
  }"
```

## Organization

Templates are organized by domain:
- `supply-chain/` - Package analysis, vulnerability scanning
- `dora-metrics/` - DevOps metrics and analysis
- `code-ownership/` - Code ownership and succession planning
- `meta/` - Testing, documentation, and meta-prompts

## Version

Generated: $(date)
EOF

# Function to convert a skill to a template
convert_skill() {
    local skill_file="$1"
    local skill_name="$2"
    local output_file="$3"

    echo "Converting: $skill_name"

    # Extract the prompt content
    local prompt_content=$(cat "$skill_file")

    # Create template with metadata
    cat > "$output_file" << EOF
# ${skill_name} - Prompt Template

## Purpose
$(head -20 "$skill_file" | grep -m1 "^#" | sed 's/^# //' || echo "Analysis and insights for ${skill_name}")

## Source
Exported from Claude Code skill: \`${skill_file}\`
Generated: $(date)

## Required Context
- Repository path or URL
- Time period for analysis (optional)
- Specific focus areas (optional)

## Prompt Template

---

${prompt_content}

---

## Usage Examples

### Claude Desktop
1. Copy the prompt template above
2. Replace any placeholders with your values
3. Paste into a new conversation

### Claude Web
Same as Desktop - copy/paste the prompt section

### Claude Code
\`\`\`bash
# Use with the skill command
claude-code skill ${skill_name}

# Or use the template directly
claude "\$(cat $(basename "$output_file"))"
\`\`\`

### Claude API
\`\`\`bash
curl https://api.anthropic.com/v1/messages \\
  -H "x-api-key: \$ANTHROPIC_API_KEY" \\
  -H "content-type: application/json" \\
  -d "{
    \\"model\\": \\"claude-sonnet-4-5-20250929\\",
    \\"max_tokens\\": 8192,
    \\"messages\\": [{
      \\"role\\": \\"user\\",
      \\"content\\": \\"\$(cat $(basename "$output_file"))\\"}]
  }"
\`\`\`

## Customization Tips

- Adjust analysis depth based on your needs
- Combine with other templates for comprehensive reviews
- Modify output format for your tooling
- Add project-specific context as needed

## Related Templates

$(ls -1 "$(dirname "$output_file")" 2>/dev/null | grep -v "$(basename "$output_file")" | head -5 | sed 's/^/- /')

EOF

    echo "  ✓ Created: $output_file"
}

# Find and convert all skills
if [[ -d "$SKILLS_DIR" ]]; then
    echo "Scanning for skills..."

    while IFS= read -r skill_file; do
        # Get skill name from directory structure
        local skill_dir=$(dirname "$skill_file")
        local skill_name=$(basename "$skill_dir")

        # Create category directory
        local category=$(basename "$(dirname "$skill_dir")" | sed 's/skills$//')
        local output_category="$OUTPUT_DIR/$skill_name"
        mkdir -p "$output_category"

        # Output file name
        local output_file="$output_category/${skill_name}-analysis.md"

        # Convert the skill
        convert_skill "$skill_file" "$skill_name" "$output_file"

    done < <(find "$SKILLS_DIR" -name "skill.md" -type f)

    echo ""
    echo "✓ Export complete!"
    echo ""
    echo "Templates location: $OUTPUT_DIR"
    echo ""
    echo "Next steps:"
    echo "1. Review the generated templates"
    echo "2. Customize placeholders for your use cases"
    echo "3. Share with your team or commit to git"
    echo "4. Use in Claude Desktop, Web, or API"

else
    echo "Error: Skills directory not found: $SKILLS_DIR"
    exit 1
fi
