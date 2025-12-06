#!/usr/bin/env python3
"""
RAG-to-Semgrep Converter

Converts Gibson Powers RAG pattern markdown files to Semgrep YAML rules.

Usage:
    python3 rag-to-semgrep.py <rag_dir> <output_dir>

Example:
    python3 rag-to-semgrep.py ../../rag/technology-identification ./rules
"""

import os
import re
import sys
import yaml
from pathlib import Path
from typing import Dict, List, Optional, Tuple


class PatternParser:
    """Parses patterns.md files into structured data."""

    def __init__(self, file_path: str):
        self.file_path = file_path
        self.content = Path(file_path).read_text()
        self.data = {
            'name': '',
            'category': '',
            'description': '',
            'packages': {'npm': [], 'pypi': [], 'go': [], 'rubygems': [], 'maven': []},
            'imports': {'python': [], 'javascript': [], 'go': [], 'ruby': [], 'java': []},
            'env_vars': [],
            'secrets': [],
            'confidence': {}
        }
        self._parse()

    def _parse(self):
        """Parse the markdown file into structured data."""
        lines = self.content.split('\n')

        # Extract name from first heading
        for line in lines:
            if line.startswith('# '):
                self.data['name'] = line[2:].strip()
                break

        # Extract category
        cat_match = re.search(r'\*\*Category\*\*:\s*(.+)', self.content)
        if cat_match:
            self.data['category'] = cat_match.group(1).strip()

        # Extract description
        desc_match = re.search(r'\*\*Description\*\*:\s*(.+)', self.content)
        if desc_match:
            self.data['description'] = desc_match.group(1).strip()

        # Parse packages
        self._parse_packages()

        # Parse import patterns
        self._parse_imports()

        # Parse environment variables
        self._parse_env_vars()

        # Parse secrets
        self._parse_secrets()

        # Parse confidence scores
        self._parse_confidence()

    def _parse_packages(self):
        """Parse package detection section."""
        pkg_section = re.search(r'## Package Detection\s*\n(.*?)(?=\n##|\Z)', self.content, re.DOTALL)
        if not pkg_section:
            return

        section = pkg_section.group(1)

        # NPM packages
        npm_match = re.search(r'### NPM.*?\n(.*?)(?=\n###|\n##|\Z)', section, re.DOTALL)
        if npm_match:
            packages = re.findall(r'^-\s*`([^`]+)`', npm_match.group(1), re.MULTILINE)
            self.data['packages']['npm'] = packages

        # PyPI packages
        pypi_match = re.search(r'### PYPI.*?\n(.*?)(?=\n###|\n##|\Z)', section, re.DOTALL)
        if pypi_match:
            packages = re.findall(r'^-\s*`([^`]+)`', pypi_match.group(1), re.MULTILINE)
            self.data['packages']['pypi'] = packages

        # Go packages
        go_match = re.search(r'### Go.*?\n(.*?)(?=\n###|\n##|\Z)', section, re.DOTALL)
        if go_match:
            packages = re.findall(r'^-\s*`([^`]+)`', go_match.group(1), re.MULTILINE)
            self.data['packages']['go'] = packages

    def _parse_imports(self):
        """Parse import detection patterns."""
        import_section = re.search(r'## Import Detection\s*\n(.*?)(?=\n## [^I]|\Z)', self.content, re.DOTALL)
        if not import_section:
            return

        section = import_section.group(1)

        # Python patterns
        py_match = re.search(r'### Python.*?\n(.*?)(?=\n###|\n##|\Z)', section, re.DOTALL)
        if py_match:
            patterns = re.findall(r'\*\*Pattern\*\*:\s*`([^`]+)`', py_match.group(1))
            self.data['imports']['python'] = patterns

        # Javascript/Typescript patterns
        js_match = re.search(r'### Javascript.*?\n(.*?)(?=\n###|\n##|\Z)', section, re.DOTALL)
        if js_match:
            patterns = re.findall(r'\*\*Pattern\*\*:\s*`([^`]+)`', js_match.group(1))
            self.data['imports']['javascript'] = patterns

        # Go patterns
        go_match = re.search(r'### Go\s*\n(.*?)(?=\n###|\n##|\Z)', section, re.DOTALL)
        if go_match:
            patterns = re.findall(r'\*\*Pattern\*\*:\s*`([^`]+)`', go_match.group(1))
            self.data['imports']['go'] = patterns

    def _parse_env_vars(self):
        """Parse environment variables section."""
        env_section = re.search(r'## Environment Variables\s*\n(.*?)(?=\n##|\Z)', self.content, re.DOTALL)
        if not env_section:
            return

        vars_found = re.findall(r'^-\s*`([^`]+)`', env_section.group(1), re.MULTILINE)
        self.data['env_vars'] = vars_found

    def _parse_secrets(self):
        """Parse secrets detection patterns."""
        secrets_section = re.search(r'## Secrets Detection\s*\n(.*?)(?=\n## [^S#]|\Z)', self.content, re.DOTALL)
        if not secrets_section:
            return

        section = secrets_section.group(1)

        # Find each secret pattern block
        pattern_blocks = re.findall(
            r'####\s*(.+?)\n.*?\*\*Pattern\*\*:\s*`([^`]+)`.*?\*\*Severity\*\*:\s*(\w+)',
            section, re.DOTALL
        )

        for name, pattern, severity in pattern_blocks:
            self.data['secrets'].append({
                'name': name.strip(),
                'pattern': pattern.strip(),
                'severity': severity.strip().upper()
            })

    def _parse_confidence(self):
        """Parse confidence scores."""
        conf_section = re.search(r'## Detection Confidence\s*\n(.*?)(?=\n##|\Z)', self.content, re.DOTALL)
        if not conf_section:
            return

        matches = re.findall(r'\*\*([^*]+)\*\*:\s*(\d+)%', conf_section.group(1))
        for name, score in matches:
            key = name.strip().lower().replace(' ', '_')
            self.data['confidence'][key] = int(score)


def regex_to_semgrep(regex_pattern: str, language: str) -> Optional[str]:
    """
    Convert regex pattern to Semgrep pattern.
    Returns None if pattern can't be converted to valid Semgrep syntax.
    """
    # Simple patterns that map directly
    pattern = regex_pattern

    # Handle Python imports
    if language == 'python':
        # `^import openai` -> `import openai`
        if pattern.startswith('^import '):
            module = pattern[8:]
            return f'import {module}'

        # `^from openai import` -> `from openai import $X`
        if pattern.startswith('^from ') and 'import' in pattern:
            match = re.match(r'\^from\s+(\S+)\s+import', pattern)
            if match:
                module = match.group(1)
                return f'from {module} import $X'

        # `from anthropic import|import anthropic` -> pattern-either
        if '|' in pattern:
            return None  # Handle separately as pattern-either

        # Function/class patterns like `Anthropic\(`
        if pattern.endswith('\\('):
            name = pattern[:-2]
            return f'{name}(...)'

    # Handle JavaScript/TypeScript imports
    if language in ('javascript', 'typescript'):
        # `from ['"]openai['"]` patterns
        if "from ['\"]" in pattern:
            match = re.search(r"from \['\"\]([^[]+)\['\"\]", pattern)
            if match:
                module = match.group(1)
                return f'import $X from "{module}"'

        # `require\(['"]...` patterns
        if "require\\(['\"]" in pattern:
            match = re.search(r"require\\\(\['\"\]([^[]+)\['\"\]", pattern)
            if match:
                module = match.group(1)
                return f'require("{module}")'

        # `new OpenAI\(` patterns
        if pattern.startswith('new ') and pattern.endswith('\\('):
            class_name = pattern[4:-2]
            return f'new {class_name}(...)'

    # Handle Go imports
    if language == 'go':
        if 'import' in pattern:
            # Extract package path
            match = re.search(r'"([^"]+)"', pattern)
            if match:
                pkg = match.group(1)
                return f'import "{pkg}"'

    return None


def create_semgrep_rule(
    rule_id: str,
    technology: str,
    category: str,
    patterns: List[Dict],
    severity: str = 'INFO',
    message: str = None
) -> Dict:
    """Create a Semgrep rule dictionary."""

    if not message:
        message = f"{technology} usage detected"

    rule = {
        'id': rule_id,
        'message': message,
        'severity': severity,
        'metadata': {
            'technology': technology,
            'category': category,
            'scanner': 'tech-discovery'
        }
    }

    # Group patterns by language
    lang_patterns = {}
    for p in patterns:
        lang = p.get('language', 'generic')
        if lang not in lang_patterns:
            lang_patterns[lang] = []
        lang_patterns[lang].append(p['pattern'])

    # If multiple languages, we need separate rules
    if len(lang_patterns) == 1:
        lang = list(lang_patterns.keys())[0]
        pats = list(lang_patterns.values())[0]

        rule['languages'] = [lang] if lang != 'generic' else ['python', 'javascript', 'typescript', 'go']

        if len(pats) == 1:
            rule['pattern'] = pats[0]
        else:
            rule['pattern-either'] = [{'pattern': p} for p in pats]
    else:
        # Multiple languages - use pattern-either with each pattern
        all_patterns = []
        all_langs = set()
        for lang, pats in lang_patterns.items():
            all_langs.add(lang)
            all_patterns.extend(pats)

        rule['languages'] = list(all_langs)
        rule['pattern-either'] = [{'pattern': p} for p in all_patterns]

    return rule


def convert_technology(parser: PatternParser, base_id: str) -> List[Dict]:
    """Convert a parsed technology pattern file to Semgrep rules."""
    rules = []
    data = parser.data

    if not data['name']:
        return rules

    # Clean technology name for rule ID
    tech_id = re.sub(r'[^a-z0-9]+', '-', data['name'].lower()).strip('-')
    category = data['category'] or 'unknown'

    # Create import detection rules
    import_patterns = []

    for lang, patterns in data['imports'].items():
        semgrep_lang = {
            'python': 'python',
            'javascript': 'javascript',
            'go': 'go',
            'ruby': 'ruby',
            'java': 'java'
        }.get(lang, lang)

        for regex in patterns:
            semgrep_pattern = regex_to_semgrep(regex, lang)
            if semgrep_pattern:
                import_patterns.append({
                    'language': semgrep_lang,
                    'pattern': semgrep_pattern
                })

    if import_patterns:
        # Group by language for better rules
        by_lang = {}
        for p in import_patterns:
            lang = p['language']
            if lang not in by_lang:
                by_lang[lang] = []
            by_lang[lang].append(p['pattern'])

        for lang, patterns in by_lang.items():
            rule_id = f"{base_id}.{tech_id}.import.{lang}"

            rule = {
                'id': rule_id,
                'message': f"{data['name']} library import detected",
                'severity': 'INFO',
                'languages': [lang],
                'metadata': {
                    'technology': data['name'],
                    'category': category,
                    'detection_type': 'import',
                    'confidence': data['confidence'].get('import_detection', 90)
                }
            }

            if len(patterns) == 1:
                rule['pattern'] = patterns[0]
            else:
                rule['pattern-either'] = [{'pattern': p} for p in patterns]

            rules.append(rule)

    # Create secrets detection rules
    for secret in data['secrets']:
        rule_id = f"{base_id}.{tech_id}.secret.{re.sub(r'[^a-z0-9]+', '-', secret['name'].lower())}"

        severity_map = {
            'CRITICAL': 'ERROR',
            'HIGH': 'WARNING',
            'MEDIUM': 'WARNING',
            'LOW': 'INFO'
        }

        rule = {
            'id': rule_id,
            'message': f"Potential {data['name']} {secret['name']} exposed",
            'severity': severity_map.get(secret['severity'], 'WARNING'),
            'languages': ['generic'],
            'metadata': {
                'technology': data['name'],
                'category': 'secrets',
                'secret_type': secret['name'],
                'confidence': 95
            },
            'pattern-regex': secret['pattern']
        }

        rules.append(rule)

    return rules


def process_rag_directory(rag_dir: str, output_dir: str):
    """Process all RAG pattern files and generate Semgrep rules."""

    rag_path = Path(rag_dir)
    output_path = Path(output_dir)
    output_path.mkdir(parents=True, exist_ok=True)

    all_rules = {
        'tech-discovery': [],
        'secrets': [],
        'tech-debt': []
    }

    # Find all patterns.md files
    pattern_files = list(rag_path.rglob('patterns.md'))
    print(f"Found {len(pattern_files)} pattern files")

    for pf in pattern_files:
        try:
            # Determine base ID from path
            rel_path = pf.relative_to(rag_path)
            base_id = str(rel_path.parent).replace('/', '.').replace('\\', '.')

            parser = PatternParser(str(pf))
            rules = convert_technology(parser, f"gibson.{base_id}")

            for rule in rules:
                if 'secret' in rule['id']:
                    all_rules['secrets'].append(rule)
                else:
                    all_rules['tech-discovery'].append(rule)

            if rules:
                print(f"  Converted: {pf.parent.name} -> {len(rules)} rules")
        except Exception as e:
            print(f"  Error processing {pf}: {e}")

    # Write output files
    for category, rules in all_rules.items():
        if rules:
            output_file = output_path / f"{category}.yaml"
            with open(output_file, 'w') as f:
                yaml.dump({'rules': rules}, f, default_flow_style=False, sort_keys=False, allow_unicode=True)
            print(f"\nWrote {len(rules)} rules to {output_file}")

    # Summary
    total = sum(len(r) for r in all_rules.values())
    print(f"\n{'='*50}")
    print(f"Total: {total} Semgrep rules generated")
    print(f"  - tech-discovery: {len(all_rules['tech-discovery'])} rules")
    print(f"  - secrets: {len(all_rules['secrets'])} rules")


def add_tech_debt_rules(output_dir: str):
    """Add standard tech debt detection rules."""

    rules = [
        {
            'id': 'gibson.tech-debt.todo',
            'message': 'TODO marker found: $MSG',
            'severity': 'INFO',
            'languages': ['generic'],
            'metadata': {
                'category': 'tech-debt',
                'debt_type': 'todo',
                'priority': 'low'
            },
            'pattern-regex': r'TODO[:\s]+(.+?)(?:\n|$)'
        },
        {
            'id': 'gibson.tech-debt.fixme',
            'message': 'FIXME marker found: $MSG',
            'severity': 'WARNING',
            'languages': ['generic'],
            'metadata': {
                'category': 'tech-debt',
                'debt_type': 'fixme',
                'priority': 'medium'
            },
            'pattern-regex': r'FIXME[:\s]+(.+?)(?:\n|$)'
        },
        {
            'id': 'gibson.tech-debt.hack',
            'message': 'HACK marker found',
            'severity': 'WARNING',
            'languages': ['generic'],
            'metadata': {
                'category': 'tech-debt',
                'debt_type': 'hack',
                'priority': 'high'
            },
            'pattern-regex': r'HACK[:\s]+(.+?)(?:\n|$)'
        },
        {
            'id': 'gibson.tech-debt.xxx',
            'message': 'XXX marker found (needs attention)',
            'severity': 'WARNING',
            'languages': ['generic'],
            'metadata': {
                'category': 'tech-debt',
                'debt_type': 'xxx',
                'priority': 'high'
            },
            'pattern-regex': r'XXX[:\s]+(.+?)(?:\n|$)'
        },
        {
            'id': 'gibson.tech-debt.deprecated-decorator',
            'message': '@deprecated usage found',
            'severity': 'INFO',
            'languages': ['python', 'javascript', 'typescript'],
            'metadata': {
                'category': 'tech-debt',
                'debt_type': 'deprecated',
                'priority': 'medium'
            },
            'pattern': '@deprecated'
        }
    ]

    output_path = Path(output_dir)
    output_file = output_path / "tech-debt.yaml"

    with open(output_file, 'w') as f:
        yaml.dump({'rules': rules}, f, default_flow_style=False, sort_keys=False)

    print(f"Wrote {len(rules)} tech-debt rules to {output_file}")


if __name__ == '__main__':
    if len(sys.argv) < 3:
        print("Usage: python3 rag-to-semgrep.py <rag_dir> <output_dir>")
        print("Example: python3 rag-to-semgrep.py ../../rag/technology-identification ./rules")
        sys.exit(1)

    rag_dir = sys.argv[1]
    output_dir = sys.argv[2]

    if not os.path.isdir(rag_dir):
        print(f"Error: RAG directory not found: {rag_dir}")
        sys.exit(1)

    print(f"Converting RAG patterns from: {rag_dir}")
    print(f"Output directory: {output_dir}")
    print("=" * 50)

    process_rag_directory(rag_dir, output_dir)
    add_tech_debt_rules(output_dir)
