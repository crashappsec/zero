# Python

**Category**: languages
**Description**: Python programming language - versatile, readable, and widely used for web, data science, AI/ML, and scripting
**Homepage**: https://www.python.org

## Package Detection

### PYPI
- `python`
- `cpython`
- `setuptools`
- `pip`
- `wheel`
- `poetry`
- `pipenv`

## Configuration Files

- `requirements.txt`
- `requirements-*.txt`
- `setup.py`
- `setup.cfg`
- `pyproject.toml`
- `Pipfile`
- `Pipfile.lock`
- `poetry.lock`
- `tox.ini`
- `.python-version`
- `runtime.txt`
- `__init__.py`
- `conftest.py`

## File Extensions

- `.py`
- `.pyi` (type stubs)
- `.pyx` (Cython)
- `.pxd` (Cython declarations)
- `.ipynb` (Jupyter notebooks)

## Import Detection

### Python
**Pattern**: `#!/usr/bin/env python`
- Shebang line
- Example: `#!/usr/bin/env python3`

**Pattern**: `^import\s+\w+`
- Standard import
- Example: `import os`

**Pattern**: `^from\s+\w+\s+import`
- From import
- Example: `from typing import List`

## Environment Variables

- `PYTHONPATH`
- `PYTHONHOME`
- `PYTHON_VERSION`
- `VIRTUAL_ENV`
- `CONDA_DEFAULT_ENV`

## Version Indicators

- Python 3.12+ (current)
- Python 3.11 (stable)
- Python 3.10 (maintenance)
- Python 3.9 (security fixes only)
- Python 2.7 (end of life - security risk)

## Detection Notes

- Look for `.py` files in repository
- Check for pyproject.toml (modern Python projects)
- requirements.txt indicates dependency management
- `__init__.py` files indicate Python packages
- Check for virtual environment directories (venv, .venv, env)

## Detection Confidence

- **File Extension Detection**: 95% (HIGH)
- **Configuration File Detection**: 95% (HIGH)
- **Package Detection**: 90% (HIGH)
