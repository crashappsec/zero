#!/usr/bin/env python3
"""
RAG Security Pattern to Semgrep Converter

Converts RAG security pattern markdown files (API security, crypto, etc.) to Semgrep YAML rules.

Pattern format in markdown:
```
PATTERN: <regex pattern>
LANGUAGES: javascript, typescript, python
```

Metadata before pattern block:
CATEGORY: api-auth
SEVERITY: critical
CONFIDENCE: 90
CWE: CWE-306
OWASP: API2:2023

Usage:
    python3 rag-security-to-semgrep.py <rag_dir> <output_file>
"""

import os
import re
import sys
import yaml
from pathlib import Path
from typing import Dict, List, Optional, Tuple


def parse_security_patterns(content: str) -> List[Dict]:
    """Parse security patterns from markdown content."""
    patterns = []

    # Current metadata state
    current_meta = {
        'category': '',
        'severity': 'medium',
        'confidence': 80,
        'cwe': [],
        'owasp': [],
        'description': ''
    }

    lines = content.split('\n')
    i = 0

    while i < len(lines):
        line = lines[i].strip()

        # Update metadata when we see these markers
        if line.startswith('CATEGORY:'):
            current_meta['category'] = line.split(':', 1)[1].strip()
        elif line.startswith('SEVERITY:'):
            current_meta['severity'] = line.split(':', 1)[1].strip().lower()
        elif line.startswith('CONFIDENCE:'):
            try:
                current_meta['confidence'] = int(line.split(':', 1)[1].strip())
            except:
                pass
        elif line.startswith('CWE:'):
            cwe = line.split(':', 1)[1].strip()
            if cwe and cwe.lower() != 'none':
                current_meta['cwe'] = [c.strip() for c in cwe.split(',')]
        elif line.startswith('OWASP:'):
            owasp = line.split(':', 1)[1].strip()
            if owasp and owasp.lower() != 'none':
                current_meta['owasp'] = [o.strip() for o in owasp.split(',')]

        # Capture description from headers
        elif line.startswith('### '):
            current_meta['description'] = line[4:].strip()

        # Look for code blocks containing patterns
        elif line.startswith('```'):
            # Read until closing ```
            block_lines = []
            i += 1
            while i < len(lines) and not lines[i].strip().startswith('```'):
                block_lines.append(lines[i])
                i += 1

            block_content = '\n'.join(block_lines)

            # Check if this block contains PATTERN:
            pattern_match = re.search(r'PATTERN:\s*(.+?)(?:\n|$)', block_content)
            lang_match = re.search(r'LANGUAGES:\s*(.+?)(?:\n|$)', block_content)

            if pattern_match:
                pattern_regex = pattern_match.group(1).strip()
                languages = ['generic']
                if lang_match:
                    languages = [l.strip() for l in lang_match.group(1).split(',')]

                # Create pattern entry
                patterns.append({
                    'category': current_meta['category'],
                    'severity': current_meta['severity'],
                    'confidence': current_meta['confidence'],
                    'cwe': current_meta['cwe'].copy(),
                    'owasp': current_meta['owasp'].copy(),
                    'description': current_meta['description'],
                    'regex': pattern_regex,
                    'languages': languages
                })

        i += 1

    return patterns


def severity_to_semgrep(severity: str) -> str:
    """Convert severity to Semgrep severity."""
    mapping = {
        'critical': 'ERROR',
        'high': 'ERROR',
        'medium': 'WARNING',
        'low': 'INFO',
        'info': 'INFO'
    }
    return mapping.get(severity.lower(), 'WARNING')


def create_rule_id(category: str, description: str, index: int) -> str:
    """Create a unique rule ID."""
    # Clean description for ID
    desc_id = re.sub(r'[^a-z0-9]+', '-', description.lower())[:50].strip('-')
    if not desc_id:
        desc_id = f"pattern-{index}"
    return f"zero.{category}.{desc_id}"


def convert_patterns_to_rules(patterns: List[Dict], base_category: str) -> List[Dict]:
    """Convert parsed patterns to Semgrep rules."""
    rules = []
    seen_ids = set()

    for i, p in enumerate(patterns):
        category = p['category'] or base_category
        description = p['description'] or f"{category} security issue"

        rule_id = create_rule_id(category, description, i)

        # Ensure unique IDs
        base_id = rule_id
        counter = 1
        while rule_id in seen_ids:
            rule_id = f"{base_id}-{counter}"
            counter += 1
        seen_ids.add(rule_id)

        # Map languages
        lang_mapping = {
            'javascript': 'javascript',
            'typescript': 'typescript',
            'python': 'python',
            'java': 'java',
            'go': 'go',
            'ruby': 'ruby',
            'generic': 'generic',
            'yaml': 'yaml',
            'json': 'json',
            'graphql': 'generic'
        }

        languages = []
        for lang in p['languages']:
            mapped = lang_mapping.get(lang.lower().strip(), 'generic')
            if mapped not in languages:
                languages.append(mapped)

        # Add typescript if we have javascript
        if 'javascript' in languages and 'typescript' not in languages:
            languages.append('typescript')

        rule = {
            'id': rule_id,
            'message': f"{description}",
            'severity': severity_to_semgrep(p['severity']),
            'languages': languages if languages else ['generic'],
            'metadata': {
                'category': category,
                'confidence': p['confidence'],
                'scanner': base_category
            },
            'pattern-regex': p['regex']
        }

        # Add CWE/OWASP if present
        if p['cwe']:
            rule['metadata']['cwe'] = p['cwe']
        if p['owasp']:
            rule['metadata']['owasp'] = p['owasp']

        rules.append(rule)

    return rules


def process_directory(rag_dir: str, output_file: str):
    """Process all markdown files in the RAG directory."""
    rag_path = Path(rag_dir)

    if not rag_path.exists():
        print(f"Error: Directory not found: {rag_dir}")
        sys.exit(1)

    # Determine base category from directory name
    base_category = rag_path.name

    all_patterns = []

    # Process all .md files
    md_files = sorted(rag_path.glob('*.md'))
    print(f"Found {len(md_files)} markdown files in {rag_dir}")

    for md_file in md_files:
        if md_file.name.startswith('_'):
            continue

        print(f"  Processing: {md_file.name}")
        try:
            content = md_file.read_text()
            patterns = parse_security_patterns(content)
            all_patterns.extend(patterns)
            print(f"    -> {len(patterns)} patterns extracted")
        except Exception as e:
            print(f"    Error: {e}")
            import traceback
            traceback.print_exc()

    # Convert to rules
    rules = convert_patterns_to_rules(all_patterns, base_category)

    # Write output
    output_path = Path(output_file)
    output_path.parent.mkdir(parents=True, exist_ok=True)

    with open(output_path, 'w') as f:
        yaml.dump({'rules': rules}, f, default_flow_style=False, sort_keys=False, allow_unicode=True)

    print(f"\nWrote {len(rules)} rules to {output_file}")

    # Print summary by category
    categories = {}
    for r in rules:
        cat = r['metadata']['category']
        categories[cat] = categories.get(cat, 0) + 1

    print("\nRules by category:")
    for cat, count in sorted(categories.items()):
        print(f"  {cat}: {count}")

    return len(rules)


if __name__ == '__main__':
    if len(sys.argv) < 3:
        print("Usage: python3 rag-security-to-semgrep.py <rag_dir> <output_file>")
        print("Example: python3 rag-security-to-semgrep.py ../../../rag/api-security ./rules/api-security.yaml")
        sys.exit(1)

    rag_dir = sys.argv[1]
    output_file = sys.argv[2]

    print(f"Converting security patterns from: {rag_dir}")
    print(f"Output file: {output_file}")
    print("=" * 50)

    process_directory(rag_dir, output_file)
