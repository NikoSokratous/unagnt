# LangGraph Implementation

This is the **same code review agent** implemented with LangGraph, showing what you need to build yourself.

## What's Missing vs Unagnt

❌ No built-in policy enforcement  
❌ No cost tracking (must implement)  
❌ No approval gates (must implement)  
❌ No automatic retries (must implement)  
❌ No deterministic replay  
❌ No multi-tenancy  
❌ Observability requires LangSmith (paid)  

## Files

- `code_review_agent.py` - Main agent (200 lines)
- `state_management.py` - State handling (80 lines)
- `error_handling.py` - Retry logic (70 lines)
- `cost_tracking.py` - Manual cost tracking (50 lines)
- `observability.py` - Logging setup (50 lines)
- `requirements.txt` - Dependencies

**Total**: ~450 lines of Python + external dependencies

## Installation

```bash
pip install -r requirements.txt

# Set up environment
export OPENAI_API_KEY="sk-..."
export GITHUB_TOKEN="ghp_..."
export LANGSMITH_API_KEY="ls-..."  # For tracing (paid)
```

## Running

```bash
python code_review_agent.py \
  --pr-url https://github.com/user/repo/pull/123 \
  --severity-threshold medium \
  --auto-comment false
```

## What You Have to Implement Yourself

### 1. Policy Enforcement (Not Included)
```python
# You'd need to build:
class PolicyEngine:
    def check_cost_limit(self, cost): ...
    def check_dangerous_command(self, cmd): ...
    def check_credentials(self, text): ...
    def require_approval(self, action): ...
    # ~100 more lines
```

### 2. Cost Tracking
```python
# Manual implementation required
class CostTracker:
    def __init__(self):
        self.costs = {}
        self.budgets = {}
    
    def track_llm_call(self, model, tokens):
        # Calculate cost manually
        cost = self.calculate_cost(model, tokens)
        self.costs[model] = self.costs.get(model, 0) + cost
    
    def check_budget(self, tenant):
        # Manual budget checking
        ...
```

### 3. Retry Logic
```python
# Must implement retry with backoff
def retry_with_backoff(func, max_attempts=3):
    for attempt in range(max_attempts):
        try:
            return func()
        except RateLimitError:
            sleep(2 ** attempt)
        except NetworkError:
            sleep(2 ** attempt)
    raise MaxRetriesExceeded()
```

### 4. Human Approval
```python
# Manual approval system
class ApprovalGate:
    def request_approval(self, action, data):
        # Send to webhook
        response = requests.post(APPROVAL_WEBHOOK, json={
            "action": action,
            "data": data
        })
        
        # Poll for approval
        while not self.is_approved(response.id):
            time.sleep(30)
        
        return self.get_approval(response.id)
```

### 5. Observability
```python
# Requires LangSmith (paid) or manual implementation
from langsmith import Client

client = Client(api_key=os.getenv("LANGSMITH_API_KEY"))

@traceable
def my_agent_step(input_data):
    # Everything must be explicitly traced
    ...
```

### 6. Error Handling
```python
# Manual error handling throughout
try:
    result = agent.invoke(input_data)
except RateLimitError as e:
    # Handle rate limiting
    time.sleep(60)
    result = agent.invoke(input_data)
except APIError as e:
    # Handle API errors
    logger.error(f"API error: {e}")
    raise
except Exception as e:
    # Handle unknown errors
    logger.error(f"Unexpected error: {e}")
    raise
```

## Complexity Breakdown

| Component | Lines of Code | Effort |
|-----------|---------------|--------|
| Core agent logic | 200 | High |
| State management | 80 | Medium |
| Error handling | 70 | Medium |
| Cost tracking | 50 | Medium |
| Observability | 50 | High |
| Policy engine* | 100* | High |
| Approval gates* | 40* | Medium |
| Multi-tenancy* | 60* | High |
| **Total** | **650+** | **3-5 days** |

*Not included in this example but needed for production

## Key Differences

### Workflow Definition

**LangGraph**:
```python
# Imperative code - must define every transition
def build_graph():
    workflow = StateGraph(AgentState)
    
    workflow.add_node("fetch_pr", fetch_pr_node)
    workflow.add_node("static_analysis", static_analysis_node)
    workflow.add_node("code_review", code_review_node)
    workflow.add_node("check_breaking", check_breaking_node)
    workflow.add_node("generate_report", generate_report_node)
    workflow.add_node("post_comment", post_comment_node)
    
    workflow.set_entry_point("fetch_pr")
    workflow.add_edge("fetch_pr", "static_analysis")
    workflow.add_edge("static_analysis", "code_review")
    workflow.add_conditional_edges(
        "code_review",
        should_check_breaking,
        {
            True: "check_breaking",
            False: "generate_report"
        }
    )
    workflow.add_edge("check_breaking", "generate_report")
    workflow.add_edge("generate_report", "post_comment")
    workflow.set_finish_point("post_comment")
    
    return workflow.compile()
```

**Unagnt**:
```yaml
# Declarative YAML - just describe the flow
steps:
  - name: fetch_pr
    agent: "github_agent"
    
  - name: static_analysis
    agent: "linter_agent"
    
  - name: code_review
    agent: "code_reviewer"
    
  - name: check_breaking
    condition: |
      Outputs.pr_data.target_branch == "main"
    
  - name: generate_report
    agent: "report_generator"
    
  - name: post_comment
    agent: "github_commenter"
    require_approval: true
```

### Error Handling

**LangGraph**: Must wrap everything in try/catch
**Unagnt**: Built-in retry with backoff

### Cost Tracking

**LangGraph**: Manual calculation for every LLM call
**Unagnt**: Automatic with per-agent/tenant attribution

### Debugging

**LangGraph**: Add print statements, check logs
**Unagnt**: `unagnt replay <run-id>` - deterministic replay

## Production Readiness

To make this LangGraph implementation production-ready, you'd need to add:

1. ✅ Policy engine (~100 lines)
2. ✅ Comprehensive error handling (~50 lines)
3. ✅ Approval workflow system (~40 lines)
4. ✅ Structured logging (~30 lines)
5. ✅ Multi-tenancy support (~60 lines)
6. ✅ Budget tracking (~40 lines)
7. ✅ Alert system (~30 lines)
8. ✅ Deployment configuration (~50 lines)
9. ✅ Infrastructure as code (~100 lines)
10. ✅ Tests (~150 lines)

**Total additional effort**: ~650 lines + 2-3 days

## Deployment

```bash
# You need to handle deployment yourself
# Options:
# 1. Docker container
# 2. Kubernetes deployment
# 3. Serverless (AWS Lambda, etc.)
# 4. Custom infrastructure

# No built-in Kubernetes operator
# No built-in auto-scaling
# No built-in multi-region support
```

## Conclusion

LangGraph is powerful and flexible, but you're responsible for implementing all production features yourself. It's great for:
- Research and prototyping
- Custom logic requirements
- Learning graph-based agent architectures

For production systems with governance, cost control, and observability requirements, Unagnt provides these out-of-the-box.
