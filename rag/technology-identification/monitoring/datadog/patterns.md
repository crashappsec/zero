# Datadog

**Category**: monitoring
**Description**: Datadog monitoring, APM, and observability platform
**Homepage**: https://www.datadoghq.com

## Package Detection

### NPM
*Datadog Node.js APM and browser SDK*

- `dd-trace`
- `datadog-metrics`
- `@datadog/browser-rum`
- `@datadog/browser-logs`

### PYPI
*Datadog Python APM and API client*

- `ddtrace`
- `datadog`
- `datadog-api-client`

### RUBYGEMS
*Datadog Ruby APM and StatsD client*

- `ddtrace`
- `dogapi`
- `dogstatsd-ruby`

### MAVEN
*Datadog Java APM*

- `com.datadoghq:dd-trace-api`
- `com.datadoghq:dd-java-agent`

### GO
*Datadog Go APM and StatsD client*

- `gopkg.in/DataDog/dd-trace-go.v1`
- `github.com/DataDog/datadog-go`

### Related Packages
- `@datadog/datadog-ci`
- `serverless-plugin-datadog`

## Import Detection

### Javascript

**Pattern**: `from\s+['"]dd-trace['"]`
- Type: esm_import

**Pattern**: `require\(['"]dd-trace['"]\)`
- Type: commonjs_require

**Pattern**: `from\s+['"]@datadog/browser-rum['"]`
- Type: esm_import

**Pattern**: `from\s+['"]@datadog/browser-logs['"]`
- Type: esm_import

### Python

**Pattern**: `from\s+ddtrace`
- Type: python_import

**Pattern**: `import\s+ddtrace`
- Type: python_import

**Pattern**: `from\s+datadog`
- Type: python_import

### Go

**Pattern**: `"gopkg\.in/DataDog/dd-trace-go`
- Type: go_import

**Pattern**: `"github\.com/DataDog/datadog-go`
- Type: go_import

## Environment Variables

*Datadog API key*

*Datadog application key*

*Datadog agent host*

*Service name for APM*

*Environment tag*

*Version tag*

*Enable APM tracing*

*Enable log correlation*

*Datadog site (datadoghq.com, datadoghq.eu)*


## Detection Notes

- Check for DD_API_KEY environment variable
- Look for datadog-agent in Docker configs
- Common with APM tracing and custom metrics

## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
- **API Endpoint Detection**: 80% (MEDIUM)
