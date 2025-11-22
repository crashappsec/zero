<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Chalk Build Analyser Skill

## 🎯 Purpose

The **Chalk Build Analyser** is an AI-powered Claude skill specifically designed for analyzing build performance and engineering efficiency from Chalk reports. Unlike security-focused tools, this skill concentrates on what matters most for day-to-day engineering: **build speed**, **team velocity**, and **operational excellence**.

## 🚀 Key Features

### Build Performance Analysis
- **Stage-by-Stage Breakdown**: Understand where time is spent (compile, test, package)
- **Bottleneck Detection**: Automatically identify slow stages holding up your pipeline
- **Resource Utilization**: Monitor CPU, memory, and infrastructure efficiency
- **Queue Time Analysis**: Detect and resolve resource contention issues

### Engineering Velocity Metrics
- **DORA Metrics**: Track deployment frequency, lead time, MTTR, and change failure rate
- **Team Productivity**: Measure builds per developer, collaboration patterns
- **Success Rate Tracking**: Monitor build reliability and flakiness
- **Velocity Scoring**: Get a single score representing overall engineering efficiency

### Regression Detection
- **Performance Comparison**: Detect when builds get slower
- **Impact Analysis**: Understand which stages are affected
- **Quality Changes**: Track test coverage and pass rate evolution
- **Efficiency Monitoring**: Cache effectiveness and parallelization changes

### Trend Analysis & Predictions
- **Historical Patterns**: Visualize performance over time
- **Cost Trends**: Track and optimize build infrastructure spending
- **Team Patterns**: Understand developer productivity cycles
- **Predictive Insights**: Forecast future performance based on trends

## 💡 Real-World Use Cases

### For Individual Developers

**"Why is my build slow?"**
```bash
python analyze_build.py my_build.json
```
Get immediate insights:
- Your tests take 48% of build time - consider parallelization
- Cache hit rate is only 65% - check your dependency management
- Queue time is 45 seconds - builds are backing up

**"Did I make the build slower?"**
```bash
python compare_builds.py before_my_changes.json after_my_changes.json
```
See the impact:
- Build time increased by 27% (+88 seconds)
- Test stage is the main contributor (+52% slower)
- Added 200 new tests causing the slowdown

### For Engineering Leaders

**"How's our engineering velocity?"**
Get DORA metrics and team insights:
- Deployment frequency: 12 builds/day
- Lead time: 324 seconds average
- MTTR: 15 minutes
- 5 active developers with 96% success rate

**"Where should we invest in infrastructure?"**
Resource analysis shows:
- Build agents at 85% capacity during peak hours
- $2,500/month build costs, 30% could be saved with optimization
- Test parallelization could reduce build time by 40%

**"Are we getting better or worse?"**
Trend analysis reveals:
- Build times increasing 5% month-over-month
- Test coverage improving from 70% to 85%
- Cache effectiveness declining, needs attention

## 📊 Example Output

### Single Build Analysis
```
🏗️  CHALK BUILD PERFORMANCE ANALYSIS
====================================

📊 BUILD SUMMARY
  Build ID: ci-build-1234
  Status: ✅ success
  Duration: 324.5s ⚡ (MODERATE)
  Platform: github
  Branch: feature/optimization

⚡ PERFORMANCE METRICS
  Build Time: 324.5s
  Queue Time: 12.3s
  Stage Breakdown:
    test      : █████████     48.3%
    compile   : ████████      44.7%
    package   : █              6.9%

📈 EFFICIENCY ANALYSIS
  Pipeline Efficiency: 67.8/100
  Cache Effectiveness: 85.5%
  Resource Utilization: 72.5%

🎯 ENGINEERING VELOCITY
  Score: 85/100 🏆 EXCELLENT

💡 RECOMMENDATIONS
  🔴 [HIGH] Tests consume 48% of build time
     Action: Parallelize test execution
     Impact: Reduce test time by 40%
     Effort: MEDIUM
```

### Regression Detection
```
🔄 BUILD REGRESSION ANALYSIS
============================

📊 COMPARISON SUMMARY
  Builds: #1234 → #1235
  Time Gap: 4 hours
  Same Branch: ✅

⚡ PERFORMANCE CHANGES
  Duration: 324.5s → 412.8s
  Change: 📈 +27.2% (+88.3s)
  
  Stage Impact:
    test: 156.8s → 238.5s 🔺 +52.1%

📋 ASSESSMENT
  Severity: 🔴 HIGH
  ⚠️ Significant build time increase detected
```

## 🛠️ Installation

1. **Download**: Get `chalk-build-analyser.skill`
2. **Upload to Claude**: Use the skill upload feature
3. **Ready to Use**: The skill activates automatically for Chalk data

## 📖 How to Use

### Basic Commands

**Analyze a single build:**
```python
report = load_chalk_report("build.json")
analysis = analyze_build_performance(report)
```

**Compare two builds:**
```python
regression = detect_regression(before_report, after_report)
```

**Analyze trends:**
```python
trends = analyze_trends(multiple_reports)
```

### What You Get

The skill provides:
- **Executive summaries** for quick understanding
- **Detailed metrics** for deep analysis
- **Actionable recommendations** with expected impact
- **Trend visualizations** for pattern recognition
- **Cost analysis** for budget planning

## 📈 Metrics Explained

### Performance Metrics
- **Build Duration**: Total time from start to finish
- **Queue Time**: Time waiting for resources
- **Stage Breakdown**: Time per build phase
- **Cache Hit Rate**: Effectiveness of caching

### DORA Metrics
- **Deployment Frequency**: How often you ship
- **Lead Time**: Commit to production time
- **MTTR**: Recovery time from failures
- **Change Failure Rate**: Deployment reliability

### Efficiency Scores
- **Pipeline Efficiency**: Overall optimization level (0-100)
- **Parallelization**: How well you use parallel execution
- **Resource Utilization**: Infrastructure efficiency

## 🎯 Optimization Recommendations

The skill automatically generates recommendations like:

### High Priority
- **Slow Tests**: "Parallelize test execution to save 2 minutes"
- **Low Cache Rate**: "Fix dependency caching to reduce build by 30%"
- **Long Queue**: "Add 2 more build agents to eliminate wait time"

### Medium Priority
- **Resource Waste**: "Downsize agents to save $500/month"
- **Test Flakiness**: "Fix flaky tests affecting 5% of builds"

### Low Priority
- **Minor Optimizations**: "Combine build steps to save 10 seconds"

## 🔄 Integration Examples

### CI/CD Pipeline
```yaml
# GitHub Actions
- name: Analyze Build
  run: |
    chalk extract > report.json
    python analyze_build.py report.json
    
- name: Check for Regression
  run: |
    python compare_builds.py baseline.json current.json
    if [ $? -ne 0 ]; then
      echo "Performance regression detected!"
      exit 1
    fi
```

### Monitoring Dashboard
Export metrics for visualization:
- Build duration trends
- Success rate over time
- Resource utilization
- Cost per build

## 🌟 Benefits

### Immediate Value
- **Find bottlenecks** in minutes, not hours
- **Detect regressions** before they reach production
- **Optimize costs** with data-driven decisions

### Long-term Impact
- **Improve velocity** with trend analysis
- **Reduce waste** through efficiency tracking
- **Scale smartly** with capacity planning

## 🔧 Automation Scripts

The Chalk Build Analyser includes command-line automation scripts for CI/CD integration and rapid analysis:

### chalk-build-analyser.sh

Basic Chalk build report analysis without AI enhancement.

**Features:**
- Analyze Chalk build reports (JSON)
- Compare two builds for regression detection
- Stage-by-stage breakdown
- Cache effectiveness metrics
- Resource utilization analysis
- Performance categorization (GOOD/MODERATE/HIGH)
- JSON export capability

**Usage:**
```bash
# Analyze a single build
./chalk-build-analyser.sh build-report.json

# Compare two builds
./chalk-build-analyser.sh --compare baseline.json current.json

# Export to JSON
./chalk-build-analyser.sh --format json --output analysis.json build.json
```

**Output Includes:**
- Build summary (ID, status, duration, platform, branch)
- Stage breakdown (if available)
- Cache hit rate
- Resource utilization
- Basic performance assessment
- Regression detection (in compare mode)

**Requirements:**
- jq: `brew install jq`

### chalk-build-analyser-claude.sh

AI-enhanced Chalk build analysis with Claude integration for intelligent performance insights.

**Features:**
- All features from basic analyser
- Executive summaries with build health scores
- Performance bottleneck identification
- Engineering velocity metrics (DORA)
- Actionable recommendations with effort/impact
- Cost & efficiency analysis
- Root cause analysis (in compare mode)

**Setup:**
```bash
# Option 1: Use .env file (recommended)
# Copy .env.example to .env and add your API key
cp ../../.env.example ../../.env
# Edit .env and set ANTHROPIC_API_KEY=sk-ant-xxx

# Option 2: Export environment variable
export ANTHROPIC_API_KEY=sk-ant-xxx
```

**Usage:**
```bash
# Analyze with AI insights (uses .env file or environment variable)
./chalk-build-analyser-claude.sh build-report.json

# Compare with regression analysis
./chalk-build-analyser-claude.sh --compare baseline.json current.json

# Or specify API key directly (overrides .env)
./chalk-build-analyser-claude.sh --api-key sk-ant-xxx build.json
```

**Single Build Output Includes:**
1. **Executive Summary** - Build health, key metrics, efficiency score
2. **Performance Analysis** - Bottlenecks, queue time, resource issues
3. **Engineering Velocity** - DORA metrics, team productivity
4. **Bottleneck Identification** - Specific slow stages with solutions
5. **Actionable Recommendations** - Prioritized by impact and effort
6. **Cost & Efficiency** - Resource waste, potential savings, ROI

**Comparison Output Includes:**
1. **Comparison Summary** - Build IDs, time gap, environment check
2. **Regression Detection** - Duration changes, stage-by-stage analysis
3. **Root Cause Analysis** - What changed and why
4. **Impact Assessment** - Severity rating, team velocity impact
5. **Recommendations** - Should build be blocked? What to fix?
6. **Efficiency Comparison** - Cache, resources, parallelization trends

**Requirements:**
- Same as basic analyser
- Anthropic API key

### compare-analysers.sh

Comparison tool that runs both basic and Claude-enhanced analysers to demonstrate value-add.

**Features:**
- Runs both analysers (single or compare mode)
- Compares outputs and capabilities
- Shows AI value-add with specific examples
- Generates comprehensive comparison report
- Optional output file preservation

**Usage:**
```bash
# Compare basic vs Claude analysis
./compare-analysers.sh build-report.json

# Compare regression detection capabilities
./compare-analysers.sh --compare baseline.json current.json

# Keep output files for review
./compare-analysers.sh --keep-outputs build.json
```

**Output:**
- Side-by-side capability comparison
- Value-add summary for build analysis
- Engineering velocity insights
- Use case recommendations
- Detailed output files (if --keep-outputs)

### CI/CD Integration

**GitHub Actions Example:**
```yaml
- name: Extract Chalk Report
  run: chalk extract > build-report.json

- name: Basic Build Analysis
  run: ./chalk-build-analyser.sh build-report.json

- name: Regression Check
  run: |
    ./chalk-build-analyser.sh --compare baseline.json build-report.json
    if grep -q "REGRESSION DETECTED" output.txt; then
      echo "Build performance regression detected!"
      exit 1
    fi

- name: AI Analysis (on main)
  if: github.ref == 'refs/heads/main'
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: |
    ./chalk-build-analyser-claude.sh build-report.json > analysis-report.txt
```

**GitLab CI Example:**
```yaml
build_analysis:
  script:
    - chalk extract > build-report.json
    - ./chalk-build-analyser.sh build-report.json
  artifacts:
    reports:
      metrics: build-report.json
```

## 📚 What's Included

- `SKILL.md` - Complete documentation
- `scripts/analyze_build.py` - Single build analyser (Python)
- `scripts/compare_builds.py` - Comparison tool (Python)
- `scripts/demo.sh` - Interactive demonstration
- `references/build_metrics.md` - Metrics guide
- `assets/` - Example reports
- `chalk-build-analyser.sh` - Basic CLI analyser
- `chalk-build-analyser-claude.sh` - AI-enhanced CLI analyser
- `compare-analysers.sh` - Analyser comparison tool

## 🤝 Perfect For

- **Development Teams** wanting faster feedback loops
- **DevOps Engineers** optimizing CI/CD pipelines
- **Engineering Managers** tracking team velocity
- **Platform Teams** managing build infrastructure
- **CTOs/VPEs** monitoring engineering efficiency

## 🚫 Not Included

This skill focuses on **build performance**, not security:
- No vulnerability scanning
- No SBOM analysis
- No supply chain security
- No compliance checking

For security features, use dedicated security tools.

## 📞 Support

- **Documentation**: See `SKILL.md` for detailed reference
- **Examples**: Run `demo.sh` for interactive examples
- **Metrics Guide**: Check `references/build_metrics.md`

## 🎉 Get Started

1. Upload the skill to Claude
2. Load your Chalk build reports
3. Get instant insights to improve your builds

Start optimizing your build pipeline today! Transform build data into engineering excellence.

---

*Built for engineers who value speed, efficiency, and continuous improvement.*
