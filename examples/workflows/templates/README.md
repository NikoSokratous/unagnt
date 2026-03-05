# Workflow Templates

This directory contains example workflow templates for the Unagnt marketplace. These templates demonstrate various use cases and best practices for multi-agent workflows.

## Available Templates

### 1. 🔍 Code Review Workflow
**File**: `code-review.yaml`  
**Category**: Code Review  
**Description**: Automated code review with multiple specialized agents

**Features**:
- Git repository cloning
- Static code analysis
- Security vulnerability scanning
- Complexity analysis
- Comprehensive report generation
- PR comment posting

**Use Case**: Automate code reviews for pull requests with consistent quality checks.

---

### 2. 📊 Data Processing Pipeline
**File**: `data-pipeline.yaml`  
**Category**: Data Pipeline  
**Description**: Complete ETL workflow with validation and quality checks

**Features**:
- Multi-source data extraction (API, database, S3, files)
- Schema validation
- Data quality scoring
- Transformation with configurable rules
- Error handling and retry logic
- Pipeline metrics and reporting

**Use Case**: Build reliable data pipelines with built-in quality gates.

---

### 3. 🔬 Research Assistant
**File**: `research.yaml`  
**Category**: Research  
**Description**: Multi-source research with analysis and synthesis

**Features**:
- Parallel searching across multiple sources (arXiv, PubMed, web)
- Source aggregation and deduplication
- Relevance ranking
- Paper analysis and insight extraction
- Synthesized findings with citations

**Use Case**: Conduct comprehensive research on any topic with AI assistance.

---

### 4. 💬 Customer Support Automation
**File**: `customer-support.yaml`  
**Category**: Customer Support  
**Description**: Intelligent ticket routing and automated responses

**Features**:
- Ticket classification and sentiment analysis
- Knowledge base search
- Automatic escalation for urgent issues
- AI-generated responses with quality review
- Multi-channel notifications

**Use Case**: Handle support tickets efficiently with AI-first approach.

---

### 5. ✍️ Content Creation Pipeline
**File**: `content-creation.yaml`  
**Category**: Content Creation  
**Description**: End-to-end content generation with SEO optimization

**Features**:
- Topic research and outline generation
- SEO keyword research and optimization
- AI writing with tone control
- Fact-checking and proofreading
- Visual asset generation
- Multi-platform publishing

**Use Case**: Create high-quality, SEO-optimized content at scale.

---

### 6. 🚀 DevOps Deployment Pipeline
**File**: `devops-deployment.yaml`  
**Category**: DevOps  
**Description**: Automated deployment with safety checks and rollback

**Features**:
- Comprehensive testing (unit, integration, smoke)
- Security scanning
- Canary deployments
- Health monitoring
- Automatic rollback on failure
- Documentation updates

**Use Case**: Deploy services safely with automated testing and monitoring.

---

## Using Templates

### Loading Templates into Marketplace

Templates are automatically loaded when the workflow engine starts:

```go
registry := workflow.NewTemplateRegistry()
if err := registry.LoadEmbeddedTemplates(); err != nil {
    log.Fatal(err)
}
```

### Customizing Templates

1. **Copy a template**:
   ```bash
   cp code-review.yaml my-custom-workflow.yaml
   ```

2. **Modify parameters**:
   - Add/remove parameters in the `parameters` section
   - Adjust default values
   - Update descriptions

3. **Customize workflow steps**:
   - Add new agents
   - Modify goals and conditions
   - Adjust timeouts and retry logic

4. **Test locally**:
   ```bash
   unagnt workflow run my-custom-workflow.yaml --param key=value
   ```

### Parameter Types

Templates support various parameter types:

- **string**: Text values
- **integer**: Numeric values
- **boolean**: true/false
- **list**: Array of values
- **object**: Complex nested data

### Template Variables

Use Go template syntax to reference parameters and outputs:

```yaml
goal: "Process data from {{.Outputs.source_location}}"
condition: "outputs.quality_score.score >= 0.7"
```

---

## Template Best Practices

### 1. **Clear Descriptions**
Write helpful descriptions for templates and parameters so users understand the purpose.

### 2. **Sensible Defaults**
Provide default values for optional parameters to make templates easier to use.

### 3. **Conditional Steps**
Use CEL conditions to skip steps when not needed:
```yaml
condition: "outputs.environment == 'production'"
```

### 4. **Error Handling**
Set appropriate `on_error` policies:
- `stop`: Fail fast for critical pipelines
- `continue`: Best-effort for reporting workflows

### 5. **Timeouts and Retries**
Configure realistic timeouts and retries for reliability:
```yaml
timeout: "5m"
retry: 3
```

### 6. **Output Keys**
Use meaningful output keys for step results:
```yaml
output_key: "validation_results"
```

### 7. **Agent Naming**
Use descriptive agent names that indicate their purpose:
- `git-agent`, `linter-agent`, `security-agent` (specific)
- Not: `agent1`, `agent2`, `helper` (vague)

---

## Template Structure

```yaml
# Metadata
name: "Template Name"
description: "Clear description of what this workflow does"
version: "1.0.0"
author: "Your Name"
category: "category-name"
tags: ["tag1", "tag2"]
icon: "🎯"

# User-configurable inputs
parameters:
  - name: param_name
    type: string
    description: "What this parameter controls"
    required: true
    default: "default-value"

# Workflow definition
workflow:
  name: "workflow-id"
  description: "Workflow purpose"
  
  # Sequential steps
  steps:
    - name: "step-name"
      agent: "agent-name"
      goal: "What the agent should do"
      output_key: "result_key"
      condition: "optional CEL expression"
      timeout: "5m"
      retry: 3
  
  # Parallel steps (optional)
  parallel:
    - name: "parallel-step"
      agent: "agent-name"
      goal: "Task to run in parallel"
  
  # Error handling
  on_error: "stop" # or "continue"
  timeout: "30m"
```

---

## Publishing Templates

To share your template with the community:

1. **Test thoroughly**: Ensure it works with various parameter combinations
2. **Document well**: Add clear descriptions and examples
3. **Publish to marketplace**:
   ```bash
   unagnt workflow publish my-template.yaml \
     --category category-name \
     --tags tag1,tag2,tag3
   ```

---

## Template Categories

- **code-review**: Code quality and review workflows
- **data-pipeline**: ETL and data processing
- **research**: Information gathering and analysis
- **customer-support**: Support automation
- **content-creation**: Content generation and publishing
- **devops**: Deployment and infrastructure
- **testing**: QA and testing workflows
- **monitoring**: System monitoring and alerting
- **security**: Security scanning and compliance
- **documentation**: Documentation generation

---

## Advanced Features

### Parallel Execution

Run multiple agents simultaneously:

```yaml
parallel:
  - name: "task-1"
    agent: "agent-1"
    goal: "First parallel task"
  - name: "task-2"
    agent: "agent-2"
    goal: "Second parallel task"
```

### CEL Conditions

Use Common Expression Language for dynamic logic:

```yaml
condition: "outputs.score > 0.8 && outputs.status == 'valid'"
```

### Template Functions

Access workflow context:

- `{{.Outputs.key}}`: Previous step outputs
- `{{.Outputs.step.field}}`: Nested output access
- Template conditionals and loops supported

---

## Contributing

To contribute new templates:

1. Create a new YAML file following the structure above
2. Test with real agents or mocks
3. Add documentation to this README
4. Submit a pull request

---

## Support

For questions or issues with templates:
- Open an issue on GitHub
- Join the community Discord
- Check the documentation at docs.Unagnt.io

---

**Happy workflow building! 🚀**
